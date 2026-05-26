package service

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

func StartFlusher(ctx context.Context, rdb *redis.Client, db *sql.DB) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			flush(ctx, rdb, db)
		}
	}
}

func flush(ctx context.Context, rdb *redis.Client, db *sql.DB) {
	shorts, err := rdb.SMembers(ctx, "click_keys").Result()

	if err != nil {
		log.Printf("failed to fetch click keys: %v\n", err)
		return
	}

	for _, short := range shorts {
		key := "clicks:" + short

		count, err := rdb.GetDel(ctx, key).Int64()
		if err != nil {
			continue
		}

		_, err = db.ExecContext(ctx, "UPDATE links SET clicks = clicks + $1 WHERE short = $2", count, short)
		if err != nil {
			continue
		}
	}
}
