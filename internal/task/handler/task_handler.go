package handler

import (
	"context"

	"my-backend-project/internal/pkg/errors"
	"my-backend-project/internal/task/model"
	"my-backend-project/internal/task/pb"
	"my-backend-project/internal/task/service"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TaskHandler struct {
	pb.UnimplementedTaskServiceServer
	taskService service.TaskService
}

func NewTaskHandler(taskService service.TaskService) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
	}
}

func (h *TaskHandler) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.Task, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.NewUnauthorizedError("could not get user_id from context", nil).GRPCStatus().Err()
	}

	task := &model.Task{
		Title:       req.Title,
		Description: req.Description,
		UserID:      userID,
		Status:      model.TaskStatus(req.Status),
	}

	createdTask, err := h.taskService.CreateTask(ctx, task)
	if err != nil {
		return nil, errors.NewInternalError("failed to create task", err).GRPCStatus().Err()
	}

	return createdTask.ToProto(), nil
}

func (h *TaskHandler) GetTask(ctx context.Context, req *pb.GetTaskRequest) (*pb.Task, error) {
	objectID, err := primitive.ObjectIDFromHex(req.TaskId)
	if err != nil {
		return nil, errors.NewInvalidInputError("invalid task id", err).GRPCStatus().Err()
	}

	task, err := h.taskService.GetTask(ctx, objectID.Hex())
	if err != nil {
		return nil, errors.NewInternalError("failed to get task", err).GRPCStatus().Err()
	}

	return task.ToProto(), nil
}

func (h *TaskHandler) ListTasks(ctx context.Context, req *pb.ListTasksRequest) (*pb.ListTasksResponse, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.NewUnauthorizedError("could not get user_id from context", nil).GRPCStatus().Err()
	}

	tasks, err := h.taskService.ListTasks(ctx, userID)
	if err != nil {
		return nil, errors.NewInternalError("failed to list tasks", err).GRPCStatus().Err()
	}

	var taskResponses []*pb.Task
	for _, task := range tasks {
		taskResponses = append(taskResponses, task.ToProto())
	}

	return &pb.ListTasksResponse{
		Tasks: taskResponses,
	}, nil
}

func (h *TaskHandler) UpdateTask(ctx context.Context, req *pb.UpdateTaskRequest) (*pb.Task, error) {
	objectID, err := primitive.ObjectIDFromHex(req.TaskId)
	if err != nil {
		return nil, errors.NewInvalidInputError("invalid task id", err).GRPCStatus().Err()
	}

	task := &model.Task{
		ID:          objectID,
		Title:       req.Title,
		Description: req.Description,
		Status:      model.TaskStatus(req.Status),
	}

	updatedTask, err := h.taskService.UpdateTask(ctx, task)
	if err != nil {
		return nil, errors.NewInternalError("failed to update task", err).GRPCStatus().Err()
	}

	return updatedTask.ToProto(), nil
}

func (h *TaskHandler) DeleteTask(ctx context.Context, req *pb.DeleteTaskRequest) (*pb.DeleteTaskResponse, error) {
	objectID, err := primitive.ObjectIDFromHex(req.TaskId)
	if err != nil {
		return nil, errors.NewInvalidInputError("invalid task id", err).GRPCStatus().Err()
	}

	_, err = h.taskService.DeleteTask(ctx, objectID.Hex())
	if err != nil {
		return nil, errors.NewInternalError("failed to delete task", err).GRPCStatus().Err()
	}

	return &pb.DeleteTaskResponse{
		Success: true,
	}, nil
}
