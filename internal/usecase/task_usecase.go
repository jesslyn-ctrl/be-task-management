package usecase

import (
	"context"
	"net/http"
	"time"

	_const "bitbucket.org/edts/go-task-management/internal/constant"
	_model "bitbucket.org/edts/go-task-management/internal/model"
	_genModel "bitbucket.org/edts/go-task-management/internal/model/_generated"
	_pubsub "bitbucket.org/edts/go-task-management/internal/pubsub"
	_repo "bitbucket.org/edts/go-task-management/internal/repository"
	_customErr "bitbucket.org/edts/go-task-management/pkg/errors"
	_logger "bitbucket.org/edts/go-task-management/pkg/logger"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type TaskUsecaseInterface interface {
	CreateTask(ctx context.Context, input _genModel.CreateTaskInput) (*_model.Task, error)
	GetTasksByTeam(ctx context.Context, teamID string, status *string) ([]*_model.Task, error)
	GetTaskByID(ctx context.Context, taskID string) (*_model.Task, error)
	UpdateTaskById(ctx context.Context, input _genModel.UpdateTaskInput) (*_model.Task, error)
	DeleteTaskById(ctx context.Context, taskID string) error
	MoveTaskByID(ctx context.Context, input _genModel.MoveTaskInput) (*_model.Task, error)
	AssignTask(ctx context.Context, input _genModel.AssignTaskInput) (*_model.Task, error)

	// Subscription triggered event
	TaskCreatedEvent(ctx context.Context, teamID string) <-chan *_model.Task
	TaskUpdatedEvent(ctx context.Context, teamID string) <-chan *_model.Task
	TaskDeletedEvent(ctx context.Context, teamID string) <-chan *_genModel.DeletedTaskNotification
}

var logs = _logger.GetContextLoggerf(nil)

type TaskUsecase struct {
	// Repo
	taskRepo _repo.TaskRepositoryInterface
	userRepo _repo.UserRepositoryInterface
	teamRepo _repo.TeamRepositoryInterface
	// PubSub
	taskPubSub _pubsub.TaskPubSubInterface
}

func NewTaskUsecase(
	taskRepo _repo.TaskRepositoryInterface,
	userRepo _repo.UserRepositoryInterface,
	teamRepo _repo.TeamRepositoryInterface,
	taskPubSub _pubsub.TaskPubSubInterface) TaskUsecaseInterface {
	return &TaskUsecase{
		taskRepo:   taskRepo,
		userRepo:   userRepo,
		teamRepo:   teamRepo,
		taskPubSub: taskPubSub,
	}
}

func (uc *TaskUsecase) CreateTask(ctx context.Context, input _genModel.CreateTaskInput) (*_model.Task, error) {
	logs.Infof("CreateTask:: Starting with payload %v", input)
	// Convert string to time.Time
	parsedDueDate, err := time.Parse(time.RFC3339, input.DueDate)
	if err != nil {
		return nil, _customErr.NewGraphQLError(http.StatusBadRequest, "Invalid due date format")
	}

	task := &_model.Task{
		Title:       input.Title,
		Description: input.Description,
		Status:      input.Status,
		AssignedTo:  input.AssignedTo,
		TeamID:      input.TeamID,
		DueDate:     parsedDueDate,
	}

	// Save to repo
	createdTask, err := uc.taskRepo.CreateTask(ctx, task)
	if err != nil {
		logs.Errorf("CreateTask:: Error CreateTask repo: %v", err)
		return nil, err
	}

	if input.AssignedTo != nil {
		// Retrieve the user based on assigned to uuid
		assignedUser, err := uc.userRepo.GetUserByID(ctx, *input.AssignedTo)
		if err != nil {
			logs.Errorf("CreateTask:: Error GetUserByID repo: %v", err)
			return nil, _customErr.NewGraphQLError(http.StatusBadRequest, err.Error())
		}
		// Assign the user entity
		createdTask.AssignedUser = assignedUser
	}

	// Publish taskCreated event
	uc.taskPubSub.Publish(input.TeamID, "created", createdTask)

	logs.Info("CreateTask:: Finish CreateTask")

	return createdTask, nil
}

func (uc *TaskUsecase) TaskCreatedEvent(ctx context.Context, teamID string) <-chan *_model.Task {
	logs.Infof("TaskCreatedEvent:: Starting with variable teamId: %s", teamID)
	taskChan := make(chan *_model.Task, 1)

	go func() {
		eventChan := uc.taskPubSub.Subscribe(teamID)
		defer uc.taskPubSub.Unsubscribe(teamID, eventChan)

		for event := range eventChan {
			if event.Type == _const.CREATED {
				taskChan <- event.Task
			}
		}
		close(taskChan)
	}()

	logs.Info("TaskCreatedEvent:: Finish subscribing taskCreated")

	return taskChan
}

func (uc *TaskUsecase) GetTasksByTeam(ctx context.Context, teamID string, status *string) ([]*_model.Task, error) {
	logs.Infof("GetTasksByTeam:: Start fetching with variables teamId: %s and status: %v", teamID, status)

	tasks, err := uc.taskRepo.GetTasksByTeam(ctx, teamID, status)
	if err != nil {
		return nil, _customErr.NewGraphQLError(http.StatusInternalServerError, err.Error())
	}

	var results []*_model.Task
	for _, row := range tasks {
		task := &_model.Task{
			ID:          row.ID,
			Title:       row.Title,
			Description: row.Description,
			Status:      row.Status,
			DueDate:     row.DueDate,
			AssignedTo:  row.AssignedTo,
			TeamID:      row.TeamID,
			Base: _model.Base{
				CreatedAt:  row.CreatedAt,
				ModifiedAt: row.ModifiedAt,
			},
		}
		results = append(results, task)
	}

	logs.Info("GetTasksByTeam:: Finish fetching..")

	return results, nil
}

func (uc *TaskUsecase) GetTaskByID(ctx context.Context, taskID string) (*_model.Task, error) {
	task, err := uc.taskRepo.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, _customErr.NewGraphQLError(http.StatusBadRequest, "task not found")
	}
	return task, nil
}

