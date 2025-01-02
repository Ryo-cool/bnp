package repository

import (
	"context"
	"testing"
	"time"

	"github.com/my-backend-project/internal/pkg/apperrors"
	"github.com/my-backend-project/internal/task/model"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestMongoTaskRepository_Create(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := &mongoTaskRepository{collection: mt.Coll}
		task := &model.Task{
			UserID:      "user1",
			Title:       "Test Task",
			Description: "Test Description",
			Status:      model.TaskStatusPending,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		result, err := repo.Create(context.Background(), task)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.ID)
	})

	mt.Run("database_error", func(mt *mtest.T) {
		repo := &mongoTaskRepository{collection: mt.Coll}
		task := &model.Task{
			UserID:      "user1",
			Title:       "Test Task",
			Description: "Test Description",
			Status:      model.TaskStatusPending,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code:    11000,
			Message: "duplicate key error",
		}))

		result, err := repo.Create(context.Background(), task)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.IsType(t, &apperrors.AppError{}, err)
	})
}

func TestMongoTaskRepository_FindByID(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := &mongoTaskRepository{collection: mt.Coll}
		taskID := primitive.NewObjectID()
		expectedTask := &model.Task{
			ID:          taskID,
			UserID:      "user1",
			Title:       "Test Task",
			Description: "Test Description",
			Status:      model.TaskStatusPending,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: taskID},
			{Key: "user_id", Value: expectedTask.UserID},
			{Key: "title", Value: expectedTask.Title},
			{Key: "description", Value: expectedTask.Description},
			{Key: "status", Value: expectedTask.Status},
			{Key: "created_at", Value: expectedTask.CreatedAt},
			{Key: "updated_at", Value: expectedTask.UpdatedAt},
		}))

		result, err := repo.FindByID(context.Background(), taskID.Hex())
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedTask.ID, result.ID)
	})

	mt.Run("invalid_id", func(mt *mtest.T) {
		repo := &mongoTaskRepository{collection: mt.Coll}
		result, err := repo.FindByID(context.Background(), "invalid-id")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.IsType(t, &apperrors.AppError{}, err)
	})

	mt.Run("not_found", func(mt *mtest.T) {
		repo := &mongoTaskRepository{collection: mt.Coll}
		taskID := primitive.NewObjectID()

		mt.AddMockResponses(mtest.CreateCursorResponse(0, "foo.bar", mtest.FirstBatch))

		result, err := repo.FindByID(context.Background(), taskID.Hex())
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrTaskNotFound, err)
	})

	mt.Run("database_error", func(mt *mtest.T) {
		repo := &mongoTaskRepository{collection: mt.Coll}
		taskID := primitive.NewObjectID()

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code:    1,
			Message: "internal error",
		}))

		result, err := repo.FindByID(context.Background(), taskID.Hex())
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.IsType(t, &apperrors.AppError{}, err)
	})
}

func TestMongoTaskRepository_Update(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := &mongoTaskRepository{collection: mt.Coll}
		taskID := primitive.NewObjectID()
		task := &model.Task{
			ID:          taskID,
			UserID:      "user1",
			Title:       "Updated Task",
			Description: "Updated Description",
			Status:      model.TaskStatusActive,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "value", Value: bson.D{
				{Key: "_id", Value: taskID},
				{Key: "user_id", Value: task.UserID},
				{Key: "title", Value: task.Title},
				{Key: "description", Value: task.Description},
				{Key: "status", Value: task.Status},
				{Key: "created_at", Value: task.CreatedAt},
				{Key: "updated_at", Value: task.UpdatedAt},
			}},
			{Key: "lastErrorObject", Value: bson.D{
				{Key: "n", Value: 1},
				{Key: "updatedExisting", Value: true},
			}},
		})

		result, err := repo.Update(context.Background(), taskID.Hex(), task)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, task.ID, result.ID)
	})

	mt.Run("not_found", func(mt *mtest.T) {
		repo := &mongoTaskRepository{collection: mt.Coll}
		taskID := primitive.NewObjectID()
		task := &model.Task{
			ID:          taskID,
			UserID:      "user1",
			Title:       "Updated Task",
			Description: "Updated Description",
			Status:      model.TaskStatusActive,
		}

		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "value", Value: nil},
			{Key: "lastErrorObject", Value: bson.D{
				{Key: "n", Value: 0},
				{Key: "updatedExisting", Value: false},
			}},
		})

		result, err := repo.Update(context.Background(), taskID.Hex(), task)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrTaskNotFound, err)
	})

	mt.Run("database_error", func(mt *mtest.T) {
		repo := &mongoTaskRepository{collection: mt.Coll}
		taskID := primitive.NewObjectID()
		task := &model.Task{
			ID:          taskID,
			UserID:      "user1",
			Title:       "Updated Task",
			Description: "Updated Description",
			Status:      model.TaskStatusActive,
		}

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code:    1,
			Message: "internal error",
		}))

		result, err := repo.Update(context.Background(), taskID.Hex(), task)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.IsType(t, &apperrors.AppError{}, err)
	})
}

