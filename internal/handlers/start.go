package handlers

import (
	"fitness-bot/internal/bot"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleStart - устаревшая функция, оставлена для совместимости
// Основная логика старта теперь в main.go handleStartCommand
func HandleStart(b *bot.Bot, message *tgbotapi.Message) {
	// Эта функция больше не используется напрямую
	// Вся логика старта перенесена в main.go
	b.SendMessage(message.Chat.ID, "Используйте /start для начала работы.")
}

// HandleRoleSelection - устаревшая функция
// В новой системе роли определяются через таблицы доступов
func HandleRoleSelection(b *bot.Bot, message *tgbotapi.Message) {
	b.SendMessage(message.Chat.ID, "Эта функция устарела. Ваша роль определяется автоматически.")
}

// HandleOrgCode - устаревшая функция
// В новой системе организации назначаются через админ-панель
func HandleOrgCode(b *bot.Bot, message *tgbotapi.Message) {
	b.SendMessage(message.Chat.ID, "Эта функция устарела. Обратитесь к администратору для добавления в организацию.")
}
