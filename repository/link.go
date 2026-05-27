package repository

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type Link struct {
	Long      string     `json:"long"`
	Short     string     `json:"short"`
	Clicks    int        `json:"clicks"`
	ExpiresAt *time.Time `json:"expires_at"`
	CreatedAt *time.Time `json:"created_at"`
}

type LinkRepository struct {
	db  *sql.DB
	rdb *redis.Client
}

func NewLinkRepository(db *sql.DB, rdb *redis.Client) *LinkRepository {
	return &LinkRepository{db: db, rdb: rdb}
}

func (r *LinkRepository) GetAll() ([]Link, error) {
	rows, err := r.db.Query("SELECT * FROM links WHERE expires_at > NOW()")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []Link

	for rows.Next() {
		var link Link
		if err := rows.Scan(&link.Short, &link.Long, &link.Clicks, &link.ExpiresAt, &link.CreatedAt); err != nil {
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
	err := r.db.QueryRowContext(
		ctx,
		"INSERT INTO links (short, long, clicks, expires_at) VALUES ($1, $2, 0, $3) RETURNING *",
		link.Short,
		link.Long,
		link.ExpiresAt,
	).Scan(
		&link.Short,
		&link.Long,
		&link.Clicks,
		&link.ExpiresAt,
		&link.CreatedAt,
	)

	if err == nil {
		var ttl time.Duration

		if link.ExpiresAt == nil || time.Until(*link.ExpiresAt) <= 0 {
			return Link{}, errors.New("invalid expiration time")
		}

		ttl = time.Until(*link.ExpiresAt)

		r.rdb.Set(ctx, "short:"+link.Short, link.Long, min(time.Minute*30, ttl))
	}

	return link, err
}

func (r *LinkRepository) Get(ctx context.Context, short string) (Link, error) {
	key := "short:" + short

	long, err := r.rdb.Get(ctx, key).Result()
	if err == nil {
		ttl, err := r.rdb.TTL(ctx, key).Result()
		if err != nil {
			return Link{Long: long, Short: short}, nil
		}
		expiresAt := time.Now().Add(ttl)
		return Link{Long: long, Short: short, ExpiresAt: &expiresAt}, nil
	}

	if err != redis.Nil {
		log.Printf("failed to get link from redis: %v\n", err)
	}

	var expiresAt time.Time
	err = r.db.QueryRowContext(ctx, "SELECT long, expires_at FROM links WHERE short = $1 AND expires_at > NOW()", short).Scan(&long, &expiresAt)
	if err != nil {
		return Link{Long: "", Short: short, ExpiresAt: &expiresAt}, err
	}

	r.rdb.Set(ctx, key, long, min(time.Minute*30, time.Until(expiresAt)))

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

func (r *LinkRepository) DeleteExpired(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM links WHERE expires_at <= NOW()")

	return err
}
