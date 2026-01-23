package database

import (
	"fitness-bot/internal/models"
	"time"
)

// CreateGroupTraining создаёт новую групповую тренировку
func (db *DB) CreateGroupTraining(gt *models.GroupTraining) error {
	return db.GORM.Create(gt).Error
}

// GetUpcomingGroupTrainings возвращает предстоящие тренировки организации
func (db *DB) GetUpcomingGroupTrainings(orgID int64) ([]*models.GroupTraining, error) {
	var trainings []*models.GroupTraining
	err := db.GORM.
		Where("organization_id = ? AND scheduled_at > ?", orgID, time.Now()).
		Order("scheduled_at ASC").
		Find(&trainings).Error
	return trainings, err
}

// JoinGroupTraining добавляет участника к групповой тренировке
func (db *DB) JoinGroupTraining(trainingID, userID int64) error {
	participant := &models.GroupTrainingParticipant{
		GroupTrainingID: trainingID,
		UserID:          userID,
	}
	return db.GORM.Create(participant).Error
}

// GetGroupTrainingParticipants возвращает список участников тренировки
func (db *DB) GetGroupTrainingParticipants(trainingID int64) ([]*models.User, error) {
	var users []*models.User
	err := db.GORM.
		Joins("JOIN group_training_participants ON users.id = group_training_participants.user_id").
		Where("group_training_participants.group_training_id = ?", trainingID).
		Find(&users).Error
	return users, err
}

// GetParticipantCount возвращает количество участников тренировки
func (db *DB) GetParticipantCount(trainingID int64) (int, error) {
	var count int64
	err := db.GORM.
		Model(&models.GroupTrainingParticipant{}).
		Where("group_training_id = ?", trainingID).
		Count(&count).Error
	return int(count), err
}
