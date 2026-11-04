//
// SPDX-License-Identifier: BSD-3-Clause
//

package redfish

import (
	"context"
	"encoding/json"
	"time"

	"github.com/coreweave/gofish/common"
)

type TaskOverWritePolicy string

const (
	// ManualTaskOverWritePolicy Completed tasks are not automatically overwritten.
	ManualTaskOverWritePolicy TaskOverWritePolicy = "Manual"
	// OldestTaskOverWritePolicy Oldest completed tasks are overwritten.
	OldestTaskOverWritePolicy TaskOverWritePolicy = "Oldest"
)

// TaskService is used to represent the task service offered by the redfish API
type TaskService struct {
	common.Resource

	// CompletedTaskOverWritePolicy how to handle the completed tasks.
	CompletedTaskOverWritePolicy TaskOverWritePolicy
	// DateTime system time.
	DateTime time.Time
	// LifeCycleEventOnTaskStateChange whether an event is reported when the task status is changed.
	LifeCycleEventOnTaskStateChange bool
	// ServiceEnabled indicates whether this service isenabled.
	ServiceEnabled bool
	// Status describes the status and health of a resource and its children.
	Status common.Status
	// TaskAutoDeleteTimeoutMinutes shall contain the number of minutes after which a completed task, where TaskState
	// contains the value 'Completed', 'Killed', 'Cancelled', or 'Exception', is deleted by the service.
	TaskAutoDeleteTimeoutMinutes int
	// TasksLink points towards the TasksLink store endpoint
	TasksLink common.Link `json:"Tasks"`
	// RawData holds the original serialized JSON so we can compare updates.
	RawData []byte
}

// UnmarshalJSON unmarshals a TaskService object from the raw JSON.
func (taskService *TaskService) UnmarshalJSON(b []byte) error {
	type temp TaskService
	var t struct {
		temp
	}

	err := json.Unmarshal(b, &t)
	if err != nil {
		return err
	}

	// Extract the links to other entities for later
	*taskService = TaskService(t.temp)

	// This is a read/write object, so we need to save the raw object data for later
	taskService.RawData = b

	return nil
}

// Tasks gets the collection of tasks of this task service
func (taskService *TaskService) Tasks(queryOpts ...common.QueryGroupOption) ([]*Task, error) {
	return taskService.TasksWithContext(common.ContextOf(taskService.GetClient()), queryOpts...)
}

// TasksWithContext gets the collection of tasks of this task service
func (taskService *TaskService) TasksWithContext(ctx context.Context, queryOpts ...common.QueryGroupOption) ([]*Task, error) {
	if taskService.TasksLink.IsZero() {
		return nil, nil
	}
	return ListReferencedTasksWithContext(ctx, taskService.GetClient(), taskService.TasksLink.String(), queryOpts...)
}

// Update commits updates to this object's properties to the running system.
func (taskService *TaskService) Update() error {
	return taskService.UpdateWithContext(common.ContextOf(taskService.GetClient()))
}

// UpdateWithContext commits updates to this object's properties to the running system.
func (taskService *TaskService) UpdateWithContext(ctx context.Context) error {
	readWriteFields := []string{
		"ServiceEnabled",
		"TaskAutoDeleteTimeoutMinutes",
	}

	return taskService.UpdateFromRawDataWithContext(ctx, taskService, taskService.RawData, readWriteFields)
}

// GetTaskService will get a TaskService instance from the service.
func GetTaskService(c common.Client, uri string, queryOpts ...common.QueryGroupOption) (*TaskService, error) {
	return GetTaskServiceWithContext(common.ContextOf(c), c, uri, queryOpts...)
}

// GetTaskServiceWithContext will get a TaskService instance from the service.
func GetTaskServiceWithContext(ctx context.Context, c common.Client, uri string, queryOpts ...common.QueryGroupOption) (*TaskService, error) {
	return common.GetObjectWithContext[TaskService](ctx, c, uri, queryOpts...)
}
