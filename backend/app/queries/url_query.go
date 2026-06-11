// Package queries have caching and db layer
package queries

import (
	"context"
	"errors"
	"fmt"
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
		INSERT INTO urls (
			user_id,
			short_url,
			long_url,
			expiry
		)
		VALUES ($1, $2, $3, $4)
	`

	_, err := q.DB.Exec(
		ctx,
		query,
		url.UserID,
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
		SELECT id, user_id, short_url, long_url, expiry, clicks, created_at
		FROM urls
		WHERE short_url = $1
	`

	var url models.URL

	err = q.DB.QueryRow(ctx, query, code).Scan(
		&url.ID,
		&url.UserID,
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

// IncrementClicks increments the click count for a given URL ID
func (q *URLQuery) IncrementClicks(ctx context.Context, id int64) error {
	query := `
		UPDATE urls
		SET clicks = clicks + 1
		WHERE id = $1
	`
	result, err := q.DB.Exec(ctx, query, id)
	if err != nil {
		fmt.Println("qery err in IncrementClicks:", err)
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("url not found")
	}
	return nil
}

// CreateClickEvent inserts a new click event into the database
func (q *URLQuery) CreateClickEvent(
	ctx context.Context,
	event *models.ClickEvent,
) error {
	query := `
		INSERT INTO click_events (url_id, ip_address, user_agent, referer, country, device, browser)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := q.DB.Exec(ctx, query,
		event.URLID,
		event.IPAddress,
		event.UserAgent,
		event.Referer,
		event.Country,
		event.Device,
		event.Browser,
	)
	return err
}

// Analytics Queries (here for now - will be moved to analytics_queries.go in future)

// GetAnalyticsOverview returns the analytics overview for a given URL ID
func (q *URLQuery) GetAnalyticsOverview(
	ctx context.Context,
	urlID int64,
) (*models.AnalyticsOverview, error) {
	query := `
		SELECT
			SUM(CASE WHEN created_at >= NOW() - INTERVAL '1 day' THEN 1 ELSE 0 END) AS today,
			SUM(CASE WHEN created_at >= NOW() - INTERVAL '7 days' THEN 1 ELSE 0 END) AS this_week,
			SUM(CASE WHEN created_at >= NOW() - INTERVAL '30 days' THEN 1 ELSE 0 END) AS this_month,
			COUNT(*) AS total_clicks
		FROM click_events
		WHERE url_id = $1
	`
	var overview models.AnalyticsOverview
	err := q.DB.QueryRow(ctx, query, urlID).Scan(
		&overview.Today,
		&overview.ThisWeek,
		&overview.ThisMonth,
		&overview.TotalClicks,
	)
	return &overview, err
}

// GetDailyClicks returns the daily click counts for a given URL ID
func (q *URLQuery) GetDailyClicks(
	ctx context.Context,
	urlID int64,
) ([]models.DailyClick, error) {
	query := `
		SELECT DATE(created_at) AS date, COUNT(*) AS clicks
		FROM click_events
		WHERE url_id = $1
		GROUP BY DATE(created_at)
		ORDER BY DATE(created_at);
	`
	rows, err := q.DB.Query(ctx, query, urlID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dailyClicks []models.DailyClick
	for rows.Next() {
		var click models.DailyClick
		if err := rows.Scan(&click.Date, &click.Clicks); err != nil {
			return nil, err
		}
		dailyClicks = append(dailyClicks, click)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return dailyClicks, nil
}

// GetRecentVisits returns the recent visits for a given URL ID
func (q *URLQuery) GetRecentVisits(
	ctx context.Context,
	urlID int64,
) ([]models.RecentVisit, error) {
	query := `
		SELECT
			ip_address,
			referer,
			browser,
			device,
			created_at
		FROM click_events
		WHERE url_id = $1
		ORDER BY created_at DESC
		LIMIT 20
	`

	rows, err := q.DB.Query(ctx, query, urlID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var visits []models.RecentVisit

	for rows.Next() {
		var visit models.RecentVisit

		err := rows.Scan(
			&visit.IPAddress,
			&visit.Referer,
			&visit.Browser,
			&visit.Device,
			&visit.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		visits = append(visits, visit)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return visits, nil
}

func (q *URLQuery) GetBrowserAnalytics(
	ctx context.Context,
	urlID int64,
) ([]models.BrowserAnalytics, error) {
	query := `
		SELECT browser, COUNT(*)
		FROM click_events
		WHERE url_id = $1
		GROUP BY browser
		ORDER BY COUNT(*) DESC;
	`

	var analytics []models.BrowserAnalytics

	rows, err := q.DB.Query(
		ctx,
		query,
		urlID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var browserAnalytics models.BrowserAnalytics
		err := rows.Scan(
			&browserAnalytics.Browser,
			&browserAnalytics.Count,
		)
		if err != nil {
			return nil, err
		}
		analytics = append(analytics, browserAnalytics)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return analytics, nil
}

func (q *URLQuery) GetDeviceAnalytics(
	ctx context.Context,
	urlID int64,
) ([]models.DeviceAnalytics, error) {
	query := `
		SELECT device, COUNT(*)
		FROM click_events
		WHERE url_id = $1
		GROUP BY device
		ORDER BY COUNT(*) DESC;
	`

	var analytics []models.DeviceAnalytics

	rows, err := q.DB.Query(
		ctx,
		query,
		urlID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var deviceAnalytics models.DeviceAnalytics
		err := rows.Scan(
			&deviceAnalytics.Device,
			&deviceAnalytics.Count,
		)
		if err != nil {
			return nil, err
		}
		analytics = append(analytics, deviceAnalytics)
	}

	if err != nil {
		return nil, err
	}

	return analytics, nil
}

// GetURLByID returns the URL for a given URL ID
func (q *URLQuery) GetURLByID(
	ctx context.Context,
	urlID int64,
) (*models.URL, error) {
	query := `
		SELECT
			id,
			user_id,
			long_url,
			short_url,
			expiry,
			clicks,
			created_at
		FROM urls
		WHERE id = $1
	`

	var url models.URL

	err := q.DB.QueryRow(
		ctx,
		query,
		urlID,
	).Scan(
		&url.ID,
		&url.UserID,
		&url.LongURL,
		&url.ShortURL,
		&url.Expiry,
		&url.Clicks,
		&url.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &url, nil
}
