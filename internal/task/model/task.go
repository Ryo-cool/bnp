package model

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TaskStatus string

const (
	TaskStatusPending  TaskStatus = "TASK_STATUS_PENDING"
	TaskStatusActive   TaskStatus = "TASK_STATUS_ACTIVE"
	TaskStatusComplete TaskStatus = "TASK_STATUS_COMPLETE"
)

type Task struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	UserID      string             `bson:"user_id"`
	Title       string             `bson:"title"`
	Description string             `bson:"description"`
	Status      TaskStatus         `bson:"status"`
	DueDate     time.Time          `bson:"due_date"`
	CreatedAt   time.Time          `bson:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"`
}

func (t *Task) Validate() error {
	if t.UserID == "" {
		return errors.New("ユーザーIDは必須です")
	}

	if t.Title == "" {
		return errors.New("タイトルは必須です")
	}

	if t.Status == TaskStatus("") {
		return errors.New("ステータスは必須です")
	}

	if !isValidStatus(t.Status) {
		return errors.New("無効なステータスです")
	}

	if t.DueDate.IsZero() {
		return errors.New("期限は必須です")
	}

	return nil
}

func isValidStatus(status TaskStatus) bool {
	switch status {
	case TaskStatusPending, TaskStatusActive, TaskStatusComplete:
		return true
	default:
		return false
	}
}

// TimeToProtoTimestamp converts time.Time to protobuf Timestamp
func TimeToProtoTimestamp(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

// ProtoTimestampToTime converts protobuf Timestamp to time.Time
func ProtoTimestampToTime(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime()
}
