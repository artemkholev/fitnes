package database

import (
	"fitness-bot/internal/models"
	"time"
)

// CreateWorkout создаёт новую тренировку
func (db *DB) CreateWorkout(workout *models.Workout) error {
	return db.GORM.Create(workout).Error
}

// GetWorkoutsByClientTelegramID возвращает последние тренировки клиента
func (db *DB) GetWorkoutsByClientTelegramID(telegramID int64, limit int) ([]*models.Workout, error) {
	var workouts []*models.Workout
	err := db.GORM.
		Where("client_telegram_id = ?", telegramID).
		Order("date DESC").
		Limit(limit).
		Find(&workouts).Error
	return workouts, err
}

// GetWorkoutsByTrainerClient возвращает тренировки клиента через trainer_client_id
func (db *DB) GetWorkoutsByTrainerClient(trainerClientID int64, limit int) ([]*models.Workout, error) {
	var workouts []*models.Workout
	err := db.GORM.
		Where("trainer_client_id = ?", trainerClientID).
		Order("date DESC").
		Limit(limit).
		Find(&workouts).Error
	return workouts, err
}

// GetWorkoutsByMuscleGroup возвращает тренировки по группе мышц
func (db *DB) GetWorkoutsByMuscleGroup(telegramID int64, muscleGroup models.MuscleGroup) ([]*models.Workout, error) {
	var workouts []*models.Workout
	err := db.GORM.
		Where("client_telegram_id = ? AND muscle_group = ?", telegramID, muscleGroup).
		Order("date DESC").
		Find(&workouts).Error
	return workouts, err
}

// CreateExercise создаёт новое упражнение в тренировке
func (db *DB) CreateExercise(exercise *models.Exercise) error {
	return db.GORM.Create(exercise).Error
}

// GetExercisesByWorkout возвращает все упражнения тренировки
func (db *DB) GetExercisesByWorkout(workoutID int64) ([]*models.Exercise, error) {
	var exercises []*models.Exercise
	err := db.GORM.
		Where("workout_id = ?", workoutID).
		Order("\"order\" ASC").
		Find(&exercises).Error
	return exercises, err
}

// GetExerciseStats возвращает статистику упражнения за период
func (db *DB) GetExerciseStats(telegramID int64, exerciseName string, from, to time.Time) ([]*models.Exercise, error) {
	var exercises []*models.Exercise
	err := db.GORM.
		Joins("JOIN workouts ON exercises.workout_id = workouts.id").
		Where("workouts.client_telegram_id = ? AND exercises.name = ? AND workouts.date BETWEEN ? AND ?",
			telegramID, exerciseName, from, to).
		Order("workouts.date DESC").
		Find(&exercises).Error
	return exercises, err
}

// GetWorkoutByID возвращает тренировку по ID
func (db *DB) GetWorkoutByID(id int64) (*models.Workout, error) {
	var workout models.Workout
	err := db.GORM.First(&workout, id).Error
	if err != nil {
		return nil, err
	}
	return &workout, nil
}
