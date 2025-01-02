package service

import (
	"context"
	"testing"
	"time"

	"github.com/my-backend-project/internal/pb"
	"github.com/my-backend-project/internal/pkg/apperrors"
	"github.com/my-backend-project/internal/task/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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

func (m *mockTaskRepository) Update(ctx context.Context, id string, task *model.Task) (*model.Task, error) {
	args := m.Called(ctx, id, task)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Task), args.Error(1)
}

func (m *mockTaskRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestTaskService_CreateTask(t *testing.T) {
	mockRepo := new(mockTaskRepository)
	service := NewTaskService(mockRepo)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		task := &model.Task{
			UserID:      "user1",
			Title:       "Test Task",
			Description: "Test Description",
			Status:      model.TaskStatusPending,
			DueDate:     time.Now(),
		}

		expectedTask := &model.Task{
			ID:          primitive.NewObjectID(),
			UserID:      task.UserID,
			Title:       task.Title,
			Description: task.Description,
			Status:      task.Status,
			DueDate:     task.DueDate,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mockRepo.On("Create", ctx, task).Return(expectedTask, nil).Once()

		createdTask, err := service.CreateTask(ctx, task)
		assert.NoError(t, err)
		assert.NotNil(t, createdTask)
		assert.Equal(t, expectedTask.ID, createdTask.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository_error", func(t *testing.T) {
		ctx := context.Background()
		task := &model.Task{
			UserID:      "user1",
			Title:       "Test Task",
			Description: "Test Description",
			Status:      model.TaskStatusPending,
			DueDate:     time.Now(),
		}

		mockRepo.On("Create", ctx, task).Return(nil, apperrors.NewInternalError("repository error", nil)).Once()

		createdTask, err := service.CreateTask(ctx, task)
		assert.Error(t, err)
		assert.Nil(t, createdTask)
		mockRepo.AssertExpectations(t)
	})
}

func TestTaskService_GetTask(t *testing.T) {
	mockRepo := new(mockTaskRepository)
	service := NewTaskService(mockRepo)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
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

		task, err := service.GetTask(ctx, taskID.Hex())
		assert.NoError(t, err)
		assert.NotNil(t, task)
		assert.Equal(t, expectedTask.ID, task.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not_found", func(t *testing.T) {
		ctx := context.Background()
		taskID := primitive.NewObjectID()

		mockRepo.On("FindByID", ctx, taskID.Hex()).Return(nil, apperrors.NewNotFoundError("タスクが見つかりません", nil)).Once()

		task, err := service.GetTask(ctx, taskID.Hex())
		assert.Error(t, err)
		assert.Nil(t, task)
		assert.True(t, apperrors.IsNotFound(err))
		mockRepo.AssertExpectations(t)
	})
}

func TestTaskService_ListTasks(t *testing.T) {
	mockRepo := new(mockTaskRepository)
	service := NewTaskService(mockRepo)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
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

		mockRepo.On("FindByUserID", ctx, userID, &status, limit, offset).Return(expectedTasks, int32(2), nil).Once()

		tasks, total, err := service.ListTasks(ctx, userID, &status, limit, offset)
		assert.NoError(t, err)
		assert.NotNil(t, tasks)
		assert.Equal(t, int32(2), total)
		assert.Len(t, tasks, 2)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository_error", func(t *testing.T) {
		ctx := context.Background()
		userID := "user1"
		status := model.TaskStatusPending
		limit := int32(10)
		offset := ""

		mockRepo.On("FindByUserID", ctx, userID, &status, limit, offset).Return(nil, int32(0), apperrors.NewInternalError("repository error", nil)).Once()

		tasks, total, err := service.ListTasks(ctx, userID, &status, limit, offset)
		assert.Error(t, err)
		assert.Nil(t, tasks)
		assert.Equal(t, int32(0), total)
		mockRepo.AssertExpectations(t)
	})
}

