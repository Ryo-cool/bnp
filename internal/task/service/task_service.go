package service

import (
	"context"

	"my-backend-project/internal/pkg/errors"
	"my-backend-project/internal/task/model"
	"my-backend-project/internal/task/repository"
)

type TaskService interface {
	CreateTask(ctx context.Context, task *model.Task) (*model.Task, error)
	GetTask(ctx context.Context, id string) (*model.Task, error)
	ListTasks(ctx context.Context, userID string) ([]*model.Task, error)
	UpdateTask(ctx context.Context, task *model.Task) (*model.Task, error)
	DeleteTask(ctx context.Context, id string) (*model.Task, error)
}

type taskService struct {
	repo repository.TaskRepository
}

func NewTaskService(repo repository.TaskRepository) TaskService {
	return &taskService{
		repo: repo,
	}
}

func (s *taskService) CreateTask(ctx context.Context, task *model.Task) (*model.Task, error) {
	if task.Title == "" {
		return nil, errors.NewInvalidInputError("title is required", nil)
	}

	createdTask, err := s.repo.Create(ctx, task)
	if err != nil {
		return nil, errors.NewInternalError("failed to create task in repository", err)
	}

	return createdTask, nil
}

func (s *taskService) GetTask(ctx context.Context, id string) (*model.Task, error) {
	task, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err == repository.ErrTaskNotFound {
			return nil, errors.NewNotFoundError("task not found", err)
		}
		return nil, errors.NewInternalError("failed to get task from repository", err)
	}

	return task, nil
}

func (s *taskService) ListTasks(ctx context.Context, userID string) ([]*model.Task, error) {
	if userID == "" {
		return nil, errors.NewInvalidInputError("user_id is required", nil)
	}

	tasks, _, err := s.repo.FindByUserID(ctx, userID, nil, 0, "")
	if err != nil {
		return nil, errors.NewInternalError("failed to list tasks from repository", err)
	}

	return tasks, nil
}

func (s *taskService) UpdateTask(ctx context.Context, task *model.Task) (*model.Task, error) {
	if task.ID.IsZero() {
		return nil, errors.NewInvalidInputError("task id is required", nil)
	}

	if task.Title == "" {
		return nil, errors.NewInvalidInputError("title is required", nil)
	}

	existingTask, err := s.repo.FindByID(ctx, task.ID.Hex())
	if err != nil {
		if err == repository.ErrTaskNotFound {
			return nil, errors.NewNotFoundError("task not found", err)
		}
		return nil, errors.NewInternalError("failed to get task from repository", err)
	}

	// 既存のタスクの値を保持
	task.UserID = existingTask.UserID
	task.CreatedAt = existingTask.CreatedAt

	updatedTask, err := s.repo.Update(ctx, task)
	if err != nil {
		return nil, errors.NewInternalError("failed to update task in repository", err)
	}

	return updatedTask, nil
}

func (s *taskService) DeleteTask(ctx context.Context, id string) (*model.Task, error) {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err == repository.ErrTaskNotFound {
			return nil, errors.NewNotFoundError("task not found", err)
		}
		return nil, errors.NewInternalError("failed to get task from repository", err)
	}

	deletedTask, err := s.repo.Delete(ctx, id)
	if err != nil {
		return nil, errors.NewInternalError("failed to delete task from repository", err)
	}

	return deletedTask, nil
}
