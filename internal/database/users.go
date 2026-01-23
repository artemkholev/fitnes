package database

import (
	"fitness-bot/internal/models"

	"gorm.io/gorm/clause"
)

// CreateUser создаёт нового пользователя
func (db *DB) CreateUser(user *models.User) error {
	return db.GORM.Create(user).Error
}

// GetUserByTelegramID получает пользователя по telegram_id
func (db *DB) GetUserByTelegramID(telegramID int64) (*models.User, error) {
	var user models.User
	err := db.GORM.Where("telegram_id = ?", telegramID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// EnsureUser создаёт или обновляет пользователя
func (db *DB) EnsureUser(telegramID int64, username, fullName string) error {
	username = NormalizeUsername(username)
	user := models.User{
		TelegramID: telegramID,
		Username:   username,
		FullName:   fullName,
	}
	// Upsert: создаём или обновляем
	return db.GORM.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "telegram_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"username", "full_name"}),
	}).Create(&user).Error
}

// GetUserByUsername получает пользователя по username
func (db *DB) GetUserByUsername(username string) (*models.User, error) {
	username = NormalizeUsername(username)
	var user models.User
	err := db.GORM.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
