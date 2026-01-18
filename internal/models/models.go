package models

import "time"

type UserRole string

const (
	RoleClient  UserRole = "client"
	RoleTrainer UserRole = "trainer"
)

type User struct {
	ID             int64     `json:"id"`
	TelegramID     int64     `json:"telegram_id"`
	Username       string    `json:"username"`
	FullName       string    `json:"full_name"`
	Role           UserRole  `json:"role"`
	OrganizationID *int64    `json:"organization_id"`
	TrainerID      *int64    `json:"trainer_id"`
	CreatedAt      time.Time `json:"created_at"`
}

type Organization struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
}

type MuscleGroup string

const (
	MuscleChest    MuscleGroup = "Грудь"
	MuscleBack     MuscleGroup = "Спина"
	MuscleLegs     MuscleGroup = "Ноги"
	MuscleShoulders MuscleGroup = "Плечи"
	MuscleBiceps   MuscleGroup = "Бицепс"
	MuscleTriceps  MuscleGroup = "Трицепс"
	MuscleAbs      MuscleGroup = "Пресс"
	MuscleCardio   MuscleGroup = "Кардио"
)

type Workout struct {
	ID         int64       `json:"id"`
	UserID     int64       `json:"user_id"`
	TrainerID  *int64      `json:"trainer_id"`
	Date       time.Time   `json:"date"`
	Notes      string      `json:"notes"`
	MuscleGroup MuscleGroup `json:"muscle_group"`
	CreatedAt  time.Time   `json:"created_at"`
}

type Exercise struct {
	ID           int64   `json:"id"`
	WorkoutID    int64   `json:"workout_id"`
	Name         string  `json:"name"`
	Sets         int     `json:"sets"`
	Reps         int     `json:"reps"`
	Weight       float64 `json:"weight"`
	RestSeconds  int     `json:"rest_seconds"`
	PhotoFileID  string  `json:"photo_file_id"`
	Notes        string  `json:"notes"`
	Order        int     `json:"order"`
	CreatedAt    time.Time `json:"created_at"`
}

type GroupTraining struct {
	ID             int64     `json:"id"`
	OrganizationID int64     `json:"organization_id"`
	TrainerID      int64     `json:"trainer_id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	ScheduledAt    time.Time `json:"scheduled_at"`
	MaxParticipants int      `json:"max_participants"`
	CreatedAt      time.Time `json:"created_at"`
}

type GroupTrainingParticipant struct {
	ID              int64     `json:"id"`
	GroupTrainingID int64     `json:"group_training_id"`
	UserID          int64     `json:"user_id"`
	JoinedAt        time.Time `json:"joined_at"`
}

type UserState struct {
	TelegramID int64
	State      string
	Data       map[string]interface{}
}
