package queries

// to be moved here in future

// // GetAnalyticsOverview returns the analytics overview for a given URL ID
// func (q *URLQuery) GetAnalyticsOverview(
// 	ctx context.Context,
// 	urlID int64,
// ) (*models.AnalyticsOverview, error) {
// 	query := `
// 		SELECT
// 			SUM(CASE WHEN created_at >= NOW() - INTERVAL '1 day' THEN 1 ELSE 0 END) AS today,
// 			SUM(CASE WHEN created_at >= NOW() - INTERVAL '7 days' THEN 1 ELSE 0 END) AS this_week,
// 			SUM(CASE WHEN created_at >= NOW() - INTERVAL '30 days' THEN 1 ELSE 0 END) AS this_month,
// 			COUNT(*) AS total_clicks
// 		FROM click_events
// 		WHERE url_id = $1
// 	`
// 	var overview models.AnalyticsOverview
// 	err := q.DB.QueryRow(ctx, query, urlID).Scan(
// 		&overview.Today,
// 		&overview.ThisWeek,
// 		&overview.ThisMonth,
// 		&overview.TotalClicks,
// 	)
// 	return &overview, err
// }

// // GetDailyClicks returns the daily click counts for a given URL ID
// func (q *URLQuery) GetDailyClicks(
// 	ctx context.Context,
// 	urlID int64,
// ) ([]models.DailyClick, error) {
// 	query := `
// 		SELECT DATE(created_at) AS date, COUNT(*) AS clicks
// 		FROM click_events
// 		WHERE url_id = $1
// 		GROUP BY DATE(created_at)
// 		ORDER BY DATE(created_at);
// 	`
// 	rows, err := q.DB.Query(ctx, query, urlID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	var dailyClicks []models.DailyClick
// 	for rows.Next() {
// 		var click models.DailyClick
// 		if err := rows.Scan(&click.Date, &click.Clicks); err != nil {
// 			return nil, err
// 		}
// 		dailyClicks = append(dailyClicks, click)
// 	}
// 	if err := rows.Err(); err != nil {
// 		return nil, err
// 	}
// 	return dailyClicks, nil
// }

// // GetRecentVisits returns the recent visits for a given URL ID
// func (q *URLQuery) GetRecentVisits(
// 	ctx context.Context,
// 	urlID int64,
// ) ([]models.RecentVisit, error) {
// 	query := `
// 		SELECT
// 			ip_address,
// 			referer,
// 			browser,
// 			device,
// 			created_at
// 		FROM click_events
// 		WHERE url_id = $1
// 		ORDER BY created_at DESC
// 		LIMIT 20
// 	`

// 	rows, err := q.DB.Query(ctx, query, urlID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	var visits []models.RecentVisit

// 	for rows.Next() {
// 		var visit models.RecentVisit

// 		err := rows.Scan(
// 			&visit.IPAddress,
// 			&visit.Referer,
// 			&visit.Browser,
// 			&visit.Device,
// 			&visit.CreatedAt,
// 		)

// 		if err != nil {
// 			return nil, err
// 		}

// 		visits = append(visits, visit)
// 	}

// 	if err := rows.Err(); err != nil {
// 		return nil, err
// 	}

// 	return visits, nil
// }
