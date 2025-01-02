package handler

import (
	"context"

	"github.com/my-backend-project/internal/pb"
	"github.com/my-backend-project/internal/pkg/apperrors"
	"github.com/my-backend-project/internal/task/model"
	"github.com/my-backend-project/internal/task/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
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

func (h *TaskHandler) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.CreateTaskResponse, error) {
	task := &model.Task{
		UserID:      req.UserId,
		Title:       req.Title,
		Description: req.Description,
		Status:      model.TaskStatus(req.Status.String()),
		DueDate:     req.DueDate.AsTime(),
	}

	createdTask, err := h.taskService.CreateTask(ctx, task)
	if err != nil {
		return nil, convertErrorToGRPCStatus(err)
	}

	return &pb.CreateTaskResponse{
		TaskId: createdTask.ID.Hex(),
	}, nil
}

func (h *TaskHandler) GetTask(ctx context.Context, req *pb.GetTaskRequest) (*pb.GetTaskResponse, error) {
	task, err := h.taskService.GetTask(ctx, req.TaskId)
	if err != nil {
		return nil, convertErrorToGRPCStatus(err)
	}

	return &pb.GetTaskResponse{
		Task: convertTaskToProto(task),
	}, nil
}

func (h *TaskHandler) ListTasks(ctx context.Context, req *pb.ListTasksRequest) (*pb.ListTasksResponse, error) {
	var taskStatus *model.TaskStatus
	if req.Status != pb.TaskStatus_TASK_STATUS_UNSPECIFIED {
		status := model.TaskStatus(req.Status.String())
		taskStatus = &status
	}

	tasks, total, err := h.taskService.ListTasks(ctx, req.UserId, taskStatus, req.PageSize, req.PageToken)
	if err != nil {
		return nil, convertErrorToGRPCStatus(err)
	}

	var nextPageToken string
	if len(tasks) > 0 {
		nextPageToken = tasks[len(tasks)-1].ID.Hex()
	}

	taskResponses := make([]*pb.Task, len(tasks))
	for i, task := range tasks {
		taskResponses[i] = convertTaskToProto(task)
	}

	return &pb.ListTasksResponse{
		Tasks:         taskResponses,
		NextPageToken: nextPageToken,
		TotalCount:    total,
	}, nil
}

func (h *TaskHandler) UpdateTask(ctx context.Context, req *pb.UpdateTaskRequest) (*pb.UpdateTaskResponse, error) {
	task := &model.Task{
		UserID:      req.UserId,
		Title:       req.Title,
		Description: req.Description,
		Status:      model.TaskStatus(req.Status.String()),
		DueDate:     req.DueDate.AsTime(),
	}

	updatedTask, err := h.taskService.UpdateTask(ctx, req.TaskId, task)
	if err != nil {
		return nil, convertErrorToGRPCStatus(err)
	}

	return &pb.UpdateTaskResponse{
		Task: convertTaskToProto(updatedTask),
	}, nil
}

func (h *TaskHandler) DeleteTask(ctx context.Context, req *pb.DeleteTaskRequest) (*pb.Empty, error) {
	err := h.taskService.DeleteTask(ctx, req.TaskId)
	if err != nil {
		return nil, convertErrorToGRPCStatus(err)
	}

	return &pb.Empty{}, nil
}

func convertTaskToProto(task *model.Task) *pb.Task {
	var status pb.TaskStatus
	switch task.Status {
	case model.TaskStatusPending:
		status = pb.TaskStatus_TASK_STATUS_PENDING
	case model.TaskStatusActive:
		status = pb.TaskStatus_TASK_STATUS_ACTIVE
	case model.TaskStatusComplete:
		status = pb.TaskStatus_TASK_STATUS_COMPLETE
	default:
		status = pb.TaskStatus_TASK_STATUS_UNSPECIFIED
	}

	return &pb.Task{
		TaskId:      task.ID.Hex(),
		UserId:      task.UserID,
		Title:       task.Title,
		Description: task.Description,
		Status:      status,
		DueDate:     timestamppb.New(task.DueDate),
		CreatedAt:   timestamppb.New(task.CreatedAt),
		UpdatedAt:   timestamppb.New(task.UpdatedAt),
	}
}

func convertErrorToGRPCStatus(err error) error {
	if apperrors.IsNotFound(err) {
		return status.Error(codes.NotFound, err.Error())
	}
	if apperrors.IsInvalidInput(err) {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return status.Error(codes.Internal, err.Error())
}