// UpdateTaskById implements TaskUsecaseInterface.
func (uc *TaskUsecase) UpdateTaskById(ctx context.Context, input _genModel.UpdateTaskInput) (*_model.Task, error) {
	logs.Infof("UpdateTaskById:: Starting with payload %v", input)
	// Retrieve existing task from database
	existingTask, err := uc.taskRepo.GetTaskByID(ctx, input.ID)
	if err != nil {
		logs.Errorf("UpdateTaskById:: Error GetTaskByID repo %v", err)
		return nil, _customErr.NewGraphQLError(http.StatusNotFound, "Task Not Found")
	}

	// Update only the fields that are provided
	if input.Title != nil {
		existingTask.Title = *input.Title
	}
	if input.Description != nil {
		existingTask.Description = input.Description
	}

	if input.DueDate != nil {
		parsedDueDate, err := time.Parse(time.RFC3339, *input.DueDate)
		if err != nil {
			return nil, _customErr.NewGraphQLError(http.StatusBadRequest, "Invalid due date format")
		}
		existingTask.DueDate = parsedDueDate
	}

	// Save the updated task to the repository
	updatedTask, err := uc.taskRepo.UpdateTaskById(ctx, existingTask)
	if err != nil {
		return nil, _customErr.NewGraphQLError(http.StatusInternalServerError, err.Error())
	}

	// Publish taskCreated event
	uc.taskPubSub.Publish(existingTask.TeamID, "updated", updatedTask)

	logs.Info("UpdateTaskById:: Finish UpdateTaskById")

	return updatedTask, nil
}

// DeleteTaskById implements TaskUsecaseInterface.
func (uc *TaskUsecase) DeleteTaskById(ctx context.Context, taskID string) error {
	// Check if task exists
	existingTask, err := uc.taskRepo.GetTaskByID(ctx, taskID)
	if err != nil {
		return _customErr.NewGraphQLError(http.StatusBadRequest, "Task Not Found")
	}

	// Delete task from repository
	err = uc.taskRepo.DeleteTaskById(ctx, existingTask.ID)
	if err != nil {
		return _customErr.NewGraphQLError(http.StatusBadRequest, err.Error())
	}

	// Publish taskDeleted event
	uc.taskPubSub.Publish(existingTask.TeamID, "deleted", existingTask)

	return nil
}

// MoveTaskByID implements TaskUsecaseInterface.
func (uc *TaskUsecase) MoveTaskByID(ctx context.Context, input _genModel.MoveTaskInput) (*_model.Task, error) {
	updatedTask, err := uc.taskRepo.GetTaskByID(ctx, input.ID)
	if err != nil {
		return nil, _customErr.NewGraphQLError(http.StatusBadRequest, "Task Not Found")
	}

	//TODO: update task status by task id
	//1. Create new task entry
	//2. Call task repo to update the status
	//3. Return the updated task

	return updatedTask, err
}

// AssignTask implements TaskUsecaseInterface.
func (uc *TaskUsecase) AssignTask(ctx context.Context, input _genModel.AssignTaskInput) (*_model.Task, error) {
	_, err := uc.taskRepo.GetTaskByID(ctx, input.ID)
	if err != nil {
		return nil, _customErr.NewGraphQLError(http.StatusBadRequest, "Task Not Found")
	}

	// Retrieve the assigned user based on ID
	if input.AssignedTo != nil {
		_, err = uc.userRepo.GetUserByID(ctx, *input.AssignedTo)
		if err != nil {
			return nil, _customErr.NewGraphQLError(http.StatusBadRequest, "Assigned user not found")
		}
	}

	task := &_model.Task{
		ID:         input.ID,
		AssignedTo: input.AssignedTo,
	}

	updatedTask, err := uc.taskRepo.AssignTask(ctx, task)
	if err != nil {
		return nil, gqlerror.Errorf(err.Error())
	}

	return updatedTask, err

}

func (uc *TaskUsecase) TaskUpdatedEvent(ctx context.Context, teamID string) <-chan *_model.Task {
	taskChan := make(chan *_model.Task, 1)

	go func() {
		eventChan := uc.taskPubSub.Subscribe(teamID)
		defer uc.taskPubSub.Unsubscribe(teamID, eventChan)

		for event := range eventChan {
			if event.Type == _const.UPDATED {
				taskChan <- event.Task
			}
		}
		close(taskChan)
	}()

	return taskChan
}

func (uc *TaskUsecase) TaskDeletedEvent(ctx context.Context, teamID string) <-chan *_genModel.DeletedTaskNotification {
	taskChan := make(chan *_genModel.DeletedTaskNotification, 1) // Create a channel for DeletedTaskNotification

	go func() {
		eventChan := uc.taskPubSub.Subscribe(teamID)
		defer uc.taskPubSub.Unsubscribe(teamID, eventChan)

		for event := range eventChan {
			if event.Type == _const.DELETED {
				// Map the Task to DeletedTaskNotification (adjust according to your actual model)
				deletedTaskNotification := &_genModel.DeletedTaskNotification{
					TaskID:  event.Task.ID,
					Deleted: true,
				}

				taskChan <- deletedTaskNotification // Send the notification to the channel
			}
		}

		close(taskChan)
	}()

	return taskChan
}
