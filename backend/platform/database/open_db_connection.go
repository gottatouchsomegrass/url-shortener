package database

import (
	"github.com/gottatouchsomegrass/url/app/queries"
)

type Queries struct {
	*queries.UrlQuery
}

func OpenDBConnection() (*Queries, error) {
	//defn a new conn
	db, err := PostgreSQLConnection()
	if err!=nil {
		return nil, err
	}

	return &Queries{
		UrlQuery: &queries.UrlQuery{DB : db},
	},nil
}
