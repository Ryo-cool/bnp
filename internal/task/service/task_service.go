package service

import (
	"context"

	"github.com/my-backend-project/internal/pb"
	"github.com/my-backend-project/internal/pkg/errors"
	"github.com/my-backend-project/internal/task/model"
	"github.com/my-backend-project/internal/task/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TaskService interface {
	CreateTask(ctx context.Context, task *model.Task) (*model.Task, error)
	GetTask(ctx context.Context, id string) (*model.Task, error)
	ListTasks(ctx context.Context, userID string, status *model.TaskStatus, limit int32, offset string) ([]*model.Task, int32, error)
	UpdateTask(ctx context.Context, id string, task *model.Task) (*model.Task, error)
	DeleteTask(ctx context.Context, id string) error
}

type taskService struct {
	taskRepo repository.TaskRepository
}

func NewTaskService(taskRepo repository.TaskRepository) TaskService {
	return &taskService{
		taskRepo: taskRepo,
	}
}

func (s *taskService) CreateTask(ctx context.Context, task *model.Task) (*model.Task, error) {
	createdTask, err := s.taskRepo.Create(ctx, task)
	if err != nil {
		return nil, errors.NewInternalError("タスクの作成に失敗しました", err)
	}
	return createdTask, nil
}

func (s *taskService) GetTask(ctx context.Context, id string) (*model.Task, error) {
	task, err := s.taskRepo.FindByID(ctx, id)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, errors.NewNotFoundError("タスクが見つかりません", err)
		}
		return nil, errors.NewInternalError("タスクの取得に失敗しました", err)
	}
	return task, nil
}

func (s *taskService) ListTasks(ctx context.Context, userID string, status *model.TaskStatus, limit int32, offset string) ([]*model.Task, int32, error) {
	tasks, total, err := s.taskRepo.FindByUserID(ctx, userID, status, limit, offset)
	if err != nil {
		return nil, 0, errors.NewInternalError("タスク一覧の取得に失敗しました", err)
	}
	return tasks, total, nil
}

func (s *taskService) UpdateTask(ctx context.Context, id string, task *model.Task) (*model.Task, error) {
	updatedTask, err := s.taskRepo.Update(ctx, id, task)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, errors.NewNotFoundError("タスクが見つかりません", err)
		}
		return nil, errors.NewInternalError("タスクの更新に失敗しました", err)
	}
	return updatedTask, nil
}

func (s *taskService) DeleteTask(ctx context.Context, id string) error {
	err := s.taskRepo.Delete(ctx, id)
	if err != nil {
		if errors.IsNotFound(err) {
			return errors.NewNotFoundError("タスクが見つかりません", err)
		}
		return errors.NewInternalError("タスクの削除に失敗しました", err)
	}
	return nil
}

// ModelToProto converts a Task model to a Task proto message
func ModelToProto(task *model.Task) *pb.Task {
	return &pb.Task{
		TaskId:      task.ID.Hex(),
		Title:       task.Title,
		Description: task.Description,
		Status:      pb.TaskStatus(pb.TaskStatus_value[string(task.Status)]),
		UserId:      task.UserID,
		CreatedAt:   model.TimeToProtoTimestamp(task.CreatedAt),
		UpdatedAt:   model.TimeToProtoTimestamp(task.UpdatedAt),
	}
}

// ProtoToModel converts a Task proto message to a Task model
func ProtoToModel(task *pb.Task) (*model.Task, error) {
	id, err := primitive.ObjectIDFromHex(task.TaskId)
	if err != nil {
		return nil, errors.NewInvalidInputError("無効なIDです", err)
	}

	return &model.Task{
		ID:          id,
		Title:       task.Title,
		Description: task.Description,
		Status:      model.TaskStatus(pb.TaskStatus_name[int32(task.Status)]),
		UserID:      task.UserId,
		CreatedAt:   model.ProtoTimestampToTime(task.CreatedAt),
		UpdatedAt:   model.ProtoTimestampToTime(task.UpdatedAt),
	}, nil
}
