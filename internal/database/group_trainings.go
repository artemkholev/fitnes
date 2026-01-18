package database

import (
	"context"
	"fitness-bot/internal/models"
	"time"
)

func (db *DB) CreateGroupTraining(ctx context.Context, gt *models.GroupTraining) error {
	query := `
		INSERT INTO group_trainings (organization_id, trainer_id, name, description, scheduled_at, max_participants)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`
	return db.Pool.QueryRow(ctx, query,
		gt.OrganizationID,
		gt.TrainerID,
		gt.Name,
		gt.Description,
		gt.ScheduledAt,
		gt.MaxParticipants,
	).Scan(&gt.ID, &gt.CreatedAt)
}

func (db *DB) GetUpcomingGroupTrainings(ctx context.Context, orgID int64) ([]*models.GroupTraining, error) {
	query := `
		SELECT id, organization_id, trainer_id, name, description, scheduled_at, max_participants, created_at
		FROM group_trainings
		WHERE organization_id = $1 AND scheduled_at > $2
		ORDER BY scheduled_at ASC
	`
	rows, err := db.Pool.Query(ctx, query, orgID, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trainings []*models.GroupTraining
	for rows.Next() {
		gt := &models.GroupTraining{}
		if err := rows.Scan(&gt.ID, &gt.OrganizationID, &gt.TrainerID, &gt.Name,
			&gt.Description, &gt.ScheduledAt, &gt.MaxParticipants, &gt.CreatedAt); err != nil {
			return nil, err
		}
		trainings = append(trainings, gt)
	}
	return trainings, rows.Err()
}

func (db *DB) JoinGroupTraining(ctx context.Context, trainingID, userID int64) error {
	query := `INSERT INTO group_training_participants (group_training_id, user_id) VALUES ($1, $2)`
	_, err := db.Pool.Exec(ctx, query, trainingID, userID)
	return err
}

func (db *DB) GetGroupTrainingParticipants(ctx context.Context, trainingID int64) ([]*models.User, error) {
	query := `
		SELECT u.id, u.telegram_id, u.username, u.full_name, u.role, u.organization_id, u.trainer_id, u.created_at
		FROM users u
		JOIN group_training_participants gtp ON u.id = gtp.user_id
		WHERE gtp.group_training_id = $1
	`
	rows, err := db.Pool.Query(ctx, query, trainingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		u := &models.User{}
		if err := rows.Scan(&u.ID, &u.TelegramID, &u.Username, &u.FullName,
			&u.Role, &u.OrganizationID, &u.TrainerID, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (db *DB) GetParticipantCount(ctx context.Context, trainingID int64) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM group_training_participants WHERE group_training_id = $1`
	err := db.Pool.QueryRow(ctx, query, trainingID).Scan(&count)
	return count, err
}
