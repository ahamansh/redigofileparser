package swagger

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mytest/api/internal/rclient"
	"github.com/mytest/api/swagger/model/v1/tasks"
)

// MVC
//	Models - Model Repo -> DBO -> DB / Cache

type HandlerFunctions struct {
	taskModel tasks.TaskRepository
	// SearchRepo
}

func InitializeHandlers() (handlerFunc HandlerFunctions) {
	handlerFunc = HandlerFunctions{}
	//Get a dynamoDB connection
	rds, err := rclient.GetRedisClient()
	if err != nil {
		panic(err)
	}

	// Init OAuth here if needed
	taskRepo := tasks.CreateRepository(rds)
	handlerFunc.taskModel = taskRepo
	return
}

func (handlerFunc HandlerFunctions) AddTask(w http.ResponseWriter, r *http.Request) {
	currRequest := CreateTasksRequest{}

	err := json.NewDecoder(r.Body).Decode(&currRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	taskID, err := handlerFunc.taskModel.CreateNewTask(currRequest.FileID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(taskID))
}

func (handlerFunc HandlerFunctions) SearchTasks(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	taskID := params["taskId"]
	currentTaskDetails, err := handlerFunc.taskModel.GetTaskDetails(taskID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
	}
	jsonResponse, err := json.Marshal(currentTaskDetails)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)

}

func (handlerFunc HandlerFunctions) SearchFilesByIP(w http.ResponseWriter, r *http.Request) {

	taskIP := r.URL.Query().Get("ip")
	currentTaskDetails, err := handlerFunc.taskModel.GetReverseLookup(taskIP)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
	}
	jsonResponse, err := json.Marshal(currentTaskDetails)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}
