package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/mytest/api/internal/constants"
)

type TaskRepository interface {
	CreateNewTask(fileID string) (string, error)
	GetTaskDetails(taskID string) (*TaskDetails, error)
	GetReverseLookup(IP string) (*ReverseLookup, error)
}

type ReverseLookup struct {
	IP        string   `json:"ip"`
	FileNames []string `json:"files"`
}

type TaskDetails struct {
	TaskID           string   `json:"taskID"`
	TaskCreationDate int64    `json:"taskCreationDate"`
	FileID           string   `json:"fileID"`
	TaskStatus       string   `json:"taskStatus"`
	TaskResult       []string `json:"taskResult"`
}

type TaskRepo struct {
	rds         *redis.Client
	taskDetails TaskDetails
}

func CreateRepository(rds *redis.Client) TaskRepository {
	return &TaskRepo{rds: rds}
}

func (t *TaskRepo) CreateNewTask(fileID string) (taskID string, err error) {

	// check if this task already exists

	newTask := TaskDetails{
		TaskID:           uuid.New().String(),
		TaskCreationDate: time.Now().UnixNano(),
		FileID:           fileID,
		TaskStatus:       constants.TASK_STATUS_INPROGRESS,
	}

	newTaskDAta, _ := json.Marshal(newTask)
	err = t.rds.Set(context.Background(), newTask.TaskID, newTaskDAta, 0).Err()
	_, err = t.rds.LPush(context.Background(), constants.REDIS_IN_PROGRESS_QUEUE, newTaskDAta).Result()
	if err != nil {
		return "", errors.New("Unable to persist the task")
	}

	return newTask.TaskID, nil
}

func (t *TaskRepo) GetTaskDetails(taskID string) (*TaskDetails, error) {

	val, err := t.rds.Get(context.Background(), taskID).Result()
	if err == redis.Nil {
		return nil, errors.New("key with taskID not found")
	}
	//fmt.Println("For key taskID, Value is ", val)

	var parsedMessage TaskDetails
	err = json.Unmarshal([]byte(val), &parsedMessage)
	return &parsedMessage, err
}

func (t *TaskRepo) GetReverseLookup(IP string) (*ReverseLookup, error) {

	val, err := t.rds.Get(context.Background(), IP).Result()
	if err == redis.Nil {
		return nil, errors.New("key with taskID not found")
	}
	//fmt.Println("For key taskID, Value is ", val)

	var parsedMessage ReverseLookup
	err = json.Unmarshal([]byte(val), &parsedMessage)
	return &parsedMessage, err
}
