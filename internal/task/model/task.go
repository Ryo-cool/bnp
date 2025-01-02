package model

import (
	"time"

	"github.com/my-backend-project/internal/task/pb"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TaskStatus int32

const (
	TaskStatusUnspecified TaskStatus = 0
	TaskStatusTodo        TaskStatus = 1
	TaskStatusInProgress  TaskStatus = 2
	TaskStatusDone        TaskStatus = 3
)

type Task struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	UserID      string             `bson:"user_id" json:"user_id"`
	Status      TaskStatus         `bson:"status" json:"status"`
	DueDate     time.Time          `bson:"due_date" json:"due_date"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// ToProto converts Task to protobuf Task message
func (t *Task) ToProto() *pb.Task {
	return &pb.Task{
		Id:          t.ID.Hex(),
		Title:       t.Title,
		Description: t.Description,
		UserId:      t.UserID,
		Status:      pb.TaskStatus(t.Status),
		DueDate:     timestamppb.New(t.DueDate),
		CreatedAt:   timestamppb.New(t.CreatedAt),
		UpdatedAt:   timestamppb.New(t.UpdatedAt),
	}
}

// FromProto creates Task from protobuf Task message
func TaskFromProto(pt *pb.Task) (*Task, error) {
	id, err := primitive.ObjectIDFromHex(pt.Id)
	if err != nil {
		return nil, err
	}

	return &Task{
		ID:          id,
		Title:       pt.Title,
		Description: pt.Description,
		UserID:      pt.UserId,
		Status:      TaskStatus(pt.Status),
		DueDate:     pt.DueDate.AsTime(),
		CreatedAt:   pt.CreatedAt.AsTime(),
		UpdatedAt:   pt.UpdatedAt.AsTime(),
	}, nil
}

// Validation rules
type CreateTaskRequest struct {
	Title       string     `json:"title" validate:"required,min=1,max=100"`
	Description string     `json:"description" validate:"max=1000"`
	UserID      string     `json:"user_id" validate:"required"`
	Status      TaskStatus `json:"status" validate:"required,min=0,max=3"`
	DueDate     time.Time  `json:"due_date" validate:"required"`
}

type UpdateTaskRequest struct {
	TaskID      string     `json:"task_id" validate:"required"`
	UserID      string     `json:"user_id" validate:"required"`
	Title       string     `json:"title" validate:"required,min=1,max=100"`
	Description string     `json:"description" validate:"max=1000"`
	Status      TaskStatus `json:"status" validate:"required,min=0,max=3"`
	DueDate     time.Time  `json:"due_date" validate:"required"`
}
