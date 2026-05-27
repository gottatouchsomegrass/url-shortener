package database

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

func RedisConnection() (*redis.Client,error) {
	rdb:=redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	_,err := rdb.Ping(context.Background()).Result()
	if err!=nil {
		return nil,err;
	}

	log.Println("redis connection successful...")

	return rdb,nil;
}
