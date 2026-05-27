package database

import (
	"github.com/gottatouchsomegrass/url/app/queries"
)

type Queries struct {
	*queries.UrlQuery
}

func OpenDBConnection() (*Queries, error) {
	//postgres conn
	db, err := PostgreSQLConnection()
	if err!=nil {
		return nil, err
	}

	//redis conn
	rdb, err := RedisConnection()
	if err!=nil {
		return nil, err
	}

	return &Queries{
		UrlQuery: &queries.UrlQuery{
			DB : db,
			RDB : rdb,
		},
	},nil
}
