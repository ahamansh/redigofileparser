package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

type Employee struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

//redis://default:redispw@localhost:55003
//redis-cli -h 127.0.0.1 -p 55003 -a 'redispw'
func main() {
	fmt.Println("Hello there ")
	rdb, err := getRedisClient()
	if err != nil {
		panic(err)
	}

	err = rdb.Set(ctx, "key", "value", 0).Err()
	if err != nil {
		panic(err)
	}
}

func getRedisClient() (*redis.Client, error) {

	rdsUrl := os.Getenv("REDIS_SERVER_URL")
	rdsPWD := os.Getenv("REDIS_SERVER_PWD")

	if len(rdsUrl) <= 0 || len(rdsPWD) <= 0 {
		return nil, errors.New("Invalid Input")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     rdsUrl,
		Password: rdsPWD, // no password set
		DB:       0,      // use default DB
	})

	return rdb, nil
}

// Step -> Load all files in set files
// // FILD ID - FIle O

// POST
// // Set InProcess
// 	TaskID - fileID
// // Set Completed
// 	Task ID -> Task object

// // POST
// 	TaskDI - FIleID

// pubsub task.*
// // Processing
// 	append ip in file name for loop and pattern match

// For the Search
// // Reverse index
// // set with IP -> []fileNames

// // Task ID -> Task object
