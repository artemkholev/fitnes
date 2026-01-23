package models

import (
	"time"

	"gorm.io/gorm"
)

// User - базовая информация о пользователе Telegram
type User struct {
	ID         int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	TelegramID int64          `gorm:"uniqueIndex;not null" json:"telegram_id"`
	Username   string         `gorm:"uniqueIndex;type:varchar(255)" json:"username"`
	FullName   string         `gorm:"type:varchar(255)" json:"full_name"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"-"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string {
	return "users"
}

// Organization - фитнес организация
type Organization struct {
	ID        int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string         `gorm:"type:varchar(255);not null" json:"name"`
	Code      string         `gorm:"uniqueIndex;type:varchar(50);not null" json:"code"`
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Organization) TableName() string {
	return "organizations"
}

// OrganizationManager - менеджер организации (ответственное лицо)
type OrganizationManager struct {
	ID             int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	OrganizationID int64          `gorm:"not null;index" json:"organization_id"`
	Username       string         `gorm:"not null;index;type:varchar(255)" json:"username"` // без @
	TelegramID     *int64         `gorm:"index" json:"telegram_id"`
	IsActive       bool           `gorm:"default:true" json:"is_active"`
	CreatedAt      time.Time      `gorm:"autoCreateTime" json:"created_at"`
	DeactivatedAt  *time.Time     `json:"deactivated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Organization Organization `gorm:"foreignKey:OrganizationID" json:"-"`
}

func (OrganizationManager) TableName() string {
	return "organization_managers"
}

// OrganizationTrainer - тренер организации
type OrganizationTrainer struct {
	ID             int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	OrganizationID int64          `gorm:"not null;index" json:"organization_id"`
	Username       string         `gorm:"not null;index;type:varchar(255)" json:"username"`
	TelegramID     *int64         `gorm:"index" json:"telegram_id"`
	IsActive       bool           `gorm:"default:true" json:"is_active"`
	CreatedAt      time.Time      `gorm:"autoCreateTime" json:"created_at"`
	DeactivatedAt  *time.Time     `json:"deactivated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Organization Organization    `gorm:"foreignKey:OrganizationID" json:"-"`
	Clients      []TrainerClient `gorm:"foreignKey:TrainerID" json:"-"`
}

func (OrganizationTrainer) TableName() string {
	return "organization_trainers"
}

// TrainerClient - связь тренер-клиент (доступ клиента)
type TrainerClient struct {
	ID            int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	TrainerID     int64          `gorm:"not null;index" json:"trainer_id"` // ссылка на organization_trainers.id
	Username      string         `gorm:"not null;index;type:varchar(255)" json:"username"`
	TelegramID    *int64         `gorm:"index" json:"telegram_id"`
	IsActive      bool           `gorm:"default:true" json:"is_active"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	DeactivatedAt *time.Time     `json:"deactivated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Trainer  OrganizationTrainer `gorm:"foreignKey:TrainerID" json:"-"`
	Workouts []Workout           `gorm:"foreignKey:TrainerClientID" json:"-"`
}

func (TrainerClient) TableName() string {
	return "trainer_clients"
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
	ID               int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	TrainerClientID  *int64         `gorm:"index" json:"trainer_client_id"`
	ClientTelegramID int64          `gorm:"not null;index" json:"client_telegram_id"`
	Date             time.Time      `gorm:"not null;index" json:"date"`
	Notes            string         `gorm:"type:text" json:"notes"`
	MuscleGroup      MuscleGroup    `gorm:"type:varchar(50);not null" json:"muscle_group"`
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime" json:"-"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	TrainerClient *TrainerClient `gorm:"foreignKey:TrainerClientID" json:"-"`
	Exercises     []Exercise     `gorm:"foreignKey:WorkoutID" json:"-"`
}

func (Workout) TableName() string {
	return "workouts"
}

// Exercise - упражнение в тренировке
type Exercise struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	WorkoutID   int64     `gorm:"not null;index" json:"workout_id"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Sets        int       `gorm:"not null" json:"sets"`
	Reps        int       `gorm:"not null" json:"reps"`
	Weight      float64   `gorm:"type:decimal(10,2)" json:"weight"`
	RestSeconds int       `json:"rest_seconds"`
	PhotoFileID string    `gorm:"type:varchar(255)" json:"photo_file_id"`
	Notes       string    `gorm:"type:text" json:"notes"`
	Order       int       `gorm:"not null" json:"order"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	Workout Workout `gorm:"foreignKey:WorkoutID" json:"-"`
}

func (Exercise) TableName() string {
	return "exercises"
}

// GroupTraining - групповая тренировка
type GroupTraining struct {
	ID              int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	OrganizationID  int64          `gorm:"not null;index" json:"organization_id"`
	TrainerID       int64          `gorm:"not null;index" json:"trainer_id"`
	Name            string         `gorm:"type:varchar(255);not null" json:"name"`
	Description     string         `gorm:"type:text" json:"description"`
	ScheduledAt     time.Time      `gorm:"not null;index" json:"scheduled_at"`
	MaxParticipants int            `gorm:"not null" json:"max_participants"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"-"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Organization Organization                 `gorm:"foreignKey:OrganizationID" json:"-"`
	Trainer      OrganizationTrainer          `gorm:"foreignKey:TrainerID" json:"-"`
	Participants []GroupTrainingParticipant   `gorm:"foreignKey:GroupTrainingID" json:"-"`
}

func (GroupTraining) TableName() string {
	return "group_trainings"
}

// GroupTrainingParticipant - участник групповой тренировки
type GroupTrainingParticipant struct {
	ID              int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	GroupTrainingID int64          `gorm:"not null;index" json:"group_training_id"`
	UserID          int64          `gorm:"not null;index" json:"user_id"`
	JoinedAt        time.Time      `gorm:"autoCreateTime" json:"joined_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	GroupTraining GroupTraining `gorm:"foreignKey:GroupTrainingID" json:"-"`
	User          User          `gorm:"foreignKey:UserID" json:"-"`
}

func (GroupTrainingParticipant) TableName() string {
	return "group_training_participants"
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
