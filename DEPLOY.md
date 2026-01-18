# Инструкция по быстрому деплою

## 🚀 Быстрый старт (5 минут)

### Шаг 1: Подготовка Telegram бота

1. Откройте [@BotFather](https://t.me/BotFather) в Telegram
2. Отправьте `/newbot`
3. Придумайте имя и username для бота
4. Скопируйте полученный **токен** (выглядит как `1234567890:ABCdefGHIjklMNOpqrsTUVwxyz`)

### Шаг 2: Подготовка сервера

Подключитесь к серверу по SSH:
```bash
ssh user@your-server-ip
```

Установите Docker (если не установлен):
```bash
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER
```

**ВАЖНО**: После установки Docker выйдите и войдите снова!

### Шаг 3: Загрузка проекта

Создайте директорию и скопируйте файлы проекта:
```bash
mkdir -p ~/fitness-bot
cd ~/fitness-bot
```

Скопируйте все файлы проекта в эту директорию.

Или через SCP с вашего компьютера:
```bash
# На вашем компьютере
scp -r /Users/artemkholev/Desktop/fitnes/* user@your-server-ip:~/fitness-bot/
```

### Шаг 4: Настройка переменных окружения

```bash
cd ~/fitness-bot
nano .env
```

Вставьте и заполните:
```env
TELEGRAM_BOT_TOKEN=ВАШ_ТОКЕН_ОТ_BOTFATHER
DB_HOST=postgres
DB_PORT=5432
DB_USER=fitness_user
DB_PASSWORD=ПРИДУМАЙТЕ_СЛОЖНЫЙ_ПАРОЛЬ
DB_NAME=fitness_bot
APP_ENV=production
```

Сохраните: `Ctrl+O`, `Enter`, `Ctrl+X`

### Шаг 5: Запуск

```bash
docker-compose up -d
```

### Шаг 6: Проверка

```bash
docker-compose logs -f bot
```

Вы должны увидеть:
```
Database connected successfully
Authorized on account YourBotName
Bot started successfully!
```

**Готово!** Найдите бота в Telegram и отправьте `/start`

---

## 📋 Пошаговая инструкция для деплоя

### 1. Проверка требований

Проверьте наличие Docker:
```bash
docker --version
docker-compose --version
```

Если команды не найдены, установите Docker:
```bash
# Ubuntu/Debian
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# После установки выйдите и войдите снова
exit
```

### 2. Загрузка файлов на сервер

**Вариант A: Через Git**
```bash
cd ~
git clone YOUR_REPOSITORY_URL fitness-bot
cd fitness-bot
```

**Вариант B: Через SCP (с локального компьютера)**
```bash
# Создайте архив локально
cd /Users/artemkholev/Desktop
tar -czf fitnes.tar.gz fitnes/

# Загрузите на сервер
scp fitnes.tar.gz user@SERVER_IP:~

# На сервере распакуйте
ssh user@SERVER_IP
tar -xzf fitnes.tar.gz
cd fitnes
```

**Вариант C: Ручное создание файлов**
```bash
mkdir -p ~/fitness-bot
cd ~/fitness-bot
# Создайте файлы вручную через nano/vim
```

### 3. Настройка .env файла

```bash
cd ~/fitness-bot
cp .env.example .env
nano .env
```

**Важные переменные:**

| Переменная | Описание | Пример |
|------------|----------|--------|
| TELEGRAM_BOT_TOKEN | Токен от @BotFather | 1234567890:ABC... |
| DB_PASSWORD | Пароль БД (придумайте сложный) | MyS3cur3P@ssw0rd! |
| DB_USER | Пользователь БД | fitness_user |
| DB_NAME | Имя БД | fitness_bot |

### 4. Запуск проекта

```bash
# Запуск в фоновом режиме
docker-compose up -d

# Проверка статуса
docker-compose ps

# Просмотр логов
docker-compose logs -f
```

### 5. Проверка работы

**Проверка контейнеров:**
```bash
docker-compose ps
```

Должно быть:
```
NAME                 STATUS
fitness_bot          Up
fitness_postgres     Up (healthy)
```

**Проверка логов бота:**
```bash
docker-compose logs bot | tail -20
```

Должно быть:
```
Database connected successfully
Authorized on account ...
Bot started successfully!
```

**Проверка БД:**
```bash
docker exec -it fitness_postgres psql -U fitness_user -d fitness_bot -c "\dt"
```

Должны быть таблицы: users, workouts, exercises, etc.

### 6. Первое использование

1. Откройте Telegram
2. Найдите вашего бота по username
3. Отправьте `/start`
4. Выберите роль (Клиент или Тренер)

### 7. (Опционально) Создание организации

```bash
# Войдите в БД
docker exec -it fitness_postgres psql -U fitness_user -d fitness_bot

# Создайте организацию
INSERT INTO organizations (name, code) VALUES ('Мой Фитнес Клуб', 'GYM2024');

# Выход
\q
```

Теперь пользователи могут использовать код `GYM2024` при регистрации.

---

## 🔧 Управление и обслуживание

### Просмотр логов

```bash
# Все логи
docker-compose logs -f

# Только бот
docker-compose logs -f bot

# Последние 100 строк
docker-compose logs --tail=100 bot
```

### Перезапуск

```bash
# Перезапуск бота
docker-compose restart bot

# Перезапуск всего
docker-compose restart
```

### Остановка

```bash
# Остановка без удаления данных
docker-compose stop

# Остановка с удалением контейнеров (данные сохраняются)
docker-compose down

# Полное удаление (ВНИМАНИЕ: удалит все данные!)
docker-compose down -v
```

### Обновление кода

```bash
cd ~/fitness-bot

# Остановите проект
docker-compose down

# Обновите код (через git pull или загрузите новые файлы)
git pull  # если используете Git

# Пересоберите и запустите
docker-compose up -d --build

# Проверьте логи
docker-compose logs -f bot
```

### Резервное копирование

**Создание backup:**
```bash
# Backup базы данных
docker exec fitness_postgres pg_dump -U fitness_user fitness_bot > backup_$(date +%Y%m%d_%H%M%S).sql

# Backup всех данных Docker volume
docker run --rm -v fitness_postgres_data:/data -v $(pwd):/backup alpine tar czf /backup/db_volume_$(date +%Y%m%d).tar.gz /data
```

**Восстановление:**
```bash
# Восстановление из SQL файла
docker exec -i fitness_postgres psql -U fitness_user fitness_bot < backup_20260118_120000.sql
```

**Автоматический backup (cron):**
```bash
# Создайте скрипт
nano ~/backup_fitness.sh
```

Содержимое:
```bash
#!/bin/bash
cd ~/fitness-bot
docker exec fitness_postgres pg_dump -U fitness_user fitness_bot > ~/backups/fitness_$(date +%Y%m%d).sql
# Удалить backups старше 30 дней
find ~/backups -name "fitness_*.sql" -mtime +30 -delete
```

```bash
# Сделайте исполняемым
chmod +x ~/backup_fitness.sh

# Добавьте в cron (каждый день в 3 утра)
crontab -e
# Добавьте строку:
0 3 * * * /home/user/backup_fitness.sh
```

---

## 🔒 Безопасность

### Настройка Firewall

```bash
# UFW (Ubuntu)
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP (если нужен)
sudo ufw allow 443/tcp   # HTTPS (если нужен)
sudo ufw enable
```

### Обновление системы

```bash
# Регулярно обновляйте сервер
sudo apt update && sudo apt upgrade -y

# Обновление Docker образов
cd ~/fitness-bot
docker-compose pull
docker-compose up -d
```

### Защита .env файла

```bash
# Ограничьте права доступа
chmod 600 .env
```

---

## ❓ Решение проблем

### Проблема: Бот не запускается

**Решение:**
```bash
# Проверьте логи
docker-compose logs bot

# Частые причины:
# 1. Неверный токен - проверьте TELEGRAM_BOT_TOKEN в .env
# 2. БД не готова - подождите 10-15 секунд и проверьте снова
# 3. Бот уже запущен где-то ещё - остановите другие инстансы
```

### Проблема: Ошибка "connection refused" к БД

**Решение:**
```bash
# Проверьте статус PostgreSQL
docker-compose ps postgres

# Проверьте логи БД
docker-compose logs postgres

# Перезапустите БД
docker-compose restart postgres
```

### Проблема: "No space left on device"

**Решение:**
```bash
# Очистите старые Docker образы
docker system prune -a

# Проверьте место на диске
df -h
```

### Проблема: Порт 5432 уже занят

**Решение:**
```bash
# Найдите процесс, использующий порт
sudo lsof -i :5432

# Остановите локальный PostgreSQL (если установлен)
sudo systemctl stop postgresql

# Или измените порт в docker-compose.yml
```

---

## 📞 Контакты и поддержка

При возникновении проблем:
1. Проверьте логи: `docker-compose logs -f`
2. Убедитесь, что все переменные в .env заполнены правильно
3. Проверьте, что Docker и Docker Compose установлены корректно

Удачного использования! 💪
