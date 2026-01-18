package database

import (
	"context"
	"fitness-bot/internal/models"
	"strings"
	"time"
)

// NormalizeUsername убирает @ из начала username
func NormalizeUsername(username string) string {
	return strings.TrimPrefix(strings.TrimSpace(username), "@")
}

// GetUserAccessInfo возвращает полную информацию о доступах пользователя
func (db *DB) GetUserAccessInfo(ctx context.Context, telegramID int64, username string) (*models.AccessInfo, error) {
	username = NormalizeUsername(username)
	info := &models.AccessInfo{}

	// Проверяем менеджерские доступы
	managerOrgs, err := db.GetManagerOrganizations(ctx, telegramID, username)
	if err != nil {
		return nil, err
	}
	info.ManagerOrgs = managerOrgs

	// Проверяем тренерские доступы
	trainerOrgs, err := db.GetTrainerOrganizations(ctx, telegramID, username)
	if err != nil {
		return nil, err
	}
	info.TrainerOrgs = trainerOrgs

	// Проверяем клиентские доступы (активные)
	clientAccess, err := db.GetClientAccess(ctx, telegramID, username, true)
	if err != nil {
		return nil, err
	}
	info.ClientAccess = clientAccess

	// Проверяем архивные клиентские доступы
	archivedAccess, err := db.GetClientAccess(ctx, telegramID, username, false)
	if err != nil {
		return nil, err
	}
	info.ArchivedAccess = archivedAccess

	return info, nil
}

// === МЕНЕДЖЕРЫ ===

// AddManager добавляет менеджера в организацию
func (db *DB) AddManager(ctx context.Context, orgID int64, username string) error {
	username = NormalizeUsername(username)
	query := `
		INSERT INTO organization_managers (organization_id, username, is_active)
		VALUES ($1, $2, true)
		ON CONFLICT (organization_id, username)
		DO UPDATE SET is_active = true, deactivated_at = NULL
	`
	_, err := db.Pool.Exec(ctx, query, orgID, username)
	return err
}

// RemoveManager деактивирует менеджера
func (db *DB) RemoveManager(ctx context.Context, orgID int64, username string) error {
	username = NormalizeUsername(username)
	query := `
		UPDATE organization_managers
		SET is_active = false, deactivated_at = $1
		WHERE organization_id = $2 AND username = $3
	`
	_, err := db.Pool.Exec(ctx, query, time.Now(), orgID, username)
	return err
}

