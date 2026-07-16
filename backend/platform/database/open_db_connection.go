// Package database opens db conn
package database

import (
	"github.com/gottatouchsomegrass/url/app/queries"
)

type Queries struct {
	*queries.URLQuery
	*queries.AnalyticsQuery
	*queries.UserQuery
}

func OpenDBConnection() (*Queries, error) {
	//postgres conn
	db, err := PostgreSQLConnection()
	if err != nil {
		return nil, err
	}

	//redis conn
	rdb, err := RedisConnection()
	if err != nil {
		return nil, err
	}

	return &Queries{
		URLQuery: &queries.URLQuery{
			DB:  db,
			RDB: rdb,
		},
		AnalyticsQuery: &queries.AnalyticsQuery{
			DB: db,
		},
		UserQuery: &queries.UserQuery{
			DB: db,
		},
	}, nil
}
