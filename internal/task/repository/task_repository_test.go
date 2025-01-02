package repository

import (
	"context"
	"my-backend-project/internal/pkg/errors"
	"my-backend-project/internal/task/model"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestMongoTaskRepository_Create(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

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
		assert.IsType(t, &errors.AppError{}, err)
	})
}

func TestMongoTaskRepository_FindByID(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

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
			{"_id", taskID},
			{"user_id", expectedTask.UserID},
			{"title", expectedTask.Title},
			{"description", expectedTask.Description},
			{"status", expectedTask.Status},
			{"created_at", expectedTask.CreatedAt},
			{"updated_at", expectedTask.UpdatedAt},
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
		assert.IsType(t, &errors.AppError{}, err)
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
		assert.IsType(t, &errors.AppError{}, err)
	})
}

func TestMongoTaskRepository_Update(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

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
			{"ok", 1},
			{"value", bson.D{
				{"_id", taskID},
				{"user_id", task.UserID},
				{"title", task.Title},
				{"description", task.Description},
				{"status", task.Status},
				{"created_at", task.CreatedAt},
				{"updated_at", task.UpdatedAt},
			}},
			{"lastErrorObject", bson.D{
				{"n", 1},
				{"updatedExisting", true},
			}},
		})

		result, err := repo.Update(context.Background(), task)
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
			{"ok", 1},
			{"value", nil},
			{"lastErrorObject", bson.D{
				{"n", 0},
				{"updatedExisting", false},
			}},
		})

		result, err := repo.Update(context.Background(), task)
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

		result, err := repo.Update(context.Background(), task)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.IsType(t, &errors.AppError{}, err)
	})
}

func TestMongoTaskRepository_Delete(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

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

		mt.AddMockResponses(bson.D{
			{"ok", 1},
			{"value", bson.D{
				{"_id", taskID},
				{"user_id", expectedTask.UserID},
				{"title", expectedTask.Title},
				{"description", expectedTask.Description},
				{"status", expectedTask.Status},
				{"created_at", expectedTask.CreatedAt},
				{"updated_at", expectedTask.UpdatedAt},
			}},
			{"lastErrorObject", bson.D{
				{"n", 1},
			}},
		})

		result, err := repo.Delete(context.Background(), taskID.Hex())
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedTask.ID, result.ID)
	})

	mt.Run("invalid_id", func(mt *mtest.T) {
		repo := &mongoTaskRepository{collection: mt.Coll}
		result, err := repo.Delete(context.Background(), "invalid-id")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.IsType(t, &errors.AppError{}, err)
	})

	mt.Run("not_found", func(mt *mtest.T) {
		repo := &mongoTaskRepository{collection: mt.Coll}
		taskID := primitive.NewObjectID()

		mt.AddMockResponses(bson.D{
			{"ok", 1},
			{"value", nil},
			{"lastErrorObject", bson.D{
				{"n", 0},
			}},
		})

		result, err := repo.Delete(context.Background(), taskID.Hex())
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

		result, err := repo.Delete(context.Background(), taskID.Hex())
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.IsType(t, &errors.AppError{}, err)
	})
}
