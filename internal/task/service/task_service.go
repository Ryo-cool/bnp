package service

import (
	"context"

	"my-backend-project/internal/task/model"
	"my-backend-project/internal/task/repository"
)

type TaskService interface {
	CreateTask(ctx context.Context, req *model.CreateTaskRequest) (*model.Task, error)
	GetTask(ctx context.Context, id string) (*model.Task, error)
	ListTasks(ctx context.Context, userID string, status *model.TaskStatus, limit int32, offset string) ([]*model.Task, int32, error)
	UpdateTask(ctx context.Context, req *model.UpdateTaskRequest) (*model.Task, error)
	DeleteTask(ctx context.Context, id string, userID string) (*model.Task, error)
}

type taskService struct {
	taskRepo repository.TaskRepository
}

func NewTaskService(taskRepo repository.TaskRepository) TaskService {
	return &taskService{
		taskRepo: taskRepo,
	}
}

func (s *taskService) CreateTask(ctx context.Context, req *model.CreateTaskRequest) (*model.Task, error) {
	task := &model.Task{
		UserID:      req.UserID,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		DueDate:     req.DueDate,
		CreatedAt:   req.DueDate,
		UpdatedAt:   req.DueDate,
	}

	return s.taskRepo.Create(ctx, task)
}

func (s *taskService) GetTask(ctx context.Context, id string) (*model.Task, error) {
	return s.taskRepo.FindByID(ctx, id)
}

func (s *taskService) ListTasks(ctx context.Context, userID string, status *model.TaskStatus, limit int32, offset string) ([]*model.Task, int32, error) {
	return s.taskRepo.FindByUserID(ctx, userID, status, limit, offset)
}

func (s *taskService) UpdateTask(ctx context.Context, req *model.UpdateTaskRequest) (*model.Task, error) {
	// 既存のタスクを取得
	task, err := s.taskRepo.FindByID(ctx, req.TaskID)
	if err != nil {
		return nil, err
	}

	// ユーザーIDの確認
	if task.UserID != req.UserID {
		return nil, repository.ErrTaskNotFound
	}

	// タスクの更新
	task.Title = req.Title
	task.Description = req.Description
	task.Status = req.Status
	task.DueDate = req.DueDate

	return s.taskRepo.Update(ctx, task)
}

func (s *taskService) DeleteTask(ctx context.Context, id string, userID string) (*model.Task, error) {
	// 既存のタスクを取得
	task, err := s.taskRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// ユーザーIDの確認
	if task.UserID != userID {
		return nil, repository.ErrTaskNotFound
	}

	return s.taskRepo.Delete(ctx, id)
}
