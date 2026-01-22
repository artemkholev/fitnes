-- 000001_init.up.sql
-- Начальная схема базы данных

-- Организации
CREATE TABLE IF NOT EXISTS organizations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Пользователи (базовая информация)
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    telegram_id BIGINT UNIQUE NOT NULL,
    username VARCHAR(255),
    full_name VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Менеджеры организаций
CREATE TABLE IF NOT EXISTS organization_managers (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    username VARCHAR(255) NOT NULL,
    telegram_id BIGINT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deactivated_at TIMESTAMP,
    UNIQUE(organization_id, username)
);

-- Тренеры организаций
CREATE TABLE IF NOT EXISTS organization_trainers (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    username VARCHAR(255) NOT NULL,
    telegram_id BIGINT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deactivated_at TIMESTAMP,
    UNIQUE(organization_id, username)
);

-- Связь тренер-клиент
CREATE TABLE IF NOT EXISTS trainer_clients (
    id SERIAL PRIMARY KEY,
    trainer_id INTEGER NOT NULL REFERENCES organization_trainers(id) ON DELETE CASCADE,
    username VARCHAR(255) NOT NULL,
    telegram_id BIGINT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deactivated_at TIMESTAMP,
    UNIQUE(trainer_id, username)
);

-- Тренировки
CREATE TABLE IF NOT EXISTS workouts (
    id SERIAL PRIMARY KEY,
    trainer_client_id INTEGER REFERENCES trainer_clients(id),
    client_telegram_id BIGINT NOT NULL,
    date TIMESTAMP NOT NULL,
    notes TEXT,
    muscle_group VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Упражнения
CREATE TABLE IF NOT EXISTS exercises (
    id SERIAL PRIMARY KEY,
    workout_id INTEGER NOT NULL REFERENCES workouts(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    sets INTEGER,
    reps INTEGER,
    weight DECIMAL(10, 2),
    rest_seconds INTEGER,
    photo_file_id VARCHAR(255),
    notes TEXT,
    "order" INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Групповые тренировки
CREATE TABLE IF NOT EXISTS group_trainings (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id),
    trainer_id INTEGER NOT NULL REFERENCES organization_trainers(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    scheduled_at TIMESTAMP NOT NULL,
    max_participants INTEGER DEFAULT 10,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Участники групповых тренировок
CREATE TABLE IF NOT EXISTS group_training_participants (
    id SERIAL PRIMARY KEY,
    group_training_id INTEGER NOT NULL REFERENCES group_trainings(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id),
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(group_training_id, user_id)
);

-- Индексы
CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON users(telegram_id);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_org_managers_username ON organization_managers(username);
CREATE INDEX IF NOT EXISTS idx_org_managers_telegram_id ON organization_managers(telegram_id);
CREATE INDEX IF NOT EXISTS idx_org_trainers_username ON organization_trainers(username);
CREATE INDEX IF NOT EXISTS idx_org_trainers_telegram_id ON organization_trainers(telegram_id);
CREATE INDEX IF NOT EXISTS idx_trainer_clients_username ON trainer_clients(username);
CREATE INDEX IF NOT EXISTS idx_trainer_clients_telegram_id ON trainer_clients(telegram_id);
CREATE INDEX IF NOT EXISTS idx_workouts_trainer_client_id ON workouts(trainer_client_id);
CREATE INDEX IF NOT EXISTS idx_workouts_client_telegram_id ON workouts(client_telegram_id);
CREATE INDEX IF NOT EXISTS idx_workouts_date ON workouts(date);
CREATE INDEX IF NOT EXISTS idx_exercises_workout_id ON exercises(workout_id);
CREATE INDEX IF NOT EXISTS idx_group_trainings_scheduled_at ON group_trainings(scheduled_at);
CREATE INDEX IF NOT EXISTS idx_group_trainings_organization_id ON group_trainings(organization_id);
