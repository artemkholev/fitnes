package database

import (
	"context"
	"fitness-bot/internal/models"
)

func (db *DB) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (telegram_id, username, full_name, role, organization_id, trainer_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`
	return db.Pool.QueryRow(ctx, query,
		user.TelegramID,
		user.Username,
		user.FullName,
		user.Role,
		user.OrganizationID,
		user.TrainerID,
	).Scan(&user.ID, &user.CreatedAt)
}

func (db *DB) GetUserByTelegramID(ctx context.Context, telegramID int64) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, telegram_id, username, full_name, role, organization_id, trainer_id, created_at
		FROM users
		WHERE telegram_id = $1
	`
	err := db.Pool.QueryRow(ctx, query, telegramID).Scan(
		&user.ID,
		&user.TelegramID,
		&user.Username,
		&user.FullName,
		&user.Role,
		&user.OrganizationID,
		&user.TrainerID,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (db *DB) UpdateUserTrainer(ctx context.Context, userID, trainerID int64) error {
	query := `UPDATE users SET trainer_id = $1 WHERE id = $2`
	_, err := db.Pool.Exec(ctx, query, trainerID, userID)
	return err
}

func (db *DB) GetTrainersByOrganization(ctx context.Context, orgID int64) ([]*models.User, error) {
	query := `
		SELECT id, telegram_id, username, full_name, role, organization_id, trainer_id, created_at
		FROM users
		WHERE role = 'trainer' AND organization_id = $1
	`
	rows, err := db.Pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trainers []*models.User
	for rows.Next() {
		user := &models.User{}
		if err := rows.Scan(
			&user.ID,
			&user.TelegramID,
			&user.Username,
			&user.FullName,
			&user.Role,
			&user.OrganizationID,
			&user.TrainerID,
			&user.CreatedAt,
		); err != nil {
			return nil, err
		}
		trainers = append(trainers, user)
	}
	return trainers, rows.Err()
}

func (db *DB) GetClientsByTrainer(ctx context.Context, trainerID int64) ([]*models.User, error) {
	query := `
		SELECT id, telegram_id, username, full_name, role, organization_id, trainer_id, created_at
		FROM users
		WHERE trainer_id = $1
	`
	rows, err := db.Pool.Query(ctx, query, trainerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []*models.User
	for rows.Next() {
		user := &models.User{}
		if err := rows.Scan(
			&user.ID,
			&user.TelegramID,
			&user.Username,
			&user.FullName,
			&user.Role,
			&user.OrganizationID,
			&user.TrainerID,
			&user.CreatedAt,
		); err != nil {
			return nil, err
		}
		clients = append(clients, user)
	}
	return clients, rows.Err()
}
