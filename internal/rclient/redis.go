package rclient

import (
	"errors"
	"os"

	"github.com/go-redis/redis/v8"
)

func GetRedisClient() (*redis.Client, error) {

	rdsUrl := os.Getenv("REDIS_SERVER_URL")
	rdsPWD := os.Getenv("REDIS_SERVER_PWD")

	if len(rdsUrl) <= 0 || len(rdsPWD) <= 0 {
		return nil, errors.New("Invalid Redis Creds Input, Please set REDIS_SERVER_URL and REDIS_SERVER_PWD in env variable")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     rdsUrl,
		Password: rdsPWD, // no password set
		DB:       0,      // use default DB
	})

	return rdb, nil
}
