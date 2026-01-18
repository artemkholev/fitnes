package models

import "time"

// User - базовая информация о пользователе Telegram
type User struct {
	ID         int64     `json:"id"`
	TelegramID int64     `json:"telegram_id"`
	Username   string    `json:"username"`
	FullName   string    `json:"full_name"`
	CreatedAt  time.Time `json:"created_at"`
}

// Organization - фитнес организация
type Organization struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
}

// OrganizationManager - менеджер организации (ответственное лицо)
type OrganizationManager struct {
	ID             int64      `json:"id"`
	OrganizationID int64      `json:"organization_id"`
	Username       string     `json:"username"` // без @
	TelegramID     *int64     `json:"telegram_id"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	DeactivatedAt  *time.Time `json:"deactivated_at"`
}

// OrganizationTrainer - тренер организации
type OrganizationTrainer struct {
	ID             int64      `json:"id"`
	OrganizationID int64      `json:"organization_id"`
	Username       string     `json:"username"`
	TelegramID     *int64     `json:"telegram_id"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	DeactivatedAt  *time.Time `json:"deactivated_at"`
}

// TrainerClient - связь тренер-клиент (доступ клиента)
type TrainerClient struct {
	ID            int64      `json:"id"`
	TrainerID     int64      `json:"trainer_id"` // ссылка на organization_trainers.id
	Username      string     `json:"username"`
	TelegramID    *int64     `json:"telegram_id"`
	IsActive      bool       `json:"is_active"`
	CreatedAt     time.Time  `json:"created_at"`
	DeactivatedAt *time.Time `json:"deactivated_at"`
}

// MuscleGroup - группа мышц
type MuscleGroup string

const (
	MuscleChest     MuscleGroup = "Грудь"
	MuscleBack      MuscleGroup = "Спина"
	MuscleLegs      MuscleGroup = "Ноги"
	MuscleShoulders MuscleGroup = "Плечи"
	MuscleBiceps    MuscleGroup = "Бицепс"
	MuscleTriceps   MuscleGroup = "Трицепс"
	MuscleAbs       MuscleGroup = "Пресс"
	MuscleCardio    MuscleGroup = "Кардио"
)

// Workout - тренировка
type Workout struct {
	ID               int64       `json:"id"`
	TrainerClientID  *int64      `json:"trainer_client_id"`
	ClientTelegramID int64       `json:"client_telegram_id"`
	Date             time.Time   `json:"date"`
	Notes            string      `json:"notes"`
	MuscleGroup      MuscleGroup `json:"muscle_group"`
	CreatedAt        time.Time   `json:"created_at"`
}

// Exercise - упражнение в тренировке
type Exercise struct {
	ID          int64     `json:"id"`
	WorkoutID   int64     `json:"workout_id"`
	Name        string    `json:"name"`
	Sets        int       `json:"sets"`
	Reps        int       `json:"reps"`
	Weight      float64   `json:"weight"`
	RestSeconds int       `json:"rest_seconds"`
	PhotoFileID string    `json:"photo_file_id"`
	Notes       string    `json:"notes"`
	Order       int       `json:"order"`
	CreatedAt   time.Time `json:"created_at"`
}

// GroupTraining - групповая тренировка
type GroupTraining struct {
	ID              int64     `json:"id"`
	OrganizationID  int64     `json:"organization_id"`
	TrainerID       int64     `json:"trainer_id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	ScheduledAt     time.Time `json:"scheduled_at"`
	MaxParticipants int       `json:"max_participants"`
	CreatedAt       time.Time `json:"created_at"`
}

// GroupTrainingParticipant - участник групповой тренировки
type GroupTrainingParticipant struct {
	ID              int64     `json:"id"`
	GroupTrainingID int64     `json:"group_training_id"`
	UserID          int64     `json:"user_id"`
	JoinedAt        time.Time `json:"joined_at"`
}

// UserRole - роль пользователя (определяется динамически)
type UserRole string

const (
	RoleAdmin   UserRole = "admin"
	RoleManager UserRole = "manager"
	RoleTrainer UserRole = "trainer"
	RoleClient  UserRole = "client"
	RoleGuest   UserRole = "guest"
)

// UserState - состояние пользователя в диалоге
type UserState struct {
	TelegramID int64
	State      string
	Data       map[string]interface{}
}

// AccessInfo - информация о доступах пользователя
type AccessInfo struct {
	IsAdmin        bool
	ManagerOrgs    []*ManagerOrgInfo    // организации где пользователь менеджер
	TrainerOrgs    []*TrainerOrgInfo    // организации где пользователь тренер
	ClientAccess   []*ClientAccessInfo  // активные доступы как клиент
	ArchivedAccess []*ClientAccessInfo  // архивные доступы (для просмотра истории)
}

// ManagerOrgInfo - информация о менеджере в организации
type ManagerOrgInfo struct {
	ManagerID    int64
	Organization *Organization
	IsActive     bool
}

// TrainerOrgInfo - информация о тренере в организации
type TrainerOrgInfo struct {
	TrainerID    int64
	Organization *Organization
	IsActive     bool
}

// ClientAccessInfo - информация о доступе клиента к тренеру
type ClientAccessInfo struct {
	TrainerClientID  int64
	OrganizationID   int64
	OrganizationName string
	TrainerID        int64
	TrainerUsername  string
	IsActive         bool
}

// TrainerWithOrg - тренер с информацией об организации
type TrainerWithOrg struct {
	Trainer      *OrganizationTrainer
	Organization *Organization
}

// ClientWithInfo - клиент с дополнительной информацией
type ClientWithInfo struct {
	Client       *TrainerClient
	FullName     string
	WorkoutCount int
	LastWorkout  *time.Time
}
