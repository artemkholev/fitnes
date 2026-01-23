package database

import (
	"fitness-bot/internal/models"
	"strings"
	"time"

	"gorm.io/gorm/clause"
)

// NormalizeUsername убирает @ из начала username
func NormalizeUsername(username string) string {
	return strings.TrimPrefix(strings.TrimSpace(username), "@")
}

// GetUserAccessInfo возвращает полную информацию о доступах пользователя
func (db *DB) GetUserAccessInfo(telegramID int64, username string) (*models.AccessInfo, error) {
	username = NormalizeUsername(username)
	info := &models.AccessInfo{}

	// Проверяем менеджерские доступы
	managerOrgs, err := db.GetManagerOrganizations(telegramID, username)
	if err != nil {
		return nil, err
	}
	info.ManagerOrgs = managerOrgs

	// Проверяем тренерские доступы
	trainerOrgs, err := db.GetTrainerOrganizations(telegramID, username)
	if err != nil {
		return nil, err
	}
	info.TrainerOrgs = trainerOrgs

	// Проверяем клиентские доступы (активные)
	clientAccess, err := db.GetClientAccess(telegramID, username, true)
	if err != nil {
		return nil, err
	}
	info.ClientAccess = clientAccess

	// Проверяем архивные клиентские доступы
	archivedAccess, err := db.GetClientAccess(telegramID, username, false)
	if err != nil {
		return nil, err
	}
	info.ArchivedAccess = archivedAccess

	return info, nil
}

// === МЕНЕДЖЕРЫ ===

// AddManager добавляет менеджера в организацию
func (db *DB) AddManager(orgID int64, username string) error {
	username = NormalizeUsername(username)
	manager := models.OrganizationManager{
		OrganizationID: orgID,
		Username:       username,
		IsActive:       true,
	}
	// Upsert: если существует - обновляем is_active и deactivated_at
	return db.GORM.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "organization_id"}, {Name: "username"}},
		DoUpdates: clause.Assignments(map[string]interface{}{"is_active": true, "deactivated_at": nil}),
	}).Create(&manager).Error
}

// RemoveManager деактивирует менеджера
func (db *DB) RemoveManager(orgID int64, username string) error {
	username = NormalizeUsername(username)
	now := time.Now()
	return db.GORM.Model(&models.OrganizationManager{}).
		Where("organization_id = ? AND username = ?", orgID, username).
		Updates(map[string]interface{}{"is_active": false, "deactivated_at": now}).Error
}

// GetManagerOrganizations возвращает организации где пользователь менеджер
func (db *DB) GetManagerOrganizations(telegramID int64, username string) ([]*models.ManagerOrgInfo, error) {
	username = NormalizeUsername(username)

	var managers []models.OrganizationManager
	err := db.GORM.
		Preload("Organization").
		Where("telegram_id = ? OR username = ?", telegramID, username).
		Find(&managers).Error
	if err != nil {
		return nil, err
	}

	var result []*models.ManagerOrgInfo
	for _, m := range managers {
		result = append(result, &models.ManagerOrgInfo{
			ManagerID:    m.ID,
			Organization: &m.Organization,
			IsActive:     m.IsActive,
		})
	}
	return result, nil
}

// GetOrganizationManagers возвращает всех менеджеров организации
func (db *DB) GetOrganizationManagers(orgID int64) ([]*models.OrganizationManager, error) {
	var managers []*models.OrganizationManager
	err := db.GORM.
		Where("organization_id = ?", orgID).
		Order("is_active DESC, created_at").
		Find(&managers).Error
	return managers, err
}

// === ТРЕНЕРЫ ===

// AddTrainer добавляет тренера в организацию
func (db *DB) AddTrainer(orgID int64, username string) error {
	username = NormalizeUsername(username)
	trainer := models.OrganizationTrainer{
		OrganizationID: orgID,
		Username:       username,
		IsActive:       true,
	}
	// Upsert: если существует - обновляем is_active и deactivated_at
	return db.GORM.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "organization_id"}, {Name: "username"}},
		DoUpdates: clause.Assignments(map[string]interface{}{"is_active": true, "deactivated_at": nil}),
	}).Create(&trainer).Error
}

// RemoveTrainer деактивирует тренера
func (db *DB) RemoveTrainer(orgID int64, username string) error {
	username = NormalizeUsername(username)
	now := time.Now()
	return db.GORM.Model(&models.OrganizationTrainer{}).
		Where("organization_id = ? AND username = ?", orgID, username).
		Updates(map[string]interface{}{"is_active": false, "deactivated_at": now}).Error
}

// GetTrainerOrganizations возвращает организации где пользователь тренер
func (db *DB) GetTrainerOrganizations(telegramID int64, username string) ([]*models.TrainerOrgInfo, error) {
	username = NormalizeUsername(username)

	var trainers []models.OrganizationTrainer
	err := db.GORM.
		Preload("Organization").
		Where("telegram_id = ? OR username = ?", telegramID, username).
		Find(&trainers).Error
	if err != nil {
		return nil, err
	}

	var result []*models.TrainerOrgInfo
	for _, t := range trainers {
		result = append(result, &models.TrainerOrgInfo{
			TrainerID:    t.ID,
			Organization: &t.Organization,
			IsActive:     t.IsActive,
		})
	}
	return result, nil
}

