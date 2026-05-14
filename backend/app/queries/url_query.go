package queries

import (
	"context"
	"errors"

	"github.com/gottatouchsomegrass/url/app/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UrlQuery struct {
	DB *pgxpool.Pool
}

//insert url to db
func (q *UrlQuery) CreateUrl(ctx context.Context, url *models.URL) error {
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

//get by short url from db
func (q *UrlQuery) GetByShortUrl(ctx context.Context, code string) (*models.URL, error) {
	query := `
		SELECT id, short_url, long_url, expiry, clicks, created_at
		FROM urls
		WHERE short_url = $1
	`

	var url models.URL

	err := q.DB.QueryRow(ctx,query,code).Scan(
		&url.ID,
		&url.ShortURL,
		&url.LongURL,
		&url.Expiry,
		&url.Clicks,
		&url.CreatedAt,
	)

	if err!=nil {
		if errors.Is(err,pgx.ErrNoRows){
			return nil,nil
		}

		return nil,err
	}

	return &url,nil
}

//check whether customcode exists or not
func (q *UrlQuery) CustomCodeExists(ctx context.Context, code string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 FROM urls WHERE short_url = $1
		)
	`
	var isexist bool

	err := q.DB.QueryRow(ctx,query,code).Scan(&isexist)

	return isexist,err
}