func TestTaskService_UpdateTask(t *testing.T) {
	mockRepo := new(mockTaskRepository)
	service := NewTaskService(mockRepo)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		taskID := primitive.NewObjectID()
		task := &model.Task{
			UserID:      "user1",
			Title:       "Updated Task",
			Description: "Updated Description",
			Status:      model.TaskStatusActive,
			DueDate:     time.Now(),
		}

		expectedTask := &model.Task{
			ID:          taskID,
			UserID:      task.UserID,
			Title:       task.Title,
			Description: task.Description,
			Status:      task.Status,
			DueDate:     task.DueDate,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mockRepo.On("Update", ctx, taskID.Hex(), task).Return(expectedTask, nil).Once()

		updatedTask, err := service.UpdateTask(ctx, taskID.Hex(), task)
		assert.NoError(t, err)
		assert.NotNil(t, updatedTask)
		assert.Equal(t, expectedTask.ID, updatedTask.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not_found", func(t *testing.T) {
		ctx := context.Background()
		taskID := primitive.NewObjectID()
		task := &model.Task{
			UserID:      "user1",
			Title:       "Updated Task",
			Description: "Updated Description",
			Status:      model.TaskStatusActive,
			DueDate:     time.Now(),
		}

		mockRepo.On("Update", ctx, taskID.Hex(), task).Return(nil, apperrors.NewNotFoundError("タスクが見つかりません", nil)).Once()

		updatedTask, err := service.UpdateTask(ctx, taskID.Hex(), task)
		assert.Error(t, err)
		assert.Nil(t, updatedTask)
		assert.True(t, apperrors.IsNotFound(err))
		mockRepo.AssertExpectations(t)
	})
}

func TestTaskService_DeleteTask(t *testing.T) {
	mockRepo := new(mockTaskRepository)
	service := NewTaskService(mockRepo)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		taskID := primitive.NewObjectID()

		mockRepo.On("Delete", ctx, taskID.Hex()).Return(nil).Once()

		err := service.DeleteTask(ctx, taskID.Hex())
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not_found", func(t *testing.T) {
		ctx := context.Background()
		taskID := primitive.NewObjectID()

		mockRepo.On("Delete", ctx, taskID.Hex()).Return(apperrors.NewNotFoundError("タスクが見つかりません", nil)).Once()

		err := service.DeleteTask(ctx, taskID.Hex())
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
		mockRepo.AssertExpectations(t)
	})
}

func TestModelToProto(t *testing.T) {
	now := time.Now()
	task := &model.Task{
		ID:          primitive.NewObjectID(),
		UserID:      "user1",
		Title:       "Test Task",
		Description: "Test Description",
		Status:      model.TaskStatusPending,
		DueDate:     now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	protoTask := ModelToProto(task)
	assert.NotNil(t, protoTask)
	assert.Equal(t, task.ID.Hex(), protoTask.TaskId)
	assert.Equal(t, task.UserID, protoTask.UserId)
	assert.Equal(t, task.Title, protoTask.Title)
	assert.Equal(t, task.Description, protoTask.Description)
	assert.Equal(t, pb.TaskStatus_TASK_STATUS_PENDING, protoTask.Status)
	assert.Equal(t, now.Unix(), protoTask.CreatedAt.GetSeconds())
	assert.Equal(t, now.Unix(), protoTask.UpdatedAt.GetSeconds())
}

func TestProtoToModel(t *testing.T) {
	now := time.Now()
	protoTask := &pb.Task{
		TaskId:      primitive.NewObjectID().Hex(),
		UserId:      "user1",
		Title:       "Test Task",
		Description: "Test Description",
		Status:      pb.TaskStatus_TASK_STATUS_PENDING,
		CreatedAt:   model.TimeToProtoTimestamp(now),
		UpdatedAt:   model.TimeToProtoTimestamp(now),
	}

	task, err := ProtoToModel(protoTask)
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, protoTask.TaskId, task.ID.Hex())
	assert.Equal(t, protoTask.UserId, task.UserID)
	assert.Equal(t, protoTask.Title, task.Title)
	assert.Equal(t, protoTask.Description, task.Description)
	assert.Equal(t, model.TaskStatusPending, task.Status)
	assert.Equal(t, now.Unix(), task.CreatedAt.Unix())
	assert.Equal(t, now.Unix(), task.UpdatedAt.Unix())

	// Invalid ID test
	invalidProtoTask := &pb.Task{
		TaskId: "invalid-id",
	}
	task, err = ProtoToModel(invalidProtoTask)
	assert.Error(t, err)
	assert.Nil(t, task)
	assert.True(t, apperrors.IsInvalidInput(err))
}