func TestMongoTaskRepository_Delete(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := &mongoTaskRepository{collection: mt.Coll}
		taskID := primitive.NewObjectID()

		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "n", Value: 1},
			{Key: "acknowledged", Value: true},
		})

		err := repo.Delete(context.Background(), taskID.Hex())
		assert.NoError(t, err)
	})

	mt.Run("invalid_id", func(mt *mtest.T) {
		repo := &mongoTaskRepository{collection: mt.Coll}
		err := repo.Delete(context.Background(), "invalid-id")
		assert.Error(t, err)
		assert.IsType(t, &apperrors.AppError{}, err)
	})

	mt.Run("not_found", func(mt *mtest.T) {
		repo := &mongoTaskRepository{collection: mt.Coll}
		taskID := primitive.NewObjectID()

		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "n", Value: 0},
			{Key: "acknowledged", Value: true},
		})

		err := repo.Delete(context.Background(), taskID.Hex())
		assert.Error(t, err)
		assert.Equal(t, ErrTaskNotFound, err)
	})

	mt.Run("database_error", func(mt *mtest.T) {
		repo := &mongoTaskRepository{collection: mt.Coll}
		taskID := primitive.NewObjectID()

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code:    1,
			Message: "internal error",
		}))

		err := repo.Delete(context.Background(), taskID.Hex())
		assert.Error(t, err)
		assert.IsType(t, &apperrors.AppError{}, err)
	})
}

func TestMongoTaskRepository_FindByUserID(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := &mongoTaskRepository{collection: mt.Coll}
		userID := "user1"
		status := model.TaskStatusPending
		limit := int32(10)
		offset := ""

		task1ID := primitive.NewObjectID()
		task2ID := primitive.NewObjectID()
		now := time.Now()

		// Findのモックレスポンス
		first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: task1ID},
			{Key: "user_id", Value: userID},
			{Key: "title", Value: "Task 1"},
			{Key: "description", Value: "Description 1"},
			{Key: "status", Value: status},
			{Key: "due_date", Value: now},
			{Key: "created_at", Value: now},
			{Key: "updated_at", Value: now},
		})
		second := mtest.CreateCursorResponse(1, "foo.bar", mtest.NextBatch, bson.D{
			{Key: "_id", Value: task2ID},
			{Key: "user_id", Value: userID},
			{Key: "title", Value: "Task 2"},
			{Key: "description", Value: "Description 2"},
			{Key: "status", Value: status},
			{Key: "due_date", Value: now},
			{Key: "created_at", Value: now},
			{Key: "updated_at", Value: now},
		})
		killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)

		// CountDocumentsのモックレスポンス
		count := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{Key: "n", Value: int32(2)},
		})

		mt.AddMockResponses(first, second, killCursors, count)

		tasks, total, err := repo.FindByUserID(context.Background(), userID, &status, limit, offset)
		assert.NoError(t, err)
		assert.NotNil(t, tasks)
		assert.Equal(t, int32(2), total)
		assert.Len(t, tasks, 2)
		assert.Equal(t, task1ID, tasks[0].ID)
		assert.Equal(t, task2ID, tasks[1].ID)
	})

	mt.Run("database_error", func(mt *mtest.T) {
		repo := &mongoTaskRepository{collection: mt.Coll}
		userID := "user1"
		status := model.TaskStatusPending
		limit := int32(10)
		offset := ""

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code:    1,
			Message: "internal error",
		}))

		tasks, total, err := repo.FindByUserID(context.Background(), userID, &status, limit, offset)
		assert.Error(t, err)
		assert.Nil(t, tasks)
		assert.Equal(t, int32(0), total)
		assert.IsType(t, &apperrors.AppError{}, err)
	})
}