// GetOrganizationTrainers возвращает всех тренеров организации
func (db *DB) GetOrganizationTrainers(orgID int64) ([]*models.OrganizationTrainer, error) {
	var trainers []*models.OrganizationTrainer
	err := db.GORM.
		Where("organization_id = ?", orgID).
		Order("is_active DESC, created_at").
		Find(&trainers).Error
	return trainers, err
}

// GetTrainerByID возвращает тренера по ID
func (db *DB) GetTrainerByID(trainerID int64) (*models.OrganizationTrainer, error) {
	var trainer models.OrganizationTrainer
	err := db.GORM.First(&trainer, trainerID).Error
	if err != nil {
		return nil, err
	}
	return &trainer, nil
}

// === КЛИЕНТЫ ===

// AddClient добавляет клиента к тренеру
func (db *DB) AddClient(trainerID int64, username string) error {
	username = NormalizeUsername(username)
	client := models.TrainerClient{
		TrainerID: trainerID,
		Username:  username,
		IsActive:  true,
	}
	// Upsert: если существует - обновляем is_active и deactivated_at
	return db.GORM.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "trainer_id"}, {Name: "username"}},
		DoUpdates: clause.Assignments(map[string]interface{}{"is_active": true, "deactivated_at": nil}),
	}).Create(&client).Error
}

// RemoveClient деактивирует клиента
func (db *DB) RemoveClient(trainerID int64, username string) error {
	username = NormalizeUsername(username)
	now := time.Now()
	return db.GORM.Model(&models.TrainerClient{}).
		Where("trainer_id = ? AND username = ?", trainerID, username).
		Updates(map[string]interface{}{"is_active": false, "deactivated_at": now}).Error
}

// GetClientAccess возвращает доступы клиента к тренерам
func (db *DB) GetClientAccess(telegramID int64, username string, activeOnly bool) ([]*models.ClientAccessInfo, error) {
	username = NormalizeUsername(username)

	query := db.GORM.Table("trainer_clients tc").
		Select("tc.id as trainer_client_id, o.id as organization_id, o.name as organization_name, ot.id as trainer_id, ot.username as trainer_username, tc.is_active").
		Joins("JOIN organization_trainers ot ON tc.trainer_id = ot.id").
		Joins("JOIN organizations o ON ot.organization_id = o.id").
		Where("tc.telegram_id = ? OR tc.username = ?", telegramID, username)

	if activeOnly {
		query = query.Where("tc.is_active = ? AND ot.is_active = ?", true, true)
	} else {
		query = query.Where("tc.is_active = ?", false)
	}

	var result []*models.ClientAccessInfo
	err := query.Scan(&result).Error
	return result, err
}

// GetTrainerClients возвращает всех клиентов тренера
func (db *DB) GetTrainerClients(trainerID int64) ([]*models.ClientWithInfo, error) {
	type ClientQueryResult struct {
		models.TrainerClient
		FullName     string
		WorkoutCount int
		LastWorkout  *time.Time
	}

	var results []ClientQueryResult
	err := db.GORM.Table("trainer_clients tc").
		Select("tc.*, COALESCE(u.full_name, '') as full_name, COUNT(w.id) as workout_count, MAX(w.date) as last_workout").
		Joins("LEFT JOIN users u ON tc.telegram_id = u.telegram_id").
		Joins("LEFT JOIN workouts w ON w.trainer_client_id = tc.id").
		Where("tc.trainer_id = ?", trainerID).
		Group("tc.id, tc.trainer_id, tc.username, tc.telegram_id, tc.is_active, tc.created_at, tc.deactivated_at, u.full_name").
		Order("tc.is_active DESC, tc.created_at DESC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	var clientsInfo []*models.ClientWithInfo
	for _, r := range results {
		clientsInfo = append(clientsInfo, &models.ClientWithInfo{
			Client:       &r.TrainerClient,
			FullName:     r.FullName,
			WorkoutCount: r.WorkoutCount,
			LastWorkout:  r.LastWorkout,
		})
	}
	return clientsInfo, nil
}

// GetTrainerClientByID возвращает связь тренер-клиент по ID
func (db *DB) GetTrainerClientByID(id int64) (*models.TrainerClient, error) {
	var client models.TrainerClient
	err := db.GORM.First(&client, id).Error
	if err != nil {
		return nil, err
	}
	return &client, nil
}

// === СВЯЗЫВАНИЕ TELEGRAM ID ===

// LinkTelegramID связывает telegram_id с username во всех таблицах доступов
func (db *DB) LinkTelegramID(telegramID int64, username string) error {
	username = NormalizeUsername(username)

	// Обновляем менеджеров
	err := db.GORM.Model(&models.OrganizationManager{}).
		Where("username = ? AND telegram_id IS NULL", username).
		Update("telegram_id", telegramID).Error
	if err != nil {
		return err
	}

	// Обновляем тренеров
	err = db.GORM.Model(&models.OrganizationTrainer{}).
		Where("username = ? AND telegram_id IS NULL", username).
		Update("telegram_id", telegramID).Error
	if err != nil {
		return err
	}

	// Обновляем клиентов
	err = db.GORM.Model(&models.TrainerClient{}).
		Where("username = ? AND telegram_id IS NULL", username).
		Update("telegram_id", telegramID).Error

	return err
}
