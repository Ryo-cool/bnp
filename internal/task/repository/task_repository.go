package repository

import (
	"context"
	"time"

	"my-backend-project/internal/task/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TaskRepository interface {
	Create(ctx context.Context, task *model.Task) error
	FindByID(ctx context.Context, id string) (*model.Task, error)
	FindByUserID(ctx context.Context, userID string, status *model.TaskStatus, limit int32, offset string) ([]*model.Task, int32, error)
	Update(ctx context.Context, task *model.Task) error
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

func (r *mongoTaskRepository) Create(ctx context.Context, task *model.Task) error {
	now := time.Now()
	task.CreatedAt = now
	task.UpdatedAt = now

	_, err := r.collection.InsertOne(ctx, task)
	return err
}

func (r *mongoTaskRepository) FindByID(ctx context.Context, id string) (*model.Task, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var task model.Task
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&task)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &task, nil
}

func (r *mongoTaskRepository) FindByUserID(ctx context.Context, userID string, status *model.TaskStatus, limit int32, offset string) ([]*model.Task, int32, error) {
	filter := bson.M{"user_id": userID}
	if status != nil {
		filter["status"] = status
	}

	// オプションの設定
	findOptions := options.Find()
	if limit > 0 {
		findOptions.SetLimit(int64(limit))
	}
	if offset != "" {
		objectID, err := primitive.ObjectIDFromHex(offset)
		if err != nil {
			return nil, 0, err
		}
		filter["_id"] = bson.M{"$gt": objectID}
	}

	// 総件数の取得
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// タスクの取得
	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var tasks []*model.Task
	if err = cursor.All(ctx, &tasks); err != nil {
		return nil, 0, err
	}

	return tasks, int32(total), nil
}

func (r *mongoTaskRepository) Update(ctx context.Context, task *model.Task) error {
	task.UpdatedAt = time.Now()

	filter := bson.M{"_id": task.ID}
	update := bson.M{"$set": task}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

func (r *mongoTaskRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}
