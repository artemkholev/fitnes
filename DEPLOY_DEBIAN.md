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
DB_HOST=postgres
DB_PORT=5432
DB_USER=fitness_user
DB_PASSWORD=ПРИДУМАЙТЕ_СЛОЖНЫЙ_ПАРОЛЬ
DB_NAME=fitness_bot
APP_ENV=production
```

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
Bot started successfully!
```

**Готово!** Найдите бота в Telegram и отправьте `/start`

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

## Решение проблем

**Бот не отвечает:**
```bash
docker-compose logs bot
# Проверьте токен в .env
```

**Ошибка "port already in use":**
```bash
docker-compose down
docker-compose up -d
```

**Перезапуск с нуля:**
```bash
docker-compose down -v
docker-compose up -d
```

---

## Создание организации (после запуска)

```bash
docker exec -it fitness_postgres psql -U fitness_user -d fitness_bot -c "INSERT INTO organizations (name, code) VALUES ('Мой Клуб', 'GYM2024');"
```

Теперь пользователи могут использовать код `GYM2024` при регистрации.
