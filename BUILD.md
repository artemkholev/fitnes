# Сборка и тестирование проекта

## Инициализация зависимостей

После клонирования проекта выполните:

```bash
# Загрузить все Go зависимости
go mod download

# Обновить go.mod и go.sum
go mod tidy
```

## Сборка проекта

### Локальная сборка

```bash
# Сборка для вашей платформы
go build -o fitness-bot ./cmd/bot

# Запуск
./fitness-bot
```

### Сборка для Linux (для деплоя)

```bash
# Если вы на macOS/Windows и деплоите на Linux сервер
GOOS=linux GOARCH=amd64 go build -o fitness-bot ./cmd/bot
```

### Сборка с оптимизацией (меньший размер)

```bash
go build -ldflags="-s -w" -o fitness-bot ./cmd/bot
```

## Проверка кода

### Форматирование

```bash
# Форматирование всего кода
go fmt ./...

# Проверка импортов
goimports -w .
```

### Линтер

```bash
# Установка golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Запуск линтера
golangci-lint run
```

### Проверка на ошибки

```bash
# Проверка компиляции без создания бинарника
go build -o /dev/null ./cmd/bot

# Проверка всех пакетов
go vet ./...
```

## Запуск в режиме разработки

### Вариант 1: С go run

```bash
# Создайте .env файл
cp .env.example .env
nano .env  # Заполните переменные

# Запустите PostgreSQL в Docker
docker-compose up -d postgres

# Измените DB_HOST в .env на localhost
# DB_HOST=localhost

# Запустите бота
go run cmd/bot/main.go
```

### Вариант 2: С air (hot reload)

```bash
# Установите air для автоперезагрузки
go install github.com/cosmtrek/air@latest

# Создайте .air.toml
cat > .air.toml << 'EOF'
root = "."
tmp_dir = "tmp"

[build]
cmd = "go build -o ./tmp/main ./cmd/bot"
bin = "tmp/main"
include_ext = ["go"]
exclude_dir = ["tmp", "vendor"]
EOF

# Запустите с автоперезагрузкой
air
```

## Docker сборка

### Локальная сборка Docker образа

```bash
# Собрать образ
docker build -t fitness-bot:latest .

# Запустить контейнер
docker run --env-file .env fitness-bot:latest
```

### Сборка с docker-compose

```bash
# Собрать и запустить
docker-compose up --build

# Только собрать без запуска
docker-compose build
```

## Тестирование

### Создание тестов

Создайте файлы `*_test.go` в соответствующих пакетах:

```go
// internal/database/users_test.go
package database

import (
    "testing"
)

func TestCreateUser(t *testing.T) {
    // Ваши тесты
}
```

### Запуск тестов

```bash
# Все тесты
go test ./...

# С покрытием
go test -cover ./...

# Подробный вывод
go test -v ./...

# Генерация HTML отчета покрытия
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Устранение проблем

### Ошибка: "could not import ... (no required module provides package)"

**Решение:**
```bash
go mod download
go mod tidy
```

### Ошибка: "undefined: pgxpool"

**Решение:**
```bash
# Убедитесь, что зависимость есть в go.mod
grep pgx go.mod

# Если нет, добавьте
go get github.com/jackc/pgx/v5
go mod tidy
```

### Проблемы с версиями Go

```bash
# Проверьте версию Go
go version

# Должно быть Go 1.21 или выше
# Обновите если необходимо: https://go.dev/dl/
```

### Конфликты зависимостей

```bash
# Очистите кеш модулей
go clean -modcache

# Заново скачайте зависимости
go mod download
```

## Оптимизация

### Уменьшение размера бинарника

```bash
# С флагами оптимизации
go build -ldflags="-s -w" -o fitness-bot ./cmd/bot

# С UPX компрессией (требует установки upx)
upx --best --lzma fitness-bot
```

### Профилирование

```go
// Добавьте в main.go для профилирования
import (
    "runtime/pprof"
    "os"
)

// CPU профиль
f, _ := os.Create("cpu.prof")
pprof.StartCPUProfile(f)
defer pprof.StopCPUProfile()
```

Анализ:
```bash
go tool pprof cpu.prof
```

## CI/CD примеры

### GitHub Actions

```yaml
# .github/workflows/build.yml
name: Build and Test

on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: go mod download
      - run: go build -v ./...
      - run: go test -v ./...
```

### GitLab CI

```yaml
# .gitlab-ci.yml
image: golang:1.21

stages:
  - build
  - test

build:
  stage: build
  script:
    - go mod download
    - go build -o fitness-bot ./cmd/bot
  artifacts:
    paths:
      - fitness-bot

test:
  stage: test
  script:
    - go test -v ./...
```

## Переменные окружения для разработки

```bash
# .env для локальной разработки
TELEGRAM_BOT_TOKEN=your_test_bot_token
DB_HOST=localhost
DB_PORT=5432
DB_USER=fitness_user
DB_PASSWORD=dev_password
DB_NAME=fitness_bot_dev
APP_ENV=development
```

## Полезные команды

```bash
# Показать зависимости
go list -m all

# Обновить зависимости
go get -u ./...
go mod tidy

# Информация о модуле
go mod graph

# Очистка
go clean
go clean -cache
go clean -modcache

# Проверка на уязвимости
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

## Готово!

После всех этих шагов проект должен компилироваться без ошибок.

Для быстрой проверки:
```bash
go mod tidy && go build -o fitness-bot ./cmd/bot && echo "✅ Всё работает!"
```
