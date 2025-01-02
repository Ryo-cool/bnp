package handler_test

import (
	"context"
	"testing"

	"my-backend-project/internal/task/handler"
	"my-backend-project/internal/task/model"
	"my-backend-project/internal/task/pb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MockTaskService struct {
	mock.Mock
}

func (m *MockTaskService) CreateTask(ctx context.Context, task *model.Task) (*model.Task, error) {
	args := m.Called(ctx, task)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Task), args.Error(1)
}

func (m *MockTaskService) GetTask(ctx context.Context, id string) (*model.Task, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Task), args.Error(1)
}

func (m *MockTaskService) ListTasks(ctx context.Context, userID string) ([]*model.Task, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*model.Task), args.Error(1)
}

func (m *MockTaskService) UpdateTask(ctx context.Context, task *model.Task) (*model.Task, error) {
	args := m.Called(ctx, task)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Task), args.Error(1)
}

func (m *MockTaskService) DeleteTask(ctx context.Context, id string) (*model.Task, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Task), args.Error(1)
}

func TestCreateTask(t *testing.T) {
	mockService := new(MockTaskService)
	taskHandler := handler.NewTaskHandler(mockService)

	ctx := context.WithValue(context.Background(), "user_id", "user123")
	req := &pb.CreateTaskRequest{
		Title:       "Test Task",
		Description: "Test Description",
		Status:      pb.TaskStatus(model.TaskStatusPending),
	}

	expectedTask := &model.Task{
		ID:          primitive.NewObjectID(),
		Title:       req.Title,
		Description: req.Description,
		UserID:      "user123",
		Status:      model.TaskStatusPending,
	}

	mockService.On("CreateTask", ctx, mock.AnythingOfType("*model.Task")).Return(expectedTask, nil)

	resp, err := taskHandler.CreateTask(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, expectedTask.ID.Hex(), resp.Id)
	assert.Equal(t, expectedTask.Title, resp.Title)
	mockService.AssertExpectations(t)
}

func TestCreateTaskWithoutUserID(t *testing.T) {
	mockService := new(MockTaskService)
	taskHandler := handler.NewTaskHandler(mockService)

	ctx := context.Background()
	req := &pb.CreateTaskRequest{
		Title:       "Test Task",
		Description: "Test Description",
		Status:      pb.TaskStatus(model.TaskStatusPending),
	}

	resp, err := taskHandler.CreateTask(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	mockService.AssertNotCalled(t, "CreateTask")
}

func TestGetTask(t *testing.T) {
	mockService := new(MockTaskService)
	taskHandler := handler.NewTaskHandler(mockService)

	ctx := context.Background()
	taskID := primitive.NewObjectID().Hex()
	req := &pb.GetTaskRequest{
		TaskId: taskID,
	}

	expectedTask := &model.Task{
		ID:          primitive.NewObjectID(),
		Title:       "Test Task",
		Description: "Test Description",
		UserID:      "user123",
		Status:      model.TaskStatusPending,
	}

	mockService.On("GetTask", ctx, taskID).Return(expectedTask, nil)

	resp, err := taskHandler.GetTask(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, expectedTask.ID.Hex(), resp.Id)
	assert.Equal(t, expectedTask.Title, resp.Title)
	mockService.AssertExpectations(t)
}

func TestListTasks(t *testing.T) {
	mockService := new(MockTaskService)
	taskHandler := handler.NewTaskHandler(mockService)

	ctx := context.WithValue(context.Background(), "user_id", "user123")
	req := &pb.ListTasksRequest{}

	tasks := []*model.Task{
		{
			ID:          primitive.NewObjectID(),
			Title:       "Task 1",
			Description: "Description 1",
			UserID:      "user123",
			Status:      model.TaskStatusPending,
		},
		{
			ID:          primitive.NewObjectID(),
			Title:       "Task 2",
			Description: "Description 2",
			UserID:      "user123",
			Status:      model.TaskStatusActive,
		},
	}

	mockService.On("ListTasks", ctx, "user123").Return(tasks, nil)

	resp, err := taskHandler.ListTasks(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Tasks, 2)
	assert.Equal(t, tasks[0].ID.Hex(), resp.Tasks[0].Id)
	assert.Equal(t, tasks[0].Title, resp.Tasks[0].Title)
	mockService.AssertExpectations(t)
}

func TestUpdateTask(t *testing.T) {
	mockService := new(MockTaskService)
	taskHandler := handler.NewTaskHandler(mockService)

	ctx := context.Background()
	taskID := primitive.NewObjectID()
	req := &pb.UpdateTaskRequest{
		TaskId:      taskID.Hex(),
		Title:       "Updated Task",
		Description: "Updated Description",
		Status:      pb.TaskStatus(model.TaskStatusActive),
	}

	expectedTask := &model.Task{
		ID:          taskID,
		Title:       req.Title,
		Description: req.Description,
		UserID:      "user123",
		Status:      model.TaskStatusActive,
	}

	mockService.On("UpdateTask", ctx, mock.AnythingOfType("*model.Task")).Return(expectedTask, nil)

	resp, err := taskHandler.UpdateTask(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, expectedTask.ID.Hex(), resp.Id)
	assert.Equal(t, expectedTask.Title, resp.Title)
	mockService.AssertExpectations(t)
}

func TestDeleteTask(t *testing.T) {
	mockService := new(MockTaskService)
	taskHandler := handler.NewTaskHandler(mockService)

	ctx := context.Background()
	taskID := primitive.NewObjectID()
	req := &pb.DeleteTaskRequest{
		TaskId: taskID.Hex(),
	}

	deletedTask := &model.Task{
		ID:     taskID,
		UserID: "user123",
	}

	mockService.On("DeleteTask", ctx, taskID.Hex()).Return(deletedTask, nil)

	resp, err := taskHandler.DeleteTask(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
	mockService.AssertExpectations(t)
}
