package handlers

import (
	"context"
	"fitness-bot/internal/bot"
	"fitness-bot/internal/database"
	"fitness-bot/internal/models"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleManagerMenu показывает меню менеджера
func HandleManagerMenu(b *bot.Bot, message *tgbotapi.Message, managerOrgs []*models.ManagerOrgInfo) {
	if len(managerOrgs) == 0 {
		b.SendMessage(message.Chat.ID, "❌ У вас нет доступа менеджера ни к одной организации.")
		return
	}

	// Если одна организация - сразу показываем управление
	if len(managerOrgs) == 1 {
		org := managerOrgs[0]
		if !org.IsActive {
			b.SendMessage(message.Chat.ID, "❌ Ваш доступ к организации был деактивирован.")
			return
		}
		showManagerOrgMenu(b, message, org.Organization.ID, org.Organization.Name)
		return
	}

	// Несколько организаций - показываем выбор
	var sb strings.Builder
	sb.WriteString("🏢 *Выберите организацию для управления:*\n\n")

	activeOrgs := []*models.ManagerOrgInfo{}
	for _, org := range managerOrgs {
		if org.IsActive {
			activeOrgs = append(activeOrgs, org)
			sb.WriteString(fmt.Sprintf("%d. %s\n", len(activeOrgs), org.Organization.Name))
		}
	}

	if len(activeOrgs) == 0 {
		b.SendMessage(message.Chat.ID, "❌ Все ваши доступы к организациям были деактивированы.")
		return
	}

	b.SendMessage(message.Chat.ID, sb.String())
	b.SetState(message.From.ID, "manager_selecting_org", map[string]interface{}{
		"organizations": activeOrgs,
	})
}

func showManagerOrgMenu(b *bot.Bot, message *tgbotapi.Message, orgID int64, orgName string) {
	b.SetState(message.From.ID, "manager_managing_org", map[string]interface{}{
		"org_id":   orgID,
		"org_name": orgName,
	})
	b.SendMessageWithKeyboard(
		message.Chat.ID,
		fmt.Sprintf("🏢 *Управление организацией %s*\n\nКак менеджер вы можете добавлять и удалять тренеров.", orgName),
		bot.GetManagerMenuKeyboard(),
	)
}

// HandleManagerSelectOrg выбор организации менеджером
func HandleManagerSelectOrg(b *bot.Bot, message *tgbotapi.Message, idx int) {
	state := b.GetState(message.From.ID)
	if state == nil {
		b.SendMessage(message.Chat.ID, "❌ Список организаций устарел. Попробуйте снова.")
		return
	}

	orgs, ok := state.Data["organizations"].([]*models.ManagerOrgInfo)
	if !ok || len(orgs) == 0 || idx < 1 || idx > len(orgs) {
		b.SendMessage(message.Chat.ID, "❌ Неверный номер или список устарел.")
		return
	}

	org := orgs[idx-1]
	showManagerOrgMenu(b, message, org.Organization.ID, org.Organization.Name)
}

// HandleAddTrainer начинает добавление тренера
func HandleAddTrainer(b *bot.Bot, message *tgbotapi.Message) {
	state := b.GetState(message.From.ID)
	if state == nil {
		b.SendMessage(message.Chat.ID, "❌ Сначала выберите организацию.")
		return
	}

	_, okID := bot.GetStateInt64(state.Data, "org_id")
	_, okName := bot.GetStateString(state.Data, "org_name")
	if !okID || !okName {
		b.SendMessage(message.Chat.ID, "❌ Сначала выберите организацию.")
		return
	}

	b.SetState(message.From.ID, "manager_adding_trainer", bot.CopyStateData(state.Data))
	b.SendMessageWithKeyboard(
		message.Chat.ID,
		"Введите @username тренера (например: @trainer\\_ivan):",
		bot.GetCancelKeyboard(),
	)
}

// HandleAddTrainerUsername обрабатывает ввод username тренера
func HandleAddTrainerUsername(b *bot.Bot, message *tgbotapi.Message) {
	ctx := context.Background()
	state := b.GetState(message.From.ID)

	// Безопасное извлечение данных
	orgID, okID := bot.GetStateInt64(state.Data, "org_id")
	orgName, okName := bot.GetStateString(state.Data, "org_name")
	if !okID || !okName {
		b.ClearState(message.From.ID)
		b.SendMessage(message.Chat.ID, "❌ Ошибка состояния. Попробуйте снова.")
		return
	}

	if message.Text == "❌ Отмена" {
		showManagerOrgMenu(b, message, orgID, orgName)
		return
	}

	username := database.NormalizeUsername(message.Text)
	if username == "" {
		b.SendWithCancel(message.Chat.ID, "❌ Некорректный username. Введите в формате @username:")
		return
	}

	if err := b.DB.AddTrainer(ctx, orgID, username); err != nil {
		log.Printf("Error adding trainer: %v", err)
		errStr := err.Error()
		if strings.Contains(errStr, "duplicate") || strings.Contains(errStr, "unique") {
			b.SendWithCancel(message.Chat.ID, fmt.Sprintf("⚠️ @%s уже является тренером этой организации.", username))
		} else {
			b.SendWithCancel(message.Chat.ID, "❌ Ошибка при добавлении тренера.")
		}
		return
	}

	showManagerOrgMenu(b, message, orgID, orgName)
	b.SendMessageWithKeyboard(
		message.Chat.ID,
		fmt.Sprintf("✅ Тренер @%s добавлен в организацию *%s*\n\nКогда тренер напишет боту, он получит доступ.", username, bot.EscapeMarkdown(orgName)),
		bot.GetManagerMenuKeyboard(),
	)
}

// HandleListTrainers показывает список тренеров
func HandleListTrainers(b *bot.Bot, message *tgbotapi.Message) {
	state := b.GetState(message.From.ID)
	if state == nil {
		b.SendMessage(message.Chat.ID, "❌ Сначала выберите организацию.")
		return
	}

	orgID, okID := bot.GetStateInt64(state.Data, "org_id")
	orgName, okName := bot.GetStateString(state.Data, "org_name")
	if !okID || !okName {
		b.SendMessage(message.Chat.ID, "❌ Сначала выберите организацию.")
		return
	}

	ctx := context.Background()

	trainers, err := b.DB.GetOrganizationTrainers(ctx, orgID)
	if err != nil {
		log.Printf("Error getting trainers: %v", err)
		b.SendMessage(message.Chat.ID, "❌ Ошибка при получении списка тренеров.")
		return
	}

	if len(trainers) == 0 {
		b.SendMessage(message.Chat.ID, fmt.Sprintf("В организации *%s* пока нет тренеров.", orgName))
		return
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🏋️ *Тренеры организации %s:*\n\n", orgName))

	for i, t := range trainers {
		status := "✅"
		if !t.IsActive {
			status = "❌"
		}
		sb.WriteString(fmt.Sprintf("%d. @%s %s\n", i+1, t.Username, status))
	}

	sb.WriteString("\nДля удаления тренера отправьте его номер.")

	b.SendMessage(message.Chat.ID, sb.String())
	b.SetState(message.From.ID, "manager_removing_trainer", map[string]interface{}{
		"org_id":   orgID,
		"org_name": orgName,
		"trainers": trainers,
	})
}

// HandleRemoveTrainer удаляет тренера
func HandleRemoveTrainer(b *bot.Bot, message *tgbotapi.Message, idx int) {
	ctx := context.Background()
	state := b.GetState(message.From.ID)
	if state == nil {
		b.SendMessage(message.Chat.ID, "❌ Сначала выберите организацию.")
		return
	}

	trainers, ok := state.Data["trainers"].([]*models.OrganizationTrainer)
	if !ok || len(trainers) == 0 || idx < 1 || idx > len(trainers) {
		b.SendMessage(message.Chat.ID, "❌ Неверный номер или список устарел.")
		return
	}

	orgID, okID := bot.GetStateInt64(state.Data, "org_id")
	orgName, okName := bot.GetStateString(state.Data, "org_name")
	if !okID || !okName {
		b.SendMessage(message.Chat.ID, "❌ Ошибка состояния.")
		return
	}

	trainer := trainers[idx-1]
	if err := b.DB.RemoveTrainer(ctx, orgID, trainer.Username); err != nil {
		log.Printf("Error removing trainer: %v", err)
		b.SendMessage(message.Chat.ID, "❌ Ошибка при удалении тренера.")
		return
	}

	showManagerOrgMenu(b, message, orgID, orgName)
	b.SendMessageWithKeyboard(
		message.Chat.ID,
		fmt.Sprintf("✅ Тренер @%s удалён из организации *%s*\n\n⚠️ Его клиенты смогут просматривать историю тренировок.", trainer.Username, bot.EscapeMarkdown(orgName)),
		bot.GetManagerMenuKeyboard(),
	)
}
