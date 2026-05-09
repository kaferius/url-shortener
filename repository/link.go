package repository

import (
	"context"
	"database/sql"
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

	return links, nil
}

func (r *LinkRepository) Create(ctx context.Context, link Link) (Link, error) {
	err := r.db.QueryRow("INSERT INTO links (short, long) VALUES ($1, $2) RETURNING short, long", link.Short, link.Long).Scan(&link.Short, &link.Long)

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

	err = r.db.QueryRow("SELECT long FROM links WHERE short = $1", short).Scan(&long)
	if err != nil {
		return Link{Long: "", Short: short}, err
	}

	r.rdb.Set(ctx, key, long, time.Minute*30)

	return Link{Long: long, Short: short}, nil
}

func (r *LinkRepository) Delete(ctx context.Context, short string) error {
	key := "short:" + short

	r.rdb.Del(ctx, key)

	var long string
	err := r.db.QueryRow("SELECT long FROM links WHERE short = $1", short).Scan(&long)
	if err != nil {
		return err
	}
	_, err = r.db.Exec("DELETE FROM links WHERE short = $1", short)
	if err != nil {
		return err
	}

	return nil
}
