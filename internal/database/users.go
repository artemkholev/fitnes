package database

import (
	"context"
	"fitness-bot/internal/models"
)

// CreateUser создаёт нового пользователя
func (db *DB) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (telegram_id, username, full_name)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	return db.Pool.QueryRow(ctx, query,
		user.TelegramID,
		user.Username,
		user.FullName,
	).Scan(&user.ID, &user.CreatedAt)
}

// GetUserByTelegramID получает пользователя по telegram_id
func (db *DB) GetUserByTelegramID(ctx context.Context, telegramID int64) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, telegram_id, username, full_name, created_at
		FROM users
		WHERE telegram_id = $1
	`
	err := db.Pool.QueryRow(ctx, query, telegramID).Scan(
		&user.ID,
		&user.TelegramID,
		&user.Username,
		&user.FullName,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// EnsureUser создаёт или обновляет пользователя
func (db *DB) EnsureUser(ctx context.Context, telegramID int64, username, fullName string) error {
	username = NormalizeUsername(username)
	query := `
		INSERT INTO users (telegram_id, username, full_name)
		VALUES ($1, $2, $3)
		ON CONFLICT (telegram_id)
		DO UPDATE SET username = EXCLUDED.username, full_name = EXCLUDED.full_name
	`
	_, err := db.Pool.Exec(ctx, query, telegramID, username, fullName)
	return err
}

// GetUserByUsername получает пользователя по username
func (db *DB) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	username = NormalizeUsername(username)
	user := &models.User{}
	query := `
		SELECT id, telegram_id, username, full_name, created_at
		FROM users
		WHERE username = $1
	`
	err := db.Pool.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.TelegramID,
		&user.Username,
		&user.FullName,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}
