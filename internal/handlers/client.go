package handlers

import (
	"fitness-bot/internal/bot"
	"fitness-bot/internal/models"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleClientMenu(b *bot.Bot, message *tgbotapi.Message, clientAccess []*models.ClientAccessInfo) {
	if len(clientAccess) == 0 {
		b.SendMessage(message.Chat.ID, "❌ У вас нет активных доступов к тренерам.\n\nПопросите тренера добавить вас по @username.")
		return
	}

	if len(clientAccess) == 1 {
		showClientTrainerMenu(b, message, clientAccess[0])
		return
	}

	var sb strings.Builder
	sb.WriteString("🏋️ *Выберите тренера:*\n\n")

	for i, access := range clientAccess {
		sb.WriteString(fmt.Sprintf("%d. @%s (%s)\n", i+1, access.TrainerUsername, access.OrganizationName))
	}

	b.SendMessage(message.Chat.ID, sb.String())
	b.SetState(message.From.ID, "client_selecting_trainer", map[string]interface{}{
		"trainers": clientAccess,
	})
}

func showClientTrainerMenu(b *bot.Bot, message *tgbotapi.Message, access *models.ClientAccessInfo) {
	b.SetState(message.From.ID, "client_with_trainer", map[string]interface{}{
		"trainer_client_id": access.TrainerClientID,
		"trainer_id":        access.TrainerID,
		"trainer_username":  access.TrainerUsername,
		"org_id":            access.OrganizationID,
		"org_name":          access.OrganizationName,
	})
	b.SendMessageWithKeyboard(
		message.Chat.ID,
		fmt.Sprintf("📝 *Тренировки с @%s*\n_Организация: %s_\n\nВыберите действие:", access.TrainerUsername, access.OrganizationName),
		bot.GetClientMenuKeyboard(),
	)
}

func HandleClientSelectTrainer(b *bot.Bot, message *tgbotapi.Message, idx int) {
	state := b.GetState(message.From.ID)
	if state == nil {
		b.SendMessage(message.Chat.ID, "❌ Список тренеров устарел. Попробуйте /start")
		return
	}

	trainers, ok := state.Data["trainers"].([]*models.ClientAccessInfo)
	if !ok || len(trainers) == 0 || idx < 1 || idx > len(trainers) {
		b.SendMessage(message.Chat.ID, "❌ Неверный номер. Попробуйте снова.")
		return
	}

	showClientTrainerMenu(b, message, trainers[idx-1])
}

func HandleArchivedAccess(b *bot.Bot, message *tgbotapi.Message, archivedAccess []*models.ClientAccessInfo) {
	if len(archivedAccess) == 0 {
		b.SendMessage(message.Chat.ID, "📚 У вас нет архивных тренировок.")
		return
	}

	var sb strings.Builder
	sb.WriteString("📚 *Архивные тренировки*\n")
	sb.WriteString("_Доступ завершён, но история сохранена_\n\n")

	for i, access := range archivedAccess {
		sb.WriteString(fmt.Sprintf("%d. @%s (%s)\n", i+1, access.TrainerUsername, access.OrganizationName))
	}

	sb.WriteString("\nВыберите номер для просмотра истории:")

	b.SendMessage(message.Chat.ID, sb.String())
	b.SetState(message.From.ID, "client_viewing_archive", map[string]interface{}{
		"archived": archivedAccess,
	})
}

func HandleSelectArchivedTrainer(b *bot.Bot, message *tgbotapi.Message, idx int) {
	state := b.GetState(message.From.ID)
	if state == nil {
		b.SendMessage(message.Chat.ID, "❌ Список устарел. Попробуйте /start")
		return
	}

	archived, ok := state.Data["archived"].([]*models.ClientAccessInfo)
	if !ok || len(archived) == 0 || idx < 1 || idx > len(archived) {
		b.SendMessage(message.Chat.ID, "❌ Неверный номер.")
		return
	}

	access := archived[idx-1]
	// TODO: показать историю тренировок с этим тренером
	b.SendMessage(message.Chat.ID, fmt.Sprintf("📋 История тренировок с @%s будет добавлена позже.", access.TrainerUsername))
}

func HandleNoAccess(b *bot.Bot, message *tgbotapi.Message) {
	msg := `👋 *Добро пожаловать в FitBot!*

У вас пока нет доступов к системе.

*Как начать:*
1. Попросите тренера добавить вас по @username
2. После добавления напишите /start снова

*Что вы сможете делать:*
• Записывать тренировки
• Отслеживать прогресс
• Смотреть статистику и графики
• Записываться на групповые тренировки`

	b.SendMessage(message.Chat.ID, msg)
}
