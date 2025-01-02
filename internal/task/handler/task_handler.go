package handler

import (
	"context"

	"my-backend-project/internal/task/model"
	"my-backend-project/internal/task/pb"
	"my-backend-project/internal/task/service"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		return nil, status.Error(codes.Internal, "could not get user_id from context")
	}

	task := &model.Task{
		Title:       req.Title,
		Description: req.Description,
		UserID:      userID,
		Status:      model.TaskStatus(req.Status),
	}

	createdTask, err := h.taskService.CreateTask(ctx, task)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create task")
	}

	return &pb.Task{
		Id:          createdTask.ID.Hex(),
		Title:       createdTask.Title,
		Description: createdTask.Description,
		Status:      pb.TaskStatus(createdTask.Status),
		UserId:      createdTask.UserID,
	}, nil
}

func (h *TaskHandler) GetTask(ctx context.Context, req *pb.GetTaskRequest) (*pb.Task, error) {
	objectID, err := primitive.ObjectIDFromHex(req.TaskId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid task id")
	}

	task, err := h.taskService.GetTask(ctx, objectID.Hex())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get task")
	}

	return &pb.Task{
		Id:          task.ID.Hex(),
		Title:       task.Title,
		Description: task.Description,
		Status:      pb.TaskStatus(task.Status),
		UserId:      task.UserID,
	}, nil
}

func (h *TaskHandler) ListTasks(ctx context.Context, req *pb.ListTasksRequest) (*pb.ListTasksResponse, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return nil, status.Error(codes.Internal, "could not get user_id from context")
	}

	tasks, err := h.taskService.ListTasks(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list tasks")
	}

	var taskResponses []*pb.Task
	for _, task := range tasks {
		taskResponses = append(taskResponses, &pb.Task{
			Id:          task.ID.Hex(),
			Title:       task.Title,
			Description: task.Description,
			Status:      pb.TaskStatus(task.Status),
			UserId:      task.UserID,
		})
	}

	return &pb.ListTasksResponse{
		Tasks: taskResponses,
	}, nil
}

func (h *TaskHandler) UpdateTask(ctx context.Context, req *pb.UpdateTaskRequest) (*pb.Task, error) {
	objectID, err := primitive.ObjectIDFromHex(req.TaskId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid task id")
	}

	task := &model.Task{
		ID:          objectID,
		Title:       req.Title,
		Description: req.Description,
		Status:      model.TaskStatus(req.Status),
	}

	updatedTask, err := h.taskService.UpdateTask(ctx, task)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update task")
	}

	return &pb.Task{
		Id:          updatedTask.ID.Hex(),
		Title:       updatedTask.Title,
		Description: updatedTask.Description,
		Status:      pb.TaskStatus(updatedTask.Status),
		UserId:      updatedTask.UserID,
	}, nil
}

func (h *TaskHandler) DeleteTask(ctx context.Context, req *pb.DeleteTaskRequest) (*pb.DeleteTaskResponse, error) {
	objectID, err := primitive.ObjectIDFromHex(req.TaskId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid task id")
	}

	_, err = h.taskService.DeleteTask(ctx, objectID.Hex())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to delete task")
	}

	return &pb.DeleteTaskResponse{
		Success: true,
	}, nil
}
