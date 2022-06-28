package rclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/mytest/api/internal/constants"
)

// Task Interface to connect to a persistent system, This can be any DB or Cache based on the implementation
type TaskDBI interface {
	UpdateTaskStatus(thisTask *TaskDetails) error
	UpdateIPLookup(ip string, rlookup *ReverseLookup) error
	AddTaskToQueue(thisTask *TaskDetails) error
	AddToDeadLetterQ(thisTask *TaskDetails) error
	GetTaskDetails(taskID string) (*TaskDetails, bool, error)
	LookupIPDetails(key string) (*ReverseLookup, bool, error)
	GetTaskToProcess() (*TaskDetails, bool, error)
}

type TaskDetails struct {
	TaskID           string   `json:"taskID"`
	TaskCreationDate int64    `json:"taskCreationDate"`
	FileID           string   `json:"fileID"`
	TaskStatus       string   `json:"taskStatus"`
	TaskResult       []string `json:"taskResult"`
}

type ReverseLookup struct {
	IP        string   `json:"ip"`
	FileNames []string `json:"files"`
}

type RedisDBClient struct {
	rds *redis.Client
}

func (r *RedisDBClient) UpdateIPLookup(ip string, rlookup *ReverseLookup) error {
	newTaskDAta, _ := json.Marshal(rlookup)
	err := r.rds.Set(context.Background(), ip, newTaskDAta, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisDBClient) UpdateTaskStatus(thisTask *TaskDetails) error {
	newTaskDAta, _ := json.Marshal(thisTask)
	err := r.rds.Set(context.Background(), thisTask.TaskID, newTaskDAta, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisDBClient) AddTaskToQueue(thisTask *TaskDetails) error {
	newTaskDAta, _ := json.Marshal(thisTask)
	_, err := r.rds.LPush(context.Background(), constants.REDIS_IN_PROGRESS_QUEUE, newTaskDAta).Result()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisDBClient) AddToDeadLetterQ(thisTask *TaskDetails) error {
	newTaskDAta, _ := json.Marshal(thisTask)
	_, err := r.rds.LPush(context.Background(), constants.REDIS_DEAD_LETTER_QUEUE, newTaskDAta).Result()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisDBClient) GetTaskDetails(taskID string) (*TaskDetails, bool, error) {
	val, err := r.rds.Get(context.Background(), taskID).Result()
	if err == redis.Nil {
		return nil, false, errors.New("key with taskID not found")
	}

	var parsedMessage TaskDetails
	err = json.Unmarshal([]byte(val), &parsedMessage)
	return &parsedMessage, true, err
}

func (r *RedisDBClient) LookupIPDetails(key string) (*ReverseLookup, bool, error) {
	val, err := r.rds.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return nil, false, errors.New("key with taskID not found")
	}

	var parsedMessage ReverseLookup
	err = json.Unmarshal([]byte(val), &parsedMessage)
	return &parsedMessage, true, err
}

func (r *RedisDBClient) GetTaskToProcess() (*TaskDetails, bool, error) {
	myResult, err := r.rds.BLPop(context.Background(), 2*time.Second, constants.REDIS_IN_PROGRESS_QUEUE).Result()
	if err != nil {
		fmt.Println(err)
	}

	if len(myResult) > 0 {
		myTasks, err := parseTaskMessage([]byte(myResult[1]))
		return myTasks, true, err
	}

	return nil, false, err
}

func parseTaskMessage(message []byte) (*TaskDetails, error) {
	// json to struct
	var parsedMessage TaskDetails
	err := json.Unmarshal(message, &parsedMessage)
	if err != nil {
		return nil, err
	}

	return &parsedMessage, nil
}

func GetRedisClient() (TaskDBI, error) {

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

	return &RedisDBClient{rds: rdb}, nil
}
