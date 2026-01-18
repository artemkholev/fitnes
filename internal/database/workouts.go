package database

import (
	"context"
	"fitness-bot/internal/models"
	"time"
)

func (db *DB) CreateWorkout(ctx context.Context, workout *models.Workout) error {
	query := `
		INSERT INTO workouts (trainer_client_id, client_telegram_id, date, notes, muscle_group)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`
	return db.Pool.QueryRow(ctx, query,
		workout.TrainerClientID,
		workout.ClientTelegramID,
		workout.Date,
		workout.Notes,
		workout.MuscleGroup,
	).Scan(&workout.ID, &workout.CreatedAt)
}

func (db *DB) GetWorkoutsByClientTelegramID(ctx context.Context, telegramID int64, limit int) ([]*models.Workout, error) {
	query := `
		SELECT id, trainer_client_id, client_telegram_id, date, notes, muscle_group, created_at
		FROM workouts
		WHERE client_telegram_id = $1
		ORDER BY date DESC
		LIMIT $2
	`
	rows, err := db.Pool.Query(ctx, query, telegramID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workouts []*models.Workout
	for rows.Next() {
		w := &models.Workout{}
		if err := rows.Scan(&w.ID, &w.TrainerClientID, &w.ClientTelegramID, &w.Date, &w.Notes, &w.MuscleGroup, &w.CreatedAt); err != nil {
			return nil, err
		}
		workouts = append(workouts, w)
	}
	return workouts, rows.Err()
}

func (db *DB) GetWorkoutsByTrainerClient(ctx context.Context, trainerClientID int64, limit int) ([]*models.Workout, error) {
	query := `
		SELECT id, trainer_client_id, client_telegram_id, date, notes, muscle_group, created_at
		FROM workouts
		WHERE trainer_client_id = $1
		ORDER BY date DESC
		LIMIT $2
	`
	rows, err := db.Pool.Query(ctx, query, trainerClientID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workouts []*models.Workout
	for rows.Next() {
		w := &models.Workout{}
		if err := rows.Scan(&w.ID, &w.TrainerClientID, &w.ClientTelegramID, &w.Date, &w.Notes, &w.MuscleGroup, &w.CreatedAt); err != nil {
			return nil, err
		}
		workouts = append(workouts, w)
	}
	return workouts, rows.Err()
}

func (db *DB) GetWorkoutsByMuscleGroup(ctx context.Context, telegramID int64, muscleGroup models.MuscleGroup) ([]*models.Workout, error) {
	query := `
		SELECT id, trainer_client_id, client_telegram_id, date, notes, muscle_group, created_at
		FROM workouts
		WHERE client_telegram_id = $1 AND muscle_group = $2
		ORDER BY date DESC
	`
	rows, err := db.Pool.Query(ctx, query, telegramID, muscleGroup)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workouts []*models.Workout
	for rows.Next() {
		w := &models.Workout{}
		if err := rows.Scan(&w.ID, &w.TrainerClientID, &w.ClientTelegramID, &w.Date, &w.Notes, &w.MuscleGroup, &w.CreatedAt); err != nil {
			return nil, err
		}
		workouts = append(workouts, w)
	}
	return workouts, rows.Err()
}

func (db *DB) CreateExercise(ctx context.Context, exercise *models.Exercise) error {
	query := `
		INSERT INTO exercises (workout_id, name, sets, reps, weight, rest_seconds, photo_file_id, notes, "order")
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at
	`
	return db.Pool.QueryRow(ctx, query,
		exercise.WorkoutID,
		exercise.Name,
		exercise.Sets,
		exercise.Reps,
		exercise.Weight,
		exercise.RestSeconds,
		exercise.PhotoFileID,
		exercise.Notes,
		exercise.Order,
	).Scan(&exercise.ID, &exercise.CreatedAt)
}

func (db *DB) GetExercisesByWorkout(ctx context.Context, workoutID int64) ([]*models.Exercise, error) {
	query := `
		SELECT id, workout_id, name, sets, reps, weight, rest_seconds, photo_file_id, notes, "order", created_at
		FROM exercises
		WHERE workout_id = $1
		ORDER BY "order"
	`
	rows, err := db.Pool.Query(ctx, query, workoutID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exercises []*models.Exercise
	for rows.Next() {
		e := &models.Exercise{}
		if err := rows.Scan(&e.ID, &e.WorkoutID, &e.Name, &e.Sets, &e.Reps, &e.Weight,
			&e.RestSeconds, &e.PhotoFileID, &e.Notes, &e.Order, &e.CreatedAt); err != nil {
			return nil, err
		}
		exercises = append(exercises, e)
	}
	return exercises, rows.Err()
}

func (db *DB) GetExerciseStats(ctx context.Context, telegramID int64, exerciseName string, from, to time.Time) ([]*models.Exercise, error) {
	query := `
		SELECT e.id, e.workout_id, e.name, e.sets, e.reps, e.weight,
		       e.rest_seconds, e.photo_file_id, e.notes, e."order", e.created_at
		FROM exercises e
		JOIN workouts w ON e.workout_id = w.id
		WHERE w.client_telegram_id = $1 AND e.name = $2 AND w.date BETWEEN $3 AND $4
		ORDER BY w.date DESC
	`
	rows, err := db.Pool.Query(ctx, query, telegramID, exerciseName, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exercises []*models.Exercise
	for rows.Next() {
		e := &models.Exercise{}
		if err := rows.Scan(&e.ID, &e.WorkoutID, &e.Name, &e.Sets, &e.Reps, &e.Weight,
			&e.RestSeconds, &e.PhotoFileID, &e.Notes, &e.Order, &e.CreatedAt); err != nil {
			return nil, err
		}
		exercises = append(exercises, e)
	}
	return exercises, rows.Err()
}

// GetWorkoutByID возвращает тренировку по ID
func (db *DB) GetWorkoutByID(ctx context.Context, id int64) (*models.Workout, error) {
	w := &models.Workout{}
	query := `
		SELECT id, trainer_client_id, client_telegram_id, date, notes, muscle_group, created_at
		FROM workouts
		WHERE id = $1
	`
	err := db.Pool.QueryRow(ctx, query, id).Scan(
		&w.ID, &w.TrainerClientID, &w.ClientTelegramID, &w.Date, &w.Notes, &w.MuscleGroup, &w.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return w, nil
}
