package tasks

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/mytest/api/internal/constants"
	"github.com/mytest/api/internal/rclient"
)

type TaskRepository interface {
	CreateNewTask(fileID string) (string, error)
	GetTaskDetails(taskID string) (*rclient.TaskDetails, error)
	GetReverseLookup(IP string) (*rclient.ReverseLookup, error)
}

type TaskRepo struct {
	rds         rclient.TaskDBI
	taskDetails rclient.TaskDetails
}

func CreateRepository(rds rclient.TaskDBI) TaskRepository {
	return &TaskRepo{rds: rds}
}

func (t *TaskRepo) CreateNewTask(fileID string) (taskID string, err error) {
	// check if this task already exists. This validation is valid only if the file is not updated all the time
	newTask := rclient.TaskDetails{
		TaskID:           uuid.New().String(),
		TaskCreationDate: time.Now().UnixNano(),
		FileID:           fileID,
		TaskStatus:       constants.TASK_STATUS_INPROGRESS,
	}

	t.rds.UpdateTaskStatus(&newTask)
	t.rds.AddTaskToQueue(&newTask)
	return newTask.TaskID, nil
}

func (t *TaskRepo) GetTaskDetails(taskID string) (*rclient.TaskDetails, error) {
	val, found, err := t.rds.GetTaskDetails(taskID)
	if !found {
		return nil, errors.New("key with taskID not found")
	}
	return val, err
}

func (t *TaskRepo) GetReverseLookup(IP string) (*rclient.ReverseLookup, error) {

	val, found, err := t.rds.LookupIPDetails(IP)
	if !found {
		return nil, errors.New("key with IP not found")
	}
	return val, err

}
