package service_test

import (
	"context"
	"testing"

	"my-backend-project/internal/task/model"
	"my-backend-project/internal/task/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MockTaskRepository struct {
	mock.Mock
}

func (m *MockTaskRepository) Create(ctx context.Context, task *model.Task) (*model.Task, error) {
	args := m.Called(ctx, task)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Task), args.Error(1)
}

func (m *MockTaskRepository) FindByID(ctx context.Context, id string) (*model.Task, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Task), args.Error(1)
}

func (m *MockTaskRepository) FindByUserID(ctx context.Context, userID string, status *model.TaskStatus, limit int32, offset string) ([]*model.Task, int32, error) {
	args := m.Called(ctx, userID, status, limit, offset)
	return args.Get(0).([]*model.Task), args.Get(1).(int32), args.Error(2)
}

func (m *MockTaskRepository) Update(ctx context.Context, task *model.Task) (*model.Task, error) {
	args := m.Called(ctx, task)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Task), args.Error(1)
}

func (m *MockTaskRepository) Delete(ctx context.Context, id string) (*model.Task, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Task), args.Error(1)
}

func TestCreateTask(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	taskService := service.NewTaskService(mockRepo)
	ctx := context.Background()

	task := &model.Task{
		Title:       "Test Task",
		Description: "Test Description",
		UserID:      "user123",
		Status:      model.TaskStatusPending,
	}

	expectedTask := &model.Task{
		ID:          primitive.NewObjectID(),
		Title:       task.Title,
		Description: task.Description,
		UserID:      task.UserID,
		Status:      task.Status,
	}

	mockRepo.On("Create", ctx, task).Return(expectedTask, nil)

	createdTask, err := taskService.CreateTask(ctx, task)

	assert.NoError(t, err)
	assert.NotNil(t, createdTask)
	assert.Equal(t, expectedTask.ID, createdTask.ID)
	assert.Equal(t, expectedTask.Title, createdTask.Title)
	mockRepo.AssertExpectations(t)
}

func TestGetTask(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	taskService := service.NewTaskService(mockRepo)
	ctx := context.Background()

	taskID := primitive.NewObjectID().Hex()
	expectedTask := &model.Task{
		ID:          primitive.NewObjectID(),
		Title:       "Test Task",
		Description: "Test Description",
		UserID:      "user123",
		Status:      model.TaskStatusPending,
	}

	mockRepo.On("FindByID", ctx, taskID).Return(expectedTask, nil)

	task, err := taskService.GetTask(ctx, taskID)

	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, expectedTask.ID, task.ID)
	assert.Equal(t, expectedTask.Title, task.Title)
	mockRepo.AssertExpectations(t)
}
