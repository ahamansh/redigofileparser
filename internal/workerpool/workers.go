package workerpool

import (
	"fmt"
	"time"

	"github.com/mytest/api/internal/constants"
	filehandler "github.com/mytest/api/internal/fileshandler"
	"github.com/mytest/api/internal/rclient"
)

// To be used when starting the server
type WorkerPoolInterface interface {
	InitWorkers()
	StartWorkers()
}

type RedisListWorkers struct {
	rds                  rclient.TaskDBI
	CurrentActiveWorkers int
	MaxWorkers           int
	isShutdown           bool
	fileParser           filehandler.FileHanderInterface
}

func GetRedisWorkerPool(rdb rclient.TaskDBI) WorkerPoolInterface {
	return &RedisListWorkers{
		rds:        rdb,
		MaxWorkers: 1,
		fileParser: filehandler.GetLocalFileHandler(),
	}
}

// Implementation can change we can add a ticker or a worker pool based on channel
func (r *RedisListWorkers) listenAndServeList() {
	for {
		if r.isShutdown {
			break
		}
		myTasks, found, err := r.rds.GetTaskToProcess()

		if found {

			// Get the Task details
			if err != nil {
				myTasks.TaskStatus = constants.TASK_STATUS_INVALID
				r.rds.UpdateTaskStatus(myTasks)
				r.rds.AddToDeadLetterQ(myTasks)
				fmt.Println("Error for procesing the message, may be file is wrong, put to dead letter q")
				continue
			}

			// Check if this file is already processed ?
			// The file can be updated all the time not checking

			ips, err := r.fileParser.ProcessFile(myTasks.FileID)
			if err != nil {
				myTasks.TaskStatus = constants.TASK_STATUS_INVALID
				r.rds.UpdateTaskStatus(myTasks)
				r.rds.AddToDeadLetterQ(myTasks)
				fmt.Println("Error for procesing the message, may be file is wrong, put to dead letter q")
				continue
			}

			myTasks.TaskStatus = constants.TASK_STATUS_COMPELTED
			myTasks.TaskResult = ips
			r.rds.UpdateTaskStatus(myTasks)
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

// Implementation can change and needs to be updated. We can use either channels(we can pass the task as a func and workers should be able to execute them), for waitgroup for worker pool implementation.
// Current implementation is just a single goroutine for the data.
func (r *RedisListWorkers) StartWorkers() {
	go r.listenAndServeList()
}

func (r *RedisListWorkers) AddToReverseIndexQ(ips []string, fileName string) {

	// Retry can be implemented
	for _, ip := range ips {

		val, found, _ := r.rds.LookupIPDetails(ip)
		if !found {
			thisData := rclient.ReverseLookup{
				IP:        ip,
				FileNames: []string{fileName},
			}
			r.rds.UpdateIPLookup(ip, &thisData)
			continue
		}

		val.FileNames = append(val.FileNames, fileName)
		r.rds.UpdateIPLookup(ip, val)
	}
}
