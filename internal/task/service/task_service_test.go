package service

import (
	"context"
	"my-backend-project/internal/pkg/errors"
	"my-backend-project/internal/task/model"
	"my-backend-project/internal/task/repository"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// モックリポジトリの定義
type mockTaskRepository struct {
	mock.Mock
}

func (m *mockTaskRepository) Create(ctx context.Context, task *model.Task) (*model.Task, error) {
	args := m.Called(ctx, task)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Task), args.Error(1)
}

func (m *mockTaskRepository) FindByID(ctx context.Context, id string) (*model.Task, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Task), args.Error(1)
}

func (m *mockTaskRepository) FindByUserID(ctx context.Context, userID string, status *model.TaskStatus, limit int32, offset string) ([]*model.Task, int32, error) {
	args := m.Called(ctx, userID, status, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.Task), args.Get(1).(int32), args.Error(2)
}

func (m *mockTaskRepository) Update(ctx context.Context, task *model.Task) (*model.Task, error) {
	args := m.Called(ctx, task)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Task), args.Error(1)
}

func (m *mockTaskRepository) Delete(ctx context.Context, id string) (*model.Task, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Task), args.Error(1)
}

func TestTaskService_CreateTask(t *testing.T) {
	mockRepo := new(mockTaskRepository)
	service := NewTaskService(mockRepo)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		req := &model.CreateTaskRequest{
			UserID:      "user1",
			Title:       "Test Task",
			Description: "Test Description",
			Status:      model.TaskStatusPending,
			DueDate:     time.Now(),
		}

		expectedTask := &model.Task{
			ID:          primitive.NewObjectID(),
			UserID:      req.UserID,
			Title:       req.Title,
			Description: req.Description,
			Status:      req.Status,
			DueDate:     req.DueDate,
			CreatedAt:   req.DueDate,
			UpdatedAt:   req.DueDate,
		}

		mockRepo.On("Create", ctx, mock.AnythingOfType("*model.Task")).Return(expectedTask, nil).Once()

		result, err := service.CreateTask(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, expectedTask, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository_error", func(t *testing.T) {
		req := &model.CreateTaskRequest{
			UserID:      "user1",
			Title:       "Test Task",
			Description: "Test Description",
			Status:      model.TaskStatusPending,
			DueDate:     time.Now(),
		}

		mockRepo.On("Create", ctx, mock.AnythingOfType("*model.Task")).Return(nil, errors.NewInternalError("db error", nil)).Once()

		result, err := service.CreateTask(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.IsType(t, &errors.AppError{}, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestTaskService_GetTask(t *testing.T) {
	mockRepo := new(mockTaskRepository)
	service := NewTaskService(mockRepo)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		taskID := primitive.NewObjectID()
		expectedTask := &model.Task{
			ID:          taskID,
			UserID:      "user1",
			Title:       "Test Task",
			Description: "Test Description",
			Status:      model.TaskStatusPending,
			DueDate:     time.Now(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mockRepo.On("FindByID", ctx, taskID.Hex()).Return(expectedTask, nil).Once()

		result, err := service.GetTask(ctx, taskID.Hex())
		assert.NoError(t, err)
		assert.Equal(t, expectedTask, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not_found", func(t *testing.T) {
		taskID := primitive.NewObjectID()
		mockRepo.On("FindByID", ctx, taskID.Hex()).Return(nil, repository.ErrTaskNotFound).Once()

		result, err := service.GetTask(ctx, taskID.Hex())
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, repository.ErrTaskNotFound, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid_id", func(t *testing.T) {
		mockRepo.On("FindByID", ctx, "invalid-id").Return(nil, errors.NewInvalidInputError("invalid id", nil)).Once()

		result, err := service.GetTask(ctx, "invalid-id")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.IsType(t, &errors.AppError{}, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestTaskService_UpdateTask(t *testing.T) {
	mockRepo := new(mockTaskRepository)
	service := NewTaskService(mockRepo)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		taskID := primitive.NewObjectID()
		existingTask := &model.Task{
			ID:          taskID,
			UserID:      "user1",
			Title:       "Old Title",
			Description: "Old Description",
			Status:      model.TaskStatusPending,
			DueDate:     time.Now(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		req := &model.UpdateTaskRequest{
			TaskID:      taskID.Hex(),
			UserID:      "user1",
			Title:       "New Title",
			Description: "New Description",
			Status:      model.TaskStatusActive,
			DueDate:     time.Now(),
		}

		updatedTask := &model.Task{
			ID:          taskID,
			UserID:      req.UserID,
			Title:       req.Title,
			Description: req.Description,
			Status:      req.Status,
			DueDate:     req.DueDate,
			CreatedAt:   existingTask.CreatedAt,
			UpdatedAt:   time.Now(),
		}

		mockRepo.On("FindByID", ctx, taskID.Hex()).Return(existingTask, nil).Once()
		mockRepo.On("Update", ctx, mock.AnythingOfType("*model.Task")).Return(updatedTask, nil).Once()

		result, err := service.UpdateTask(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, updatedTask, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not_found", func(t *testing.T) {
		taskID := primitive.NewObjectID()
		req := &model.UpdateTaskRequest{
			TaskID:      taskID.Hex(),
			UserID:      "user1",
			Title:       "New Title",
			Description: "New Description",
			Status:      model.TaskStatusActive,
			DueDate:     time.Now(),
		}

		mockRepo.On("FindByID", ctx, taskID.Hex()).Return(nil, repository.ErrTaskNotFound).Once()

		result, err := service.UpdateTask(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, repository.ErrTaskNotFound, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("wrong_user", func(t *testing.T) {
		taskID := primitive.NewObjectID()
		existingTask := &model.Task{
			ID:          taskID,
			UserID:      "user1",
			Title:       "Old Title",
			Description: "Old Description",
			Status:      model.TaskStatusPending,
			DueDate:     time.Now(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		req := &model.UpdateTaskRequest{
			TaskID:      taskID.Hex(),
			UserID:      "user2", // 異なるユーザーID
			Title:       "New Title",
			Description: "New Description",
			Status:      model.TaskStatusActive,
			DueDate:     time.Now(),
		}

		mockRepo.On("FindByID", ctx, taskID.Hex()).Return(existingTask, nil).Once()

		result, err := service.UpdateTask(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, repository.ErrTaskNotFound, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestTaskService_DeleteTask(t *testing.T) {
	mockRepo := new(mockTaskRepository)
	service := NewTaskService(mockRepo)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		taskID := primitive.NewObjectID()
		existingTask := &model.Task{
			ID:          taskID,
			UserID:      "user1",
			Title:       "Test Task",
			Description: "Test Description",
			Status:      model.TaskStatusPending,
			DueDate:     time.Now(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mockRepo.On("FindByID", ctx, taskID.Hex()).Return(existingTask, nil).Once()
		mockRepo.On("Delete", ctx, taskID.Hex()).Return(existingTask, nil).Once()

		result, err := service.DeleteTask(ctx, taskID.Hex(), "user1")
		assert.NoError(t, err)
		assert.Equal(t, existingTask, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not_found", func(t *testing.T) {
		taskID := primitive.NewObjectID()
		mockRepo.On("FindByID", ctx, taskID.Hex()).Return(nil, repository.ErrTaskNotFound).Once()

		result, err := service.DeleteTask(ctx, taskID.Hex(), "user1")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, repository.ErrTaskNotFound, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("wrong_user", func(t *testing.T) {
		taskID := primitive.NewObjectID()
		existingTask := &model.Task{
			ID:          taskID,
			UserID:      "user1",
			Title:       "Test Task",
			Description: "Test Description",
			Status:      model.TaskStatusPending,
			DueDate:     time.Now(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mockRepo.On("FindByID", ctx, taskID.Hex()).Return(existingTask, nil).Once()

		result, err := service.DeleteTask(ctx, taskID.Hex(), "user2")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, repository.ErrTaskNotFound, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestTaskService_ListTasks(t *testing.T) {
	mockRepo := new(mockTaskRepository)
	service := NewTaskService(mockRepo)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		userID := "user1"
		status := model.TaskStatusPending
		limit := int32(10)
		offset := ""

		expectedTasks := []*model.Task{
			{
				ID:          primitive.NewObjectID(),
				UserID:      userID,
				Title:       "Task 1",
				Description: "Description 1",
				Status:      status,
				DueDate:     time.Now(),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			{
				ID:          primitive.NewObjectID(),
				UserID:      userID,
				Title:       "Task 2",
				Description: "Description 2",
				Status:      status,
				DueDate:     time.Now(),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		}
		expectedTotal := int32(2)

		mockRepo.On("FindByUserID", ctx, userID, &status, limit, offset).Return(expectedTasks, expectedTotal, nil).Once()

		tasks, total, err := service.ListTasks(ctx, userID, &status, limit, offset)
		assert.NoError(t, err)
		assert.Equal(t, expectedTasks, tasks)
		assert.Equal(t, expectedTotal, total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository_error", func(t *testing.T) {
		userID := "user1"
		status := model.TaskStatusPending
		limit := int32(10)
		offset := ""

		mockRepo.On("FindByUserID", ctx, userID, &status, limit, offset).Return(nil, int32(0), errors.NewInternalError("db error", nil)).Once()

		tasks, total, err := service.ListTasks(ctx, userID, &status, limit, offset)
		assert.Error(t, err)
		assert.Nil(t, tasks)
		assert.Equal(t, int32(0), total)
		assert.IsType(t, &errors.AppError{}, err)
		mockRepo.AssertExpectations(t)
	})
}
