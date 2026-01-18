#!/bin/bash

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=====================================${NC}"
echo -e "${GREEN}Фитнес Бот - Скрипт установки${NC}"
echo -e "${GREEN}=====================================${NC}\n"

# Проверка Docker
echo -e "${YELLOW}Проверка Docker...${NC}"
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Docker не установлен!${NC}"
    echo -e "${YELLOW}Установить Docker? (y/n)${NC}"
    read -r answer
    if [ "$answer" == "y" ]; then
        echo -e "${GREEN}Установка Docker...${NC}"
        curl -fsSL https://get.docker.com -o get-docker.sh
        sudo sh get-docker.sh
        sudo usermod -aG docker $USER
        rm get-docker.sh
        echo -e "${GREEN}Docker установлен!${NC}"
        echo -e "${YELLOW}Перезайдите в систему для применения изменений!${NC}"
        exit 0
    else
        echo -e "${RED}Установка прервана.${NC}"
        exit 1
    fi
else
    echo -e "${GREEN}✓ Docker установлен${NC}"
fi

# Проверка Docker Compose
echo -e "${YELLOW}Проверка Docker Compose...${NC}"
if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}Docker Compose не установлен!${NC}"
    echo -e "${GREEN}Установка Docker Compose...${NC}"
    sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    sudo chmod +x /usr/local/bin/docker-compose
    echo -e "${GREEN}✓ Docker Compose установлен${NC}"
else
    echo -e "${GREEN}✓ Docker Compose установлен${NC}"
fi

# Создание .env файла
echo -e "\n${YELLOW}Настройка переменных окружения...${NC}"
if [ -f .env ]; then
    echo -e "${YELLOW}.env файл уже существует. Перезаписать? (y/n)${NC}"
    read -r answer
    if [ "$answer" != "y" ]; then
        echo -e "${GREEN}Используется существующий .env файл${NC}"
    else
        rm .env
    fi
fi

if [ ! -f .env ]; then
    echo -e "${GREEN}Создание .env файла...${NC}\n"

    # Telegram Bot Token
    echo -e "${YELLOW}Введите Telegram Bot Token (от @BotFather):${NC}"
    read -r bot_token

    # Database Password
    echo -e "${YELLOW}Введите пароль для базы данных (или нажмите Enter для автогенерации):${NC}"
    read -r db_password
    if [ -z "$db_password" ]; then
        db_password=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-25)
        echo -e "${GREEN}Сгенерирован пароль: $db_password${NC}"
    fi

    # Создание .env
    cat > .env << EOF
# Telegram Bot Token
TELEGRAM_BOT_TOKEN=$bot_token

# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=fitness_user
DB_PASSWORD=$db_password
DB_NAME=fitness_bot

# Application
APP_ENV=production
EOF

    echo -e "${GREEN}✓ .env файл создан${NC}"
fi

# Создание директории для backup
if [ ! -d backups ]; then
    mkdir backups
    echo -e "${GREEN}✓ Директория backups создана${NC}"
fi

# Запуск проекта
echo -e "\n${YELLOW}Запустить проект сейчас? (y/n)${NC}"
read -r answer
if [ "$answer" == "y" ]; then
    echo -e "${GREEN}Запуск проекта...${NC}"
    docker-compose up -d

    echo -e "\n${YELLOW}Ожидание запуска контейнеров...${NC}"
    sleep 10

    echo -e "\n${GREEN}Статус контейнеров:${NC}"
    docker-compose ps

    echo -e "\n${GREEN}Последние логи бота:${NC}"
    docker-compose logs --tail=20 bot

    echo -e "\n${GREEN}=====================================${NC}"
    echo -e "${GREEN}Установка завершена!${NC}"
    echo -e "${GREEN}=====================================${NC}\n"

    echo -e "Полезные команды:"
    echo -e "  ${YELLOW}docker-compose logs -f bot${NC}  - Просмотр логов"
    echo -e "  ${YELLOW}docker-compose ps${NC}          - Статус контейнеров"
    echo -e "  ${YELLOW}docker-compose restart${NC}      - Перезапуск"
    echo -e "  ${YELLOW}docker-compose down${NC}         - Остановка"
    echo -e "\nИли используйте ${YELLOW}make help${NC} для списка команд\n"
else
    echo -e "${GREEN}Для запуска выполните: ${YELLOW}docker-compose up -d${NC}"
fi
