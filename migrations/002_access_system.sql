-- Новая система доступов

-- Обновляем таблицу users - убираем role, добавляем поля
ALTER TABLE users DROP COLUMN IF EXISTS role;
ALTER TABLE users DROP COLUMN IF EXISTS trainer_id;
ALTER TABLE users DROP COLUMN IF EXISTS organization_id;

-- Таблица менеджеров организаций
CREATE TABLE IF NOT EXISTS organization_managers (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    username VARCHAR(255) NOT NULL, -- @username без @
    telegram_id BIGINT, -- заполняется когда пользователь напишет боту
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deactivated_at TIMESTAMP,
    UNIQUE(organization_id, username)
);

-- Таблица тренеров организаций
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

-- Таблица связи тренер-клиент (доступы клиентов)
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

-- Обновляем таблицу workouts для новой структуры
ALTER TABLE workouts DROP CONSTRAINT IF EXISTS workouts_user_id_fkey;
ALTER TABLE workouts DROP CONSTRAINT IF EXISTS workouts_trainer_id_fkey;
ALTER TABLE workouts ADD COLUMN IF NOT EXISTS trainer_client_id INTEGER REFERENCES trainer_clients(id);
ALTER TABLE workouts ADD COLUMN IF NOT EXISTS client_telegram_id BIGINT;

-- Индексы
CREATE INDEX IF NOT EXISTS idx_org_managers_username ON organization_managers(username);
CREATE INDEX IF NOT EXISTS idx_org_managers_telegram_id ON organization_managers(telegram_id);
CREATE INDEX IF NOT EXISTS idx_org_trainers_username ON organization_trainers(username);
CREATE INDEX IF NOT EXISTS idx_org_trainers_telegram_id ON organization_trainers(telegram_id);
CREATE INDEX IF NOT EXISTS idx_trainer_clients_username ON trainer_clients(username);
CREATE INDEX IF NOT EXISTS idx_trainer_clients_telegram_id ON trainer_clients(telegram_id);
CREATE INDEX IF NOT EXISTS idx_workouts_trainer_client_id ON workouts(trainer_client_id);
