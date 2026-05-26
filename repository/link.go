package repository

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type Link struct {
	Long  string `json:"long"`
	Short string `json:"short"`
}

type LinkRepository struct {
	db  *sql.DB
	rdb *redis.Client
}

func NewLinkRepository(db *sql.DB, rdb *redis.Client) *LinkRepository {
	return &LinkRepository{db: db, rdb: rdb}
}

func (r *LinkRepository) GetAll() ([]Link, error) {
	rows, err := r.db.Query("SELECT short, long FROM links")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []Link

	for rows.Next() {
		var link Link
		if err := rows.Scan(&link.Short, &link.Long); err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return links, nil
}

func (r *LinkRepository) Create(ctx context.Context, link Link) (Link, error) {
	err := r.db.QueryRowContext(ctx, "INSERT INTO links (short, long, clicks) VALUES ($1, $2, 0) RETURNING short, long", link.Short, link.Long).Scan(&link.Short, &link.Long)

	if err == nil {
		r.rdb.Set(ctx, "short:"+link.Short, link.Long, time.Minute*30)
	}

	return link, err
}

func (r *LinkRepository) Get(ctx context.Context, short string) (Link, error) {
	key := "short:" + short

	long, err := r.rdb.Get(ctx, key).Result()
	if err == nil {
		return Link{Long: long, Short: short}, nil
	}

	if err != redis.Nil {
		log.Printf("failed to get link from redis: %v\n", err)
	}

	err = r.db.QueryRowContext(ctx, "SELECT long FROM links WHERE short = $1", short).Scan(&long)
	if err != nil {
		return Link{Long: "", Short: short}, err
	}

	r.rdb.Set(ctx, key, long, time.Minute*30)

	return Link{Long: long, Short: short}, nil
}

func (r *LinkRepository) Delete(ctx context.Context, short string) error {
	key := "short:" + short

	r.rdb.Del(ctx, key)

	_, err := r.db.ExecContext(ctx, "DELETE FROM links WHERE short = $1", short)
	if err != nil {
		return err
	}

	return nil
}

func (r *LinkRepository) Increase(ctx context.Context, short string) {
	err := r.rdb.Incr(ctx, "clicks:"+short).Err()
	if err != nil {
		log.Printf("failed to increase clicks: %v\n", err)
	}
	err = r.rdb.SAdd(ctx, "click_keys", short).Err()
	if err != nil {
		log.Printf("failed to increase clicks: %v\n", err)
	}
}

func (r *LinkRepository) GetClicks(ctx context.Context, short string) (int, error) {
	var res int
	err := r.db.QueryRowContext(ctx, "SELECT clicks FROM links WHERE short = $1", short).Scan(&res)

	if err != nil {
		return 0, err
	}

	delta, err := r.rdb.Get(ctx, "clicks:"+short).Int()
	if err != nil {
		log.Printf("failed to get clicks from redis: %v\n", err)
		return res, nil
	}

	return res + delta, nil
}
