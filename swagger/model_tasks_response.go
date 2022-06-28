/*
 * Simple Task API
 *
 * This is a Task API
 *
 * API version: 1.0.0
 * Contact: you@your-company.com
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package swagger

type TasksResponse struct {

	TaskID string `json:"taskID"`

	TaskCreationDate int64 `json:"taskCreationDate"`

	FileID string `json:"fileID"`

	TaskStatus string `json:"taskStatus"`

	TaskResult []string `json:"taskResult"`
}