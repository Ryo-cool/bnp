package repository

import (
	"context"
	"time"

	"github.com/my-backend-project/internal/pkg/errors"
	"github.com/my-backend-project/internal/task/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ErrTaskNotFound is returned when a task is not found
var ErrTaskNotFound = errors.NewNotFoundError("タスクが見つかりません", nil)

type TaskRepository interface {
	Create(ctx context.Context, task *model.Task) (*model.Task, error)
	FindByID(ctx context.Context, id string) (*model.Task, error)
	FindByUserID(ctx context.Context, userID string, status *model.TaskStatus, limit int32, offset string) ([]*model.Task, int32, error)
	Update(ctx context.Context, id string, task *model.Task) (*model.Task, error)
	Delete(ctx context.Context, id string) error
}

type mongoTaskRepository struct {
	collection *mongo.Collection
}

func NewTaskRepository(db *mongo.Database) TaskRepository {
	return &mongoTaskRepository{
		collection: db.Collection("tasks"),
	}
}

func (r *mongoTaskRepository) Create(ctx context.Context, task *model.Task) (*model.Task, error) {
	task.ID = primitive.NewObjectID()
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, task)
	if err != nil {
		return nil, errors.NewInternalError("タスクの作成に失敗しました", err)
	}

	return task, nil
}

func (r *mongoTaskRepository) FindByID(ctx context.Context, id string) (*model.Task, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.NewInvalidInputError("無効なIDです", err)
	}

	var task model.Task
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&task)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrTaskNotFound
		}
		return nil, errors.NewInternalError("タスクの取得に失敗しました", err)
	}

	return &task, nil
}

func (r *mongoTaskRepository) FindByUserID(ctx context.Context, userID string, status *model.TaskStatus, limit int32, offset string) ([]*model.Task, int32, error) {
	filter := bson.M{"user_id": userID}
	if status != nil {
		filter["status"] = *status
	}

	opts := options.Find().SetLimit(int64(limit))
	if offset != "" {
		objectID, err := primitive.ObjectIDFromHex(offset)
		if err != nil {
			return nil, 0, errors.NewInvalidInputError("無効なオフセットです", err)
		}
		filter["_id"] = bson.M{"$gt": objectID}
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, errors.NewInternalError("タスクの取得に失敗しました", err)
	}
	defer cursor.Close(ctx)

	var tasks []*model.Task
	if err := cursor.All(ctx, &tasks); err != nil {
		return nil, 0, errors.NewInternalError("タスクの取得に失敗しました", err)
	}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, errors.NewInternalError("タスクの総数の取得に失敗しました", err)
	}

	return tasks, int32(total), nil
}

func (r *mongoTaskRepository) Update(ctx context.Context, id string, task *model.Task) (*model.Task, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.NewInvalidInputError("無効なIDです", err)
	}

	task.UpdatedAt = time.Now()

	update := bson.M{
		"$set": bson.M{
			"title":       task.Title,
			"description": task.Description,
			"status":      task.Status,
			"due_date":    task.DueDate,
			"updated_at":  task.UpdatedAt,
		},
	}

	var updatedTask model.Task
	err = r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": objectID},
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updatedTask)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrTaskNotFound
		}
		return nil, errors.NewInternalError("タスクの更新に失敗しました", err)
	}

	return &updatedTask, nil
}

func (r *mongoTaskRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.NewInvalidInputError("無効なIDです", err)
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return errors.NewInternalError("タスクの削除に失敗しました", err)
	}

	if result.DeletedCount == 0 {
		return ErrTaskNotFound
	}

	return nil
}
