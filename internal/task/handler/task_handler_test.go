package handler

import (
	"context"
	"testing"
	"time"

	"github.com/my-backend-project/internal/pb"
	"github.com/my-backend-project/internal/pkg/errors"
	"github.com/my-backend-project/internal/task/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type mockTaskService struct {
	mock.Mock
}

func (m *mockTaskService) CreateTask(ctx context.Context, task *model.Task) (*model.Task, error) {
	args := m.Called(ctx, task)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Task), args.Error(1)
}

func (m *mockTaskService) GetTask(ctx context.Context, id string) (*model.Task, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Task), args.Error(1)
}

func (m *mockTaskService) ListTasks(ctx context.Context, userID string, status *model.TaskStatus, limit int32, offset string) ([]*model.Task, int32, error) {
	args := m.Called(ctx, userID, status, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.Task), args.Get(1).(int32), args.Error(2)
}

func (m *mockTaskService) UpdateTask(ctx context.Context, id string, task *model.Task) (*model.Task, error) {
	args := m.Called(ctx, id, task)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Task), args.Error(1)
}

func (m *mockTaskService) DeleteTask(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestTaskHandler_CreateTask(t *testing.T) {
	mockService := new(mockTaskService)
	handler := NewTaskHandler(mockService)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		dueDate := timestamppb.Now()
		req := &pb.CreateTaskRequest{
			UserId:      "user1",
			Title:       "Test Task",
			Description: "Test Description",
			Status:      pb.TaskStatus_TASK_STATUS_PENDING,
			DueDate:     dueDate,
		}

		expectedTask := &model.Task{
			ID:          primitive.NewObjectID(),
			UserID:      req.UserId,
			Title:       req.Title,
			Description: req.Description,
			Status:      model.TaskStatus(req.Status.String()),
			DueDate:     req.DueDate.AsTime(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mockService.On("CreateTask", ctx, mock.AnythingOfType("*model.Task")).Return(expectedTask, nil).Once()

		resp, err := handler.CreateTask(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, expectedTask.ID.Hex(), resp.TaskId)
		mockService.AssertExpectations(t)
	})

	t.Run("service_error", func(t *testing.T) {
		ctx := context.Background()
		dueDate := timestamppb.Now()
		req := &pb.CreateTaskRequest{
			UserId:      "user1",
			Title:       "Test Task",
			Description: "Test Description",
			Status:      pb.TaskStatus_TASK_STATUS_PENDING,
			DueDate:     dueDate,
		}

		mockService.On("CreateTask", ctx, mock.AnythingOfType("*model.Task")).Return(nil, errors.NewInternalError("service error", nil)).Once()

		resp, err := handler.CreateTask(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		mockService.AssertExpectations(t)
	})
}

func TestTaskHandler_GetTask(t *testing.T) {
	mockService := new(mockTaskService)
	handler := NewTaskHandler(mockService)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		taskID := primitive.NewObjectID()
		req := &pb.GetTaskRequest{
			TaskId: taskID.Hex(),
		}

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

		mockService.On("GetTask", ctx, taskID.Hex()).Return(expectedTask, nil).Once()

		resp, err := handler.GetTask(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, expectedTask.ID.Hex(), resp.Task.TaskId)
		mockService.AssertExpectations(t)
	})

	t.Run("not_found", func(t *testing.T) {
		ctx := context.Background()
		taskID := primitive.NewObjectID()
		req := &pb.GetTaskRequest{
			TaskId: taskID.Hex(),
		}

		mockService.On("GetTask", ctx, taskID.Hex()).Return(nil, errors.NewNotFoundError("タスクが見つかりません", nil)).Once()

		resp, err := handler.GetTask(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		mockService.AssertExpectations(t)
	})
}

func TestTaskHandler_ListTasks(t *testing.T) {
	mockService := new(mockTaskService)
	handler := NewTaskHandler(mockService)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		req := &pb.ListTasksRequest{
			UserId:    "user1",
			Status:    pb.TaskStatus_TASK_STATUS_PENDING,
			PageSize:  10,
			PageToken: "",
		}

		expectedTasks := []*model.Task{
			{
				ID:          primitive.NewObjectID(),
				UserID:      "user1",
				Title:       "Task 1",
				Description: "Description 1",
				Status:      model.TaskStatusPending,
				DueDate:     time.Now(),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			{
				ID:          primitive.NewObjectID(),
				UserID:      "user1",
				Title:       "Task 2",
				Description: "Description 2",
				Status:      model.TaskStatusPending,
				DueDate:     time.Now(),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		}

		mockService.On("ListTasks", ctx, "user1", mock.AnythingOfType("*model.TaskStatus"), int32(10), "").Return(expectedTasks, int32(2), nil).Once()

		resp, err := handler.ListTasks(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Tasks, 2)
		assert.Equal(t, int32(2), resp.TotalCount)
		mockService.AssertExpectations(t)
	})

	t.Run("service_error", func(t *testing.T) {
		ctx := context.Background()
		req := &pb.ListTasksRequest{
			UserId:    "user1",
			Status:    pb.TaskStatus_TASK_STATUS_PENDING,
			PageSize:  10,
			PageToken: "",
		}

		mockService.On("ListTasks", ctx, "user1", mock.AnythingOfType("*model.TaskStatus"), int32(10), "").Return(nil, int32(0), errors.NewInternalError("service error", nil)).Once()

		resp, err := handler.ListTasks(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		mockService.AssertExpectations(t)
	})
}

func TestTaskHandler_UpdateTask(t *testing.T) {
	mockService := new(mockTaskService)
	handler := NewTaskHandler(mockService)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		taskID := primitive.NewObjectID()
		dueDate := timestamppb.Now()
		req := &pb.UpdateTaskRequest{
			TaskId:      taskID.Hex(),
			UserId:      "user1",
			Title:       "Updated Task",
			Description: "Updated Description",
			Status:      pb.TaskStatus_TASK_STATUS_ACTIVE,
			DueDate:     dueDate,
		}

		expectedTask := &model.Task{
			ID:          taskID,
			UserID:      req.UserId,
			Title:       req.Title,
			Description: req.Description,
			Status:      model.TaskStatus(req.Status.String()),
			DueDate:     req.DueDate.AsTime(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mockService.On("UpdateTask", ctx, taskID.Hex(), mock.AnythingOfType("*model.Task")).Return(expectedTask, nil).Once()

		resp, err := handler.UpdateTask(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, expectedTask.ID.Hex(), resp.Task.TaskId)
		mockService.AssertExpectations(t)
	})

	t.Run("not_found", func(t *testing.T) {
		ctx := context.Background()
		taskID := primitive.NewObjectID()
		dueDate := timestamppb.Now()
		req := &pb.UpdateTaskRequest{
			TaskId:      taskID.Hex(),
			UserId:      "user1",
			Title:       "Updated Task",
			Description: "Updated Description",
			Status:      pb.TaskStatus_TASK_STATUS_ACTIVE,
			DueDate:     dueDate,
		}

		mockService.On("UpdateTask", ctx, taskID.Hex(), mock.AnythingOfType("*model.Task")).Return(nil, errors.NewNotFoundError("タスクが見つかりません", nil)).Once()

		resp, err := handler.UpdateTask(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		mockService.AssertExpectations(t)
	})
}

func TestTaskHandler_DeleteTask(t *testing.T) {
	mockService := new(mockTaskService)
	handler := NewTaskHandler(mockService)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		taskID := primitive.NewObjectID()
		req := &pb.DeleteTaskRequest{
			TaskId: taskID.Hex(),
		}

		mockService.On("DeleteTask", ctx, taskID.Hex()).Return(nil).Once()

		resp, err := handler.DeleteTask(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.IsType(t, &emptypb.Empty{}, resp)
		mockService.AssertExpectations(t)
	})

	t.Run("not_found", func(t *testing.T) {
		ctx := context.Background()
		taskID := primitive.NewObjectID()
		req := &pb.DeleteTaskRequest{
			TaskId: taskID.Hex(),
		}

		mockService.On("DeleteTask", ctx, taskID.Hex()).Return(errors.NewNotFoundError("タスクが見つかりません", nil)).Once()

		resp, err := handler.DeleteTask(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		mockService.AssertExpectations(t)
	})
}
