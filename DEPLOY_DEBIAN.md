# Деплой на Debian сервер

## Подготовка

### 1. Получите токен бота
Откройте [@BotFather](https://t.me/BotFather) в Telegram:
```
/newbot → Введите имя → Введите username → Скопируйте токен
```

### 2. Подключитесь к серверу
```bash
ssh root@ВАШ_IP_СЕРВЕРА
```

---

## Установка (выполняйте по порядку)

### Шаг 1: Обновите систему
```bash
apt update && apt upgrade -y
```

### Шаг 2: Установите Docker
```bash
curl -fsSL https://get.docker.com | sh
```

### Шаг 3: Установите Docker Compose
```bash
apt install docker-compose -y
```

### Шаг 4: Создайте директорию проекта
```bash
mkdir -p /opt/fitness-bot
cd /opt/fitness-bot
```

### Шаг 5: Загрузите файлы проекта

**Вариант A: С вашего компьютера через SCP**

На вашем Mac выполните:
```bash
scp -r /Users/artemkholev/Desktop/fitnes/* root@ВАШ_IP:/opt/fitness-bot/
```

**Вариант B: Через Git (если загрузили в репозиторий)**
```bash
apt install git -y
git clone ВАШ_РЕПОЗИТОРИЙ /opt/fitness-bot
```

### Шаг 6: Создайте .env файл
```bash
cd /opt/fitness-bot
nano .env
```

Вставьте (замените значения на свои):
```
TELEGRAM_BOT_TOKEN=ВАШ_ТОКЕН_ОТ_BOTFATHER
ADMIN_USERNAME=ваш_telegram_username
DB_HOST=postgres
DB_PORT=5432
DB_USER=fitness_user
DB_PASSWORD=ПРИДУМАЙТЕ_СЛОЖНЫЙ_ПАРОЛЬ
DB_NAME=fitness_bot
APP_ENV=production
```

**ВАЖНО:** `ADMIN_USERNAME` - это ваш username в Telegram (без @). Только этот пользователь будет иметь доступ к админ-панели.

Сохраните: `Ctrl+O` → `Enter` → `Ctrl+X`

### Шаг 7: Запустите бота
```bash
docker-compose up -d
```

### Шаг 8: Проверьте работу
```bash
docker-compose logs -f bot
```

Должно появиться:
```
Database connected successfully
Admin username: @ваш_username
Bot started successfully!
```

**Готово!** Найдите бота в Telegram и отправьте `/start`

---

## Как работает система доступов

Бот использует иерархическую систему доступов:

```
АДМИН (указан в ADMIN_USERNAME в .env)
    ↓ создаёт организации и назначает менеджеров
МЕНЕДЖЕРЫ (@username)
    ↓ добавляют тренеров в свою организацию
ТРЕНЕРЫ (@username)
    ↓ добавляют клиентов
КЛИЕНТЫ (@username)
    → тренируются и отслеживают прогресс
```

### Первый запуск:

1. Напишите боту `/start` от имени админа
2. Нажмите "👑 Админ-панель"
3. Создайте организацию
4. Добавьте менеджера по @username
5. Менеджер добавит тренеров
6. Тренеры добавят клиентов

---

## Полезные команды

```bash
# Просмотр логов
docker-compose logs -f

# Перезапуск
docker-compose restart

# Остановка
docker-compose down

# Статус
docker-compose ps
```

---

## Автозапуск при перезагрузке сервера

```bash
cat > /etc/systemd/system/fitness-bot.service << 'EOF'
[Unit]
Description=Fitness Bot
After=docker.service
Requires=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=/opt/fitness-bot
ExecStart=/usr/bin/docker-compose up -d
ExecStop=/usr/bin/docker-compose down

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable fitness-bot
```

---

## Обновление бота

При обновлении кода на сервере:
```bash
cd /opt/fitness-bot
git pull  # если используете git
docker-compose build
docker-compose up -d
```

---

## Решение проблем

**Бот не отвечает:**
```bash
docker-compose logs bot
# Проверьте токен и ADMIN_USERNAME в .env
```

**Ошибка "port already in use":**
```bash
docker-compose down
docker-compose up -d
```

**Сброс базы данных (ВНИМАНИЕ: удаляются все данные!):**
```bash
docker-compose down -v
docker-compose up -d
```

**Применение новых миграций:**
```bash
docker exec -it fitness_postgres psql -U fitness_user -d fitness_bot -f /docker-entrypoint-initdb.d/002_access_system.sql
```
