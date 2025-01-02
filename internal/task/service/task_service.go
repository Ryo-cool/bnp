package service

import (
	"context"

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
	return s.repo.Create(ctx, task)
}

func (s *taskService) GetTask(ctx context.Context, id string) (*model.Task, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *taskService) ListTasks(ctx context.Context, userID string) ([]*model.Task, error) {
	tasks, _, err := s.repo.FindByUserID(ctx, userID, nil, 0, "")
	return tasks, err
}

func (s *taskService) UpdateTask(ctx context.Context, task *model.Task) (*model.Task, error) {
	return s.repo.Update(ctx, task)
}

func (s *taskService) DeleteTask(ctx context.Context, id string) (*model.Task, error) {
	return s.repo.Delete(ctx, id)
}
