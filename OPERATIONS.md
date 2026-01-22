# FitBot - Полное руководство по развёртыванию и обслуживанию

## Содержание

1. [Быстрый старт](#быстрый-старт)
2. [Локальная разработка](#локальная-разработка)
3. [Деплой на сервер](#деплой-на-сервер)
4. [Миграции базы данных](#миграции-базы-данных)
5. [Обновление бота](#обновление-бота)
6. [Мониторинг и логи](#мониторинг-и-логи)
7. [Бэкапы](#бэкапы)
8. [Решение проблем](#решение-проблем)

---

## Быстрый старт

### Локально (для разработки)

```bash
# 1. Клонировать проект
git clone <repo-url>
cd fitness-bot

# 2. Создать конфигурацию
cp .env.example .env
nano .env  # Добавить TELEGRAM_BOT_TOKEN

# 3. Запустить
docker-compose up -d

# 4. Смотреть логи
docker-compose logs -f bot
```

### На сервере (production)

```bash
# 1. Создать директорию
mkdir -p /opt/fitness-bot
cd /opt/fitness-bot

# 2. Скопировать файлы
scp -r * user@server:/opt/fitness-bot/

# 3. На сервере настроить .env
nano .env

# 4. Запустить
docker-compose up -d --build
```

---

## Локальная разработка

### Требования

- Go 1.21+
- Docker & Docker Compose
- Telegram Bot Token от [@BotFather](https://t.me/BotFather)

### Шаг 1: Подготовка окружения

```bash
# Клонирование проекта
git clone <repo-url>
cd fitness-bot

# Установка зависимостей Go
go mod download
go mod tidy
```

### Шаг 2: Создание Telegram бота

1. Откройте [@BotFather](https://t.me/BotFather) в Telegram
2. Отправьте `/newbot`
3. Укажите имя бота (например: My Fitness Bot)
4. Укажите username (например: myfitness_bot)
5. Скопируйте токен

### Шаг 3: Конфигурация .env

Создайте `.env` файл:

```bash
cp .env.example .env
```

Заполните переменные:

```env
# Telegram Bot Token от @BotFather
TELEGRAM_BOT_TOKEN=1234567890:ABCdefGHIjklMNOpqrsTUVwxyz

# Admin username (БЕЗ @)
ADMIN_USERNAME=your_telegram_username

# Database (для Docker)
DB_HOST=postgres
DB_PORT=5432
DB_USER=fitness_user
DB_PASSWORD=secure_password_123
DB_NAME=fitness_bot
```

> ⚠️ **Важно:** Никогда не коммитьте `.env` в Git!

### Шаг 4: Запуск через Docker

```bash
# Запуск всех сервисов
docker-compose up -d

# Проверка статуса
docker-compose ps

# Просмотр логов
docker-compose logs -f bot

# Остановка
docker-compose down
```

### Шаг 5: Запуск без Docker (опционально)

Если нужно запустить бота локально без Docker:

```bash
# 1. Запустите PostgreSQL отдельно
docker run -d \
  --name fitness_postgres \
  -e POSTGRES_USER=fitness_user \
  -e POSTGRES_PASSWORD=secure_password_123 \
  -e POSTGRES_DB=fitness_bot \
  -p 5432:5432 \
  postgres:15

# 2. Обновите .env
DB_HOST=localhost

# 3. Запустите бота
go run cmd/bot/main.go
```

### Полезные команды для разработки

```bash
# Форматирование кода
go fmt ./...

# Проверка на ошибки
go vet ./...

# Сборка бинарника
go build -o bin/bot cmd/bot/main.go

# Запуск тестов (если есть)
go test ./...

# Пересборка Docker образа
docker-compose build --no-cache bot

# Перезапуск только бота
docker-compose restart bot

# Просмотр логов с фильтром
docker-compose logs -f bot | grep ERROR
```

---

## Деплой на сервер

### Требования к серверу

**Минимальные:**
- Ubuntu 20.04+ / Debian 11+
- 1 CPU, 1GB RAM
- 10GB свободного места
- Доступ по SSH

**Рекомендуемые:**
- 2 CPU, 2GB RAM
- 20GB SSD
- Статический IP-адрес

### Шаг 1: Подготовка сервера

```bash
# Подключение к серверу
ssh user@your-server-ip

# Обновление системы
sudo apt update && sudo apt upgrade -y

# Установка Docker
curl -fsSL https://get.docker.com | sudo sh

# Установка Docker Compose
sudo apt install docker-compose-plugin -y

# Добавление пользователя в группу docker
sudo usermod -aG docker $USER

# Выход и повторный вход (для применения группы)
exit
ssh user@your-server-ip

# Проверка установки
docker --version
docker compose version
```

### Шаг 2: Копирование проекта на сервер

**Вариант 1: Через Git (рекомендуется)**

```bash
# На сервере
cd /opt
sudo mkdir fitness-bot
sudo chown $USER:$USER fitness-bot
cd fitness-bot

# Клонирование
git clone https://github.com/your-repo/fitness-bot.git .
```

**Вариант 2: Через SCP**

```bash
# На локальной машине
cd /path/to/fitness-bot
scp -r * user@your-server-ip:/opt/fitness-bot/
```

**Вариант 3: Через rsync (исключает ненужное)**

```bash
# На локальной машине
rsync -avz --exclude='.git' --exclude='.env' \
  . user@your-server-ip:/opt/fitness-bot/
```

### Шаг 3: Конфигурация на сервере

```bash
# Переход в директорию проекта
cd /opt/fitness-bot

# Создание .env
nano .env
```

Пример production конфигурации:

```env
# Telegram
TELEGRAM_BOT_TOKEN=ваш_токен_от_BotFather
ADMIN_USERNAME=ваш_username_без_@

# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=fitness_user
DB_PASSWORD=СИЛЬНЫЙ_ПАРОЛЬ_ДЛЯ_PRODUCTION
DB_NAME=fitness_bot
```

> 🔒 **Безопасность:** Используйте сильный пароль для БД (минимум 20 символов, буквы+цифры+символы)

```bash
# Установка правильных прав доступа
chmod 600 .env
```

### Шаг 4: Запуск на сервере

```bash
# Запуск
docker compose up -d --build

# Проверка статуса
docker compose ps

# Просмотр логов
docker compose logs -f bot

# Ожидаемый вывод:
# ✓ Database connected successfully
# ✓ Migrations: applied successfully
# ✓ Bot started successfully!
```

### Шаг 5: Настройка автозапуска

Создайте systemd сервис для автоматического запуска при перезагрузке:

```bash
sudo nano /etc/systemd/system/fitness-bot.service
```

Содержимое файла:

```ini
[Unit]
Description=FitBot Telegram Bot
After=docker.service
Requires=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=/opt/fitness-bot
ExecStart=/usr/bin/docker compose up -d
ExecStop=/usr/bin/docker compose down
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
```

Активация сервиса:

```bash
# Перезагрузка systemd
sudo systemctl daemon-reload

# Включение автозапуска
sudo systemctl enable fitness-bot

# Запуск сервиса
sudo systemctl start fitness-bot

# Проверка статуса
sudo systemctl status fitness-bot

# Тестирование перезагрузки (опционально)
sudo reboot
# После перезагрузки бот должен запуститься автоматически
```

### Шаг 6: Настройка Firewall (рекомендуется)

```bash
# Установка UFW
sudo apt install ufw -y

# Разрешить SSH
sudo ufw allow 22/tcp

# Разрешить HTTP/HTTPS (если планируете веб-панель)
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Включить firewall
sudo ufw enable

# Проверка статуса
sudo ufw status
```

---

## Миграции базы данных

### Как работают миграции

Бот использует **golang-migrate** для автоматического управления схемой БД.

**Миграции применяются автоматически при запуске бота!**

Файлы миграций находятся в: `internal/database/migrations/`

```
migrations/
├── 000001_init.up.sql         # Создание начальной схемы
├── 000001_init.down.sql       # Откат начальной схемы
├── 000002_add_field.up.sql    # Следующая миграция
└── 000002_add_field.down.sql  # Откат
```

### Автоматические миграции (рекомендуется)

При запуске бота миграции применяются автоматически:

```bash
# Локально
docker-compose up -d
docker-compose logs bot | grep -i migrat

# На сервере
cd /opt/fitness-bot
docker compose up -d
docker compose logs bot | grep -i migrat

# Вывод должен содержать:
# Migrations: applied successfully
# ИЛИ
# Migrations: no changes
```

### Создание новой миграции

```bash
# 1. Установите migrate CLI (если ещё не установлен)
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# 2. Создайте новую миграцию
migrate create -ext sql -dir internal/database/migrations -seq название_миграции

# Будут созданы файлы:
# internal/database/migrations/000003_название_миграции.up.sql
# internal/database/migrations/000003_название_миграции.down.sql
```

Пример миграции `000003_add_user_phone.up.sql`:

```sql
-- Добавление поля телефона к пользователям
ALTER TABLE users ADD COLUMN phone VARCHAR(20);
CREATE INDEX idx_users_phone ON users(phone);
```

Откат `000003_add_user_phone.down.sql`:

```sql
-- Откат: удаление поля телефона
DROP INDEX IF EXISTS idx_users_phone;
ALTER TABLE users DROP COLUMN IF EXISTS phone;
```

### Ручное применение миграций (если нужно)

**На локальной машине:**

```bash
# Применить все миграции
migrate -path ./internal/database/migrations \
  -database "postgres://fitness_user:password@localhost:5432/fitness_bot?sslmode=disable" \
  up

# Применить конкретное количество миграций
migrate -path ./internal/database/migrations \
  -database "postgres://fitness_user:password@localhost:5432/fitness_bot?sslmode=disable" \
  up 1

# Откатить последнюю миграцию
migrate -path ./internal/database/migrations \
  -database "postgres://fitness_user:password@localhost:5432/fitness_bot?sslmode=disable" \
  down 1

# Узнать текущую версию
migrate -path ./internal/database/migrations \
  -database "postgres://fitness_user:password@localhost:5432/fitness_bot?sslmode=disable" \
  version
```

**На сервере (через Docker):**

```bash
# Войти в контейнер бота
docker exec -it fitness_bot sh

# Внутри контейнера выполнить миграции вручную не получится,
# так как migrate CLI не установлен в образе.
# Вместо этого подключитесь к БД напрямую:

docker exec -it fitness_postgres psql -U fitness_user -d fitness_bot

# В psql можно выполнить SQL вручную
# Например, из файла миграции:
\i /path/to/migration.sql
```

### Проверка состояния миграций

```bash
# Подключение к БД
docker exec -it fitness_postgres psql -U fitness_user -d fitness_bot

# Проверка таблицы миграций
SELECT * FROM schema_migrations;

# Вывод:
#  version | dirty
# ---------+-------
#        1 | f
```

- `version` - номер последней применённой миграции
- `dirty` - если `true`, миграция не завершилась корректно

### Исправление "грязных" миграций

Если миграция не завершилась (`dirty = true`):

```bash
# Подключение к БД
docker exec -it fitness_postgres psql -U fitness_user -d fitness_bot

# Исправление статуса (ВНИМАНИЕ: делайте это только если уверены!)
UPDATE schema_migrations SET dirty = false;

# Перезапуск бота
docker compose restart bot
```

---

## Обновление бота

### Обновление без изменений БД

**Локально:**

```bash
cd /path/to/fitness-bot

# Получить новый код
git pull

# Пересобрать и перезапустить
docker-compose up -d --build bot

# Проверить логи
docker-compose logs -f bot
```

**На сервере:**

```bash
cd /opt/fitness-bot

# Получить новый код
git pull
# ИЛИ загрузить новые файлы через scp

# Пересобрать и перезапустить
docker compose up -d --build bot

# Проверить логи
docker compose logs -f bot
```

### Обновление с новыми миграциями

Миграции применяются автоматически!

```bash
cd /opt/fitness-bot

# 1. Создать бэкап БД (ВАЖНО!)
docker exec fitness_postgres pg_dump -U fitness_user fitness_bot | \
  gzip > backup_$(date +%Y%m%d_%H%M%S).sql.gz

# 2. Получить новый код
git pull

# 3. Пересобрать образ
docker compose build bot

# 4. Остановить бота (БД продолжает работать)
docker compose stop bot

# 5. Запустить бота (миграции применятся автоматически)
docker compose up -d bot

# 6. Проверить логи миграций
docker compose logs bot | grep -i migrat

# Ожидаемый вывод:
# Migrations: applied successfully

# 7. Проверить работу бота
docker compose logs -f bot
```

### Откат к предыдущей версии

```bash
cd /opt/fitness-bot

# Если используете Git
git log --oneline  # Найти нужный коммит
git checkout <commit_hash>

# Пересобрать и перезапустить
docker compose up -d --build bot

# Если нужно откатить миграции - восстановите из бэкапа:
gunzip < backup_YYYYMMDD_HHMMSS.sql.gz | \
  docker exec -i fitness_postgres psql -U fitness_user fitness_bot
```

### Обновление с Zero Downtime (продвинутый метод)

Для production с высокими требованиями к доступности:

```bash
# 1. Запустить второй экземпляр бота на другом порту
docker run -d --name fitness_bot_new \
  --env-file .env \
  -e TELEGRAM_BOT_TOKEN=new_token \
  fitness_bot:latest

# 2. Проверить работу нового экземпляра
docker logs -f fitness_bot_new

# 3. Остановить старый экземпляр
docker stop fitness_bot

# 4. Переключить токен (или использовать Load Balancer)
# 5. Удалить старый контейнер
docker rm fitness_bot
docker rename fitness_bot_new fitness_bot
```

---

## Мониторинг и логи

### Просмотр логов

```bash
# Все логи в реальном времени
docker compose logs -f

# Только бот
docker compose logs -f bot

# Только PostgreSQL
docker compose logs -f postgres

# Последние 100 строк
docker compose logs --tail=100 bot

# Логи за определённое время
docker compose logs --since="2026-01-20T10:00:00" bot
docker compose logs --since="1h" bot

# Фильтр по ключевым словам
docker compose logs -f bot | grep ERROR
docker compose logs -f bot | grep -i "panic\|error\|fatal"
```

### Сохранение логов в файл

```bash
# Создать директорию
mkdir -p /var/log/fitness-bot

# Запись в файл (в фоне)
nohup docker compose logs -f bot >> /var/log/fitness-bot/bot.log 2>&1 &

# Ротация логов (добавить в cron)
echo "0 0 * * * find /var/log/fitness-bot -name '*.log' -mtime +7 -delete" | crontab -
```

### Проверка состояния системы

```bash
# Статус контейнеров
docker compose ps

# Использование ресурсов
docker stats

# Использование диска
docker system df

# Здоровье PostgreSQL
docker exec fitness_postgres pg_isready -U fitness_user

# Проверка подключения к Telegram API
curl -s https://api.telegram.org/bot<YOUR_TOKEN>/getMe | jq
```

### Мониторинг через systemd

```bash
# Статус сервиса
sudo systemctl status fitness-bot

# Логи сервиса
sudo journalctl -u fitness-bot -f

# Логи с фильтром
sudo journalctl -u fitness-bot --since="1 hour ago"
```

### Настройка алертов (опционально)

Простой мониторинг через cron:

```bash
# Создать скрипт проверки
cat > /opt/fitness-bot/healthcheck.sh << 'EOF'
#!/bin/bash
if ! docker ps | grep -q fitness_bot; then
    echo "ALERT: FitBot is down!" | mail -s "FitBot Alert" admin@example.com
    docker compose -f /opt/fitness-bot/docker-compose.yml restart bot
fi
EOF

chmod +x /opt/fitness-bot/healthcheck.sh

# Добавить в cron (каждые 5 минут)
echo "*/5 * * * * /opt/fitness-bot/healthcheck.sh" | crontab -
```

---

## Бэкапы

### Ручное создание бэкапа

```bash
# Создать директорию для бэкапов
mkdir -p /opt/fitness-bot/backups

# Создать бэкап БД
docker exec fitness_postgres pg_dump -U fitness_user fitness_bot | \
  gzip > /opt/fitness-bot/backups/backup_$(date +%Y%m%d_%H%M%S).sql.gz

# Проверить размер бэкапа
ls -lh /opt/fitness-bot/backups/
```

### Автоматические бэкапы (cron)

```bash
# Создать скрипт бэкапа
cat > /opt/fitness-bot/backup.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/opt/fitness-bot/backups"
DATE=$(date +%Y%m%d_%H%M%S)
docker exec fitness_postgres pg_dump -U fitness_user fitness_bot | \
  gzip > "$BACKUP_DIR/backup_$DATE.sql.gz"
echo "Backup created: backup_$DATE.sql.gz"
EOF

chmod +x /opt/fitness-bot/backup.sh

# Добавить в cron (каждый день в 3:00)
crontab -e

# Добавить строку:
0 3 * * * /opt/fitness-bot/backup.sh >> /var/log/fitness-bot/backup.log 2>&1
```

### Очистка старых бэкапов

```bash
# Удаление бэкапов старше 30 дней
find /opt/fitness-bot/backups -name "*.sql.gz" -mtime +30 -delete

# Добавить в cron (каждый день в 4:00)
0 4 * * * find /opt/fitness-bot/backups -name "*.sql.gz" -mtime +30 -delete
```

### Восстановление из бэкапа

```bash
# Остановить бота
docker compose stop bot

# Восстановить БД
gunzip < /opt/fitness-bot/backups/backup_YYYYMMDD_HHMMSS.sql.gz | \
  docker exec -i fitness_postgres psql -U fitness_user fitness_bot

# Запустить бота
docker compose start bot

# Проверить логи
docker compose logs -f bot
```

### Бэкап в облако (опционально)

```bash
# Установить rclone
curl https://rclone.org/install.sh | sudo bash

# Настроить rclone (например, для Google Drive)
rclone config

# Скрипт для бэкапа в облако
cat > /opt/fitness-bot/backup_cloud.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/opt/fitness-bot/backups"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="backup_$DATE.sql.gz"

# Создать бэкап
docker exec fitness_postgres pg_dump -U fitness_user fitness_bot | \
  gzip > "$BACKUP_DIR/$BACKUP_FILE"

# Загрузить в облако
rclone copy "$BACKUP_DIR/$BACKUP_FILE" remote:fitness-bot-backups/

echo "Backup uploaded to cloud: $BACKUP_FILE"
EOF

chmod +x /opt/fitness-bot/backup_cloud.sh
```

---

## Решение проблем

### Бот не запускается

```bash
# 1. Проверить статус контейнеров
docker compose ps

# 2. Просмотреть логи
docker compose logs bot

# 3. Проверить .env файл
cat .env | grep -v "PASSWORD"

# 4. Проверить токен бота
curl https://api.telegram.org/bot<YOUR_TOKEN>/getMe

# 5. Перезапустить с пересборкой
docker compose down
docker compose up -d --build
```

### Ошибка подключения к БД

```bash
# Проверить статус PostgreSQL
docker compose ps postgres

# Проверить логи PostgreSQL
docker compose logs postgres

# Проверить сетевое подключение
docker exec fitness_bot ping postgres

# Перезапустить PostgreSQL
docker compose restart postgres

# Подождать и перезапустить бота
sleep 10
docker compose restart bot
```

### Ошибки миграций

```bash
# Проверить статус миграций
docker exec -it fitness_postgres psql -U fitness_user -d fitness_bot \
  -c "SELECT * FROM schema_migrations;"

# Если миграция "грязная" (dirty=true)
docker exec -it fitness_postgres psql -U fitness_user -d fitness_bot \
  -c "UPDATE schema_migrations SET dirty=false WHERE version=<NUMBER>;"

# Перезапустить бота
docker compose restart bot

# Если ничего не помогает - восстановить из бэкапа
gunzip < backup_LATEST.sql.gz | \
  docker exec -i fitness_postgres psql -U fitness_user fitness_bot
```

### Бот не отвечает на команды

```bash
# 1. Проверить, что бот работает
docker compose ps

# 2. Проверить логи на ошибки
docker compose logs -f bot | grep -i error

# 3. Проверить, не запущен ли бот где-то ещё
# (Telegram позволяет только 1 активное подключение на токен)

# 4. Проверить токен
curl https://api.telegram.org/bot<YOUR_TOKEN>/getMe

# 5. Перезапустить бота
docker compose restart bot
```

### Высокое использование памяти

```bash
# Проверить использование ресурсов
docker stats

# Ограничить память для контейнера
# Отредактировать docker-compose.yml:
services:
  bot:
    deploy:
      resources:
        limits:
          memory: 512M

# Перезапустить
docker compose up -d
```

### Заканчивается место на диске

```bash
# Проверить использование диска
df -h

# Очистить неиспользуемые Docker-образы
docker system prune -a

# Очистить старые логи
find /var/log -name "*.log" -mtime +30 -delete

# Удалить старые бэкапы
find /opt/fitness-bot/backups -name "*.sql.gz" -mtime +30 -delete
```

### Полный сброс (ВНИМАНИЕ: удалит все данные!)

```bash
# Остановить и удалить всё
docker compose down -v

# Удалить образы
docker rmi $(docker images -q fitness_bot)

# Запустить заново
docker compose up -d --build
```

---

## Дополнительные команды

### Работа с базой данных

```bash
# Подключение к psql
docker exec -it fitness_postgres psql -U fitness_user -d fitness_bot

# Внутри psql:
\dt                           # Список таблиц
\d table_name                 # Структура таблицы
\du                           # Список пользователей
\l                            # Список баз данных
\q                            # Выход

# Полезные SQL-запросы:

# Количество пользователей
SELECT COUNT(*) FROM users;

# Список организаций
SELECT id, name, code FROM organizations;

# Статистика тренировок за последний месяц
SELECT DATE(date), COUNT(*)
FROM workouts
WHERE date >= NOW() - INTERVAL '30 days'
GROUP BY DATE(date)
ORDER BY DATE(date) DESC;
```

### Очистка и оптимизация

```bash
# Очистка неиспользуемых Docker-ресурсов
docker system prune -a --volumes

# Оптимизация PostgreSQL
docker exec -it fitness_postgres vacuumdb -U fitness_user -d fitness_bot --analyze --verbose

# Пересоздание индексов
docker exec -it fitness_postgres psql -U fitness_user -d fitness_bot -c "REINDEX DATABASE fitness_bot;"
```

### Проверка безопасности

```bash
# Проверка открытых портов
netstat -tulpn | grep LISTEN

# Проверка firewall
sudo ufw status

# Проверка Docker security
docker scan fitness_bot:latest

# Аудит контейнера
docker inspect fitness_bot
```

---

## Архитектура системы

```
┌─────────────────────┐
│   Telegram API      │
└──────────┬──────────┘
           │
           │ HTTPS
           │
┌──────────▼──────────┐
│     FitBot          │
│   (Go 1.21+)        │
│                     │
│  • handlers         │
│  • bot logic        │
│  • inline keyboards │
│  • state machine    │
└──────────┬──────────┘
           │
           │ SQL
           │
┌──────────▼──────────┐
│   PostgreSQL 15     │
│                     │
│  • users            │
│  • organizations    │
│  • workouts         │
│  • migrations       │
└─────────────────────┘
```

### Иерархия доступа

```
ADMIN (@username из .env)
  │
  └── Организация (Organization)
         │
         └── Менеджеры (Organization Managers)
                │
                └── Тренеры (Trainers)
                       │
                       └── Клиенты (Clients)
```

---

## Контакты и поддержка

При возникновении проблем:

1. **Проверьте логи:** `docker compose logs -f bot`
2. **Перезапустите:** `docker compose restart`
3. **Проверьте конфигурацию:** `cat .env`
4. **Проверьте статус:** `docker compose ps`
5. **Создайте issue:** на GitHub (если используете)

---

**Версия документа:** 2.0
**Последнее обновление:** 2026-01-23
