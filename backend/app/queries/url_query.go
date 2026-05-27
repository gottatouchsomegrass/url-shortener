// Package queries have caching and db layer
package queries

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/gottatouchsomegrass/url/app/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type URLQuery struct {
	DB  *pgxpool.Pool
	RDB *redis.Client
}

// CreateURL insert url to db
func (q *URLQuery) CreateURL(ctx context.Context, url *models.URL) error {
	query := `
		INSERT INTO urls (short_url, long_url, expiry)
		VALUES ($1,$2,$3)
	`
	_, err := q.DB.Exec(ctx, query,
		url.ShortURL,
		url.LongURL,
		url.Expiry,
	)

	return err
}

// GetByShortURL get by short url from db
func (q *URLQuery) GetByShortURL(ctx context.Context, code string) (*models.URL, error) {
	//cache-aside implementation
	cached, err := q.RDB.Get(ctx, code).Result()
	if err == nil {
		return &models.URL{
			ShortURL: code,
			LongURL:  cached,
		}, nil
	}
	if err != redis.Nil {
		log.Println("redis err:", err)
	}

	query := `
		SELECT id, short_url, long_url, expiry, clicks, created_at
		FROM urls
		WHERE short_url = $1
	`

	var url models.URL

	err = q.DB.QueryRow(ctx, query, code).Scan(
		&url.ID,
		&url.ShortURL,
		&url.LongURL,
		&url.Expiry,
		&url.Clicks,
		&url.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	ttl := time.Hour

	if url.Expiry != nil {
		remaining := time.Until(*url.Expiry)
		if remaining <= 0 {
			return nil, errors.New("link expired")
		}
		if remaining <= ttl {
			ttl = remaining
		}
	}
	err = q.RDB.Set(
		ctx,
		code,
		url.LongURL,
		ttl,
	).Err()

	if err != nil {
		log.Println("redis set err:", err)
	}

	return &url, nil
}

// CustomCodeExists check whether customcode exists or not
func (q *URLQuery) CustomCodeExists(ctx context.Context, code string) (bool, error) {
	// _, err := q.RDB.Get(ctx,code).Result()
	// if err==nil {
	// 	return true, nil
	// }
	// if err!=redis.Nil {
	// 	return false, err
	// }
	query := `
		SELECT EXISTS (
			SELECT 1 FROM urls WHERE short_url = $1
		)
	`
	var isexist bool

	err := q.DB.QueryRow(ctx, query, code).Scan(&isexist)

	return isexist, err
}
