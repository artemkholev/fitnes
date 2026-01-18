.PHONY: help build up down restart logs clean backup

help: ## Показать помощь
	@echo "Доступные команды:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Собрать проект
	docker-compose build

up: ## Запустить проект
	docker-compose up -d
	@echo "Проект запущен! Проверьте логи: make logs"

down: ## Остановить проект
	docker-compose down

restart: ## Перезапустить проект
	docker-compose restart

logs: ## Показать логи
	docker-compose logs -f

logs-bot: ## Показать логи бота
	docker-compose logs -f bot

logs-db: ## Показать логи БД
	docker-compose logs -f postgres

status: ## Показать статус контейнеров
	docker-compose ps

clean: ## Очистить все (ВНИМАНИЕ: удалит данные!)
	docker-compose down -v
	@echo "Все данные удалены!"

backup: ## Создать backup БД
	@mkdir -p backups
	docker exec fitness_postgres pg_dump -U fitness_user fitness_bot > backups/backup_$$(date +%Y%m%d_%H%M%S).sql
	@echo "Backup создан в директории backups/"

restore: ## Восстановить БД из последнего backup
	@if [ -z "$$(ls -t backups/*.sql 2>/dev/null | head -1)" ]; then \
		echo "Backup файлы не найдены!"; \
		exit 1; \
	fi
	@LATEST=$$(ls -t backups/*.sql | head -1); \
	echo "Восстановление из $$LATEST..."; \
	docker exec -i fitness_postgres psql -U fitness_user fitness_bot < $$LATEST
	@echo "Восстановление завершено!"

db-shell: ## Войти в shell БД
	docker exec -it fitness_postgres psql -U fitness_user fitness_bot

update: ## Обновить и перезапустить проект
	git pull
	docker-compose up -d --build
	@echo "Проект обновлён!"

dev: ## Запустить в режиме разработки (с логами)
	docker-compose up

install-deps: ## Установить зависимости Go
	go mod download
	go mod tidy

run-local: ## Запустить локально (без Docker)
	@if [ ! -f .env ]; then echo ".env файл не найден!"; exit 1; fi
	go run cmd/bot/main.go
