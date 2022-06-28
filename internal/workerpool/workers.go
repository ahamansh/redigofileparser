package workerpool

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/mytest/api/internal/constants"
	filehandler "github.com/mytest/api/internal/fileshandler"
	"github.com/mytest/api/swagger/model/v1/tasks"
)

type WorkerPoolInterface interface {
	InitWorkers()
	StartWorkers()
}

type RedisListWorkers struct {
	rds                  *redis.Client
	CurrentActiveWorkers int
	MaxWorkers           int
	isShutdown           bool
	fileParser           filehandler.FileHanderInterface
}

func GetRedisWorkerPool(rds *redis.Client) WorkerPoolInterface {
	return &RedisListWorkers{
		rds:        rds,
		MaxWorkers: 1,
		fileParser: filehandler.GetLocalFileHandler(),
	}
}

func (r *RedisListWorkers) parseTaskMessage(message []byte) (*tasks.TaskDetails, error) {
	// json to struct
	var parsedMessage tasks.TaskDetails
	err := json.Unmarshal(message, &parsedMessage)
	if err != nil {
		return nil, err
	}

	return &parsedMessage, nil
}

// Implementation can change we can add a ticker or a worker pool based on channel
func (r *RedisListWorkers) listenAndServeList() {
	for {
		if r.isShutdown {
			break
		}
		myResult, err := r.rds.BLPop(context.Background(), 2*time.Second, constants.REDIS_IN_PROGRESS_QUEUE).Result()
		if err != nil {
			fmt.Println(err)
		}

		if len(myResult) > 0 {

			// Get the Task details
			myTasks, err := r.parseTaskMessage([]byte(myResult[1]))
			if err != nil {
				myTasks.TaskStatus = constants.TASK_STATUS_INVALID
				r.UpdateTaskStatus(myTasks)
				fmt.Println("Error for procesing the message, may be file is wrong, put to dead letter q")
				continue
			}

			// Check if this file is already processed ?
			// The file can be updated all the time not checking

			ips, err := r.fileParser.ProcessFile(myTasks.FileID)
			if err != nil {
				myTasks.TaskStatus = constants.TASK_STATUS_INVALID
				r.UpdateTaskStatus(myTasks)
				fmt.Println("Error for procesing the message, may be file is wrong, put to dead letter q")
				continue
			}

			myTasks.TaskStatus = constants.TASK_STATUS_COMPELTED
			myTasks.TaskResult = ips
			r.UpdateTaskStatus(myTasks)
			if err != nil {
				fmt.Println("Error for procesing the message, may be file is wrong, put to dead letter q")
				continue
			}
			r.AddToReverseIndexQ(myTasks.TaskResult, myTasks.FileID)

			fmt.Println(myTasks, err)
		}
		time.Sleep(2 * time.Second)
	}
}

func (r *RedisListWorkers) InitWorkers() {

}

func (r *RedisListWorkers) StartWorkers() {
	go r.listenAndServeList()
}

func (r *RedisListWorkers) AddToDeadLetterQ(thisTask *tasks.TaskDetails) {
	newTaskDAta, _ := json.Marshal(thisTask)
	_, err := r.rds.LPush(context.Background(), constants.REDIS_DEAD_LETTER_QUEUE, newTaskDAta).Result()
	if err != nil {
		return
	}
}

func (r *RedisListWorkers) UpdateTaskStatus(thisTask *tasks.TaskDetails) error {
	newTaskDAta, _ := json.Marshal(thisTask)
	err := r.rds.Set(context.Background(), thisTask.TaskID, newTaskDAta, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisListWorkers) AddToReverseIndexQ(ips []string, fileName string) {

	// Check for the retry if the entry is simultaneously updated

	for _, ip := range ips {

		val, err := r.rds.Get(context.Background(), ip).Result()
		if err != redis.Nil {
			thisData := tasks.ReverseLookup{
				IP:        ip,
				FileNames: []string{fileName},
			}
			newTaskDAta, _ := json.Marshal(thisData)

			r.rds.Set(context.Background(), ip, newTaskDAta, 0).Err()
		}

		var parsedMessage tasks.ReverseLookup
		err = json.Unmarshal([]byte(val), &parsedMessage)
		parsedMessage.FileNames = append(parsedMessage.FileNames, fileName)
		newTaskDAta, _ := json.Marshal(parsedMessage)
		r.rds.Set(context.Background(), ip, newTaskDAta, 0).Err()
	}

}

// file laoder
