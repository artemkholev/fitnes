-- Исправление таблицы workouts для новой системы доступов

-- Удаляем старые обязательные поля если они есть
ALTER TABLE workouts DROP COLUMN IF EXISTS user_id;
ALTER TABLE workouts DROP COLUMN IF EXISTS trainer_id;

-- Убеждаемся что новые поля существуют
ALTER TABLE workouts ADD COLUMN IF NOT EXISTS trainer_client_id INTEGER REFERENCES trainer_clients(id);
ALTER TABLE workouts ADD COLUMN IF NOT EXISTS client_telegram_id BIGINT;

-- Делаем client_telegram_id обязательным для новых записей
-- (не применяем NOT NULL чтобы не сломать старые данные)

-- Индексы
CREATE INDEX IF NOT EXISTS idx_workouts_client_telegram_id ON workouts(client_telegram_id);