// GetManagerOrganizations возвращает организации где пользователь менеджер
func (db *DB) GetManagerOrganizations(ctx context.Context, telegramID int64, username string) ([]*models.ManagerOrgInfo, error) {
	username = NormalizeUsername(username)
	query := `
		SELECT om.id, om.is_active, o.id, o.name, o.code, o.created_at
		FROM organization_managers om
		JOIN organizations o ON om.organization_id = o.id
		WHERE (om.telegram_id = $1 OR om.username = $2)
	`
	rows, err := db.Pool.Query(ctx, query, telegramID, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*models.ManagerOrgInfo
	for rows.Next() {
		info := &models.ManagerOrgInfo{Organization: &models.Organization{}}
		if err := rows.Scan(
			&info.ManagerID, &info.IsActive,
			&info.Organization.ID, &info.Organization.Name,
			&info.Organization.Code, &info.Organization.CreatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, info)
	}
	return result, rows.Err()
}

// GetOrganizationManagers возвращает всех менеджеров организации
func (db *DB) GetOrganizationManagers(ctx context.Context, orgID int64) ([]*models.OrganizationManager, error) {
	query := `
		SELECT id, organization_id, username, telegram_id, is_active, created_at, deactivated_at
		FROM organization_managers
		WHERE organization_id = $1
		ORDER BY is_active DESC, created_at
	`
	rows, err := db.Pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var managers []*models.OrganizationManager
	for rows.Next() {
		m := &models.OrganizationManager{}
		if err := rows.Scan(&m.ID, &m.OrganizationID, &m.Username, &m.TelegramID,
			&m.IsActive, &m.CreatedAt, &m.DeactivatedAt); err != nil {
			return nil, err
		}
		managers = append(managers, m)
	}
	return managers, rows.Err()
}

// === ТРЕНЕРЫ ===

// AddTrainer добавляет тренера в организацию
func (db *DB) AddTrainer(ctx context.Context, orgID int64, username string) error {
	username = NormalizeUsername(username)
	query := `
		INSERT INTO organization_trainers (organization_id, username, is_active)
		VALUES ($1, $2, true)
		ON CONFLICT (organization_id, username)
		DO UPDATE SET is_active = true, deactivated_at = NULL
	`
	_, err := db.Pool.Exec(ctx, query, orgID, username)
	return err
}

// RemoveTrainer деактивирует тренера
func (db *DB) RemoveTrainer(ctx context.Context, orgID int64, username string) error {
	username = NormalizeUsername(username)
	query := `
		UPDATE organization_trainers
		SET is_active = false, deactivated_at = $1
		WHERE organization_id = $2 AND username = $3
	`
	_, err := db.Pool.Exec(ctx, query, time.Now(), orgID, username)
	return err
}

// GetTrainerOrganizations возвращает организации где пользователь тренер
func (db *DB) GetTrainerOrganizations(ctx context.Context, telegramID int64, username string) ([]*models.TrainerOrgInfo, error) {
	username = NormalizeUsername(username)
	query := `
		SELECT ot.id, ot.is_active, o.id, o.name, o.code, o.created_at
		FROM organization_trainers ot
		JOIN organizations o ON ot.organization_id = o.id
		WHERE (ot.telegram_id = $1 OR ot.username = $2)
	`
	rows, err := db.Pool.Query(ctx, query, telegramID, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*models.TrainerOrgInfo
	for rows.Next() {
		info := &models.TrainerOrgInfo{Organization: &models.Organization{}}
		if err := rows.Scan(
			&info.TrainerID, &info.IsActive,
			&info.Organization.ID, &info.Organization.Name,
			&info.Organization.Code, &info.Organization.CreatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, info)
	}
	return result, rows.Err()
}

// GetOrganizationTrainers возвращает всех тренеров организации
func (db *DB) GetOrganizationTrainers(ctx context.Context, orgID int64) ([]*models.OrganizationTrainer, error) {
	query := `
		SELECT id, organization_id, username, telegram_id, is_active, created_at, deactivated_at
		FROM organization_trainers
		WHERE organization_id = $1
		ORDER BY is_active DESC, created_at
	`
	rows, err := db.Pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trainers []*models.OrganizationTrainer
	for rows.Next() {
		t := &models.OrganizationTrainer{}
		if err := rows.Scan(&t.ID, &t.OrganizationID, &t.Username, &t.TelegramID,
			&t.IsActive, &t.CreatedAt, &t.DeactivatedAt); err != nil {
			return nil, err
		}
		trainers = append(trainers, t)
	}
	return trainers, rows.Err()
}

// GetTrainerByID возвращает тренера по ID
func (db *DB) GetTrainerByID(ctx context.Context, trainerID int64) (*models.OrganizationTrainer, error) {
	query := `
		SELECT id, organization_id, username, telegram_id, is_active, created_at, deactivated_at
		FROM organization_trainers
		WHERE id = $1
	`
	t := &models.OrganizationTrainer{}
	err := db.Pool.QueryRow(ctx, query, trainerID).Scan(
		&t.ID, &t.OrganizationID, &t.Username, &t.TelegramID,
		&t.IsActive, &t.CreatedAt, &t.DeactivatedAt,
	)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// === КЛИЕНТЫ ===

// AddClient добавляет клиента к тренеру
func (db *DB) AddClient(ctx context.Context, trainerID int64, username string) error {
	username = NormalizeUsername(username)
	query := `
		INSERT INTO trainer_clients (trainer_id, username, is_active)
		VALUES ($1, $2, true)
		ON CONFLICT (trainer_id, username)
		DO UPDATE SET is_active = true, deactivated_at = NULL
	`
	_, err := db.Pool.Exec(ctx, query, trainerID, username)
	return err
}

// RemoveClient деактивирует клиента
func (db *DB) RemoveClient(ctx context.Context, trainerID int64, username string) error {
	username = NormalizeUsername(username)
	query := `
		UPDATE trainer_clients
		SET is_active = false, deactivated_at = $1
		WHERE trainer_id = $2 AND username = $3
	`
	_, err := db.Pool.Exec(ctx, query, time.Now(), trainerID, username)
	return err
}

// GetClientAccess возвращает доступы клиента к тренерам
func (db *DB) GetClientAccess(ctx context.Context, telegramID int64, username string, activeOnly bool) ([]*models.ClientAccessInfo, error) {
	username = NormalizeUsername(username)
	query := `
		SELECT tc.id, o.id, o.name, ot.id, ot.username, tc.is_active
		FROM trainer_clients tc
		JOIN organization_trainers ot ON tc.trainer_id = ot.id
		JOIN organizations o ON ot.organization_id = o.id
		WHERE (tc.telegram_id = $1 OR tc.username = $2)
	`
	if activeOnly {
		query += " AND tc.is_active = true AND ot.is_active = true"
	} else {
		query += " AND tc.is_active = false"
	}

	rows, err := db.Pool.Query(ctx, query, telegramID, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*models.ClientAccessInfo
	for rows.Next() {
		info := &models.ClientAccessInfo{}
		if err := rows.Scan(
			&info.TrainerClientID, &info.OrganizationID, &info.OrganizationName,
			&info.TrainerID, &info.TrainerUsername, &info.IsActive,
		); err != nil {
			return nil, err
		}
		result = append(result, info)
	}
	return result, rows.Err()
}

// GetTrainerClients возвращает всех клиентов тренера
func (db *DB) GetTrainerClients(ctx context.Context, trainerID int64) ([]*models.ClientWithInfo, error) {
	query := `
		SELECT tc.id, tc.trainer_id, tc.username, tc.telegram_id, tc.is_active,
		       tc.created_at, tc.deactivated_at,
		       COALESCE(u.full_name, ''),
		       COUNT(w.id) as workout_count,
		       MAX(w.date) as last_workout
		FROM trainer_clients tc
		LEFT JOIN users u ON tc.telegram_id = u.telegram_id
		LEFT JOIN workouts w ON w.trainer_client_id = tc.id
		WHERE tc.trainer_id = $1
		GROUP BY tc.id, tc.trainer_id, tc.username, tc.telegram_id, tc.is_active,
		         tc.created_at, tc.deactivated_at, u.full_name
		ORDER BY tc.is_active DESC, tc.created_at DESC
	`
	rows, err := db.Pool.Query(ctx, query, trainerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*models.ClientWithInfo
	for rows.Next() {
		info := &models.ClientWithInfo{Client: &models.TrainerClient{}}
		if err := rows.Scan(
			&info.Client.ID, &info.Client.TrainerID, &info.Client.Username,
			&info.Client.TelegramID, &info.Client.IsActive,
			&info.Client.CreatedAt, &info.Client.DeactivatedAt,
			&info.FullName, &info.WorkoutCount, &info.LastWorkout,
		); err != nil {
			return nil, err
		}
		result = append(result, info)
	}
	return result, rows.Err()
}

// GetTrainerClientByID возвращает связь тренер-клиент по ID
func (db *DB) GetTrainerClientByID(ctx context.Context, id int64) (*models.TrainerClient, error) {
	query := `
		SELECT id, trainer_id, username, telegram_id, is_active, created_at, deactivated_at
		FROM trainer_clients
		WHERE id = $1
	`
	tc := &models.TrainerClient{}
	err := db.Pool.QueryRow(ctx, query, id).Scan(
		&tc.ID, &tc.TrainerID, &tc.Username, &tc.TelegramID,
		&tc.IsActive, &tc.CreatedAt, &tc.DeactivatedAt,
	)
	if err != nil {
		return nil, err
	}
	return tc, nil
}

// === СВЯЗЫВАНИЕ TELEGRAM ID ===

// LinkTelegramID связывает telegram_id с username во всех таблицах доступов
func (db *DB) LinkTelegramID(ctx context.Context, telegramID int64, username string) error {
	username = NormalizeUsername(username)

	// Обновляем менеджеров
	_, err := db.Pool.Exec(ctx,
		"UPDATE organization_managers SET telegram_id = $1 WHERE username = $2 AND telegram_id IS NULL",
		telegramID, username)
	if err != nil {
		return err
	}

	// Обновляем тренеров
	_, err = db.Pool.Exec(ctx,
		"UPDATE organization_trainers SET telegram_id = $1 WHERE username = $2 AND telegram_id IS NULL",
		telegramID, username)
	if err != nil {
		return err
	}

	// Обновляем клиентов
	_, err = db.Pool.Exec(ctx,
		"UPDATE trainer_clients SET telegram_id = $1 WHERE username = $2 AND telegram_id IS NULL",
		telegramID, username)

	return err
}
