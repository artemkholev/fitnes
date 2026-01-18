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

// HandleAdminMenu показывает админ-меню
func HandleAdminMenu(b *bot.Bot, message *tgbotapi.Message) {
	if !b.IsAdmin(message.From.UserName) {
		b.SendMessage(message.Chat.ID, "❌ У вас нет прав администратора.")
		return
	}

	b.SendMessageWithKeyboard(
		message.Chat.ID,
		"🔧 *Панель администратора*\n\nВыберите действие:",
		bot.GetAdminMenuKeyboard(),
	)
}

// HandleCreateOrganization начинает создание организации
func HandleCreateOrganization(b *bot.Bot, message *tgbotapi.Message) {
	if !b.IsAdmin(message.From.UserName) {
		b.SendMessage(message.Chat.ID, "❌ У вас нет прав администратора.")
		return
	}

	b.SendMessageWithKeyboard(
		message.Chat.ID,
		"Введите название новой организации:",
		bot.GetCancelKeyboard(),
	)
	b.SetState(message.From.ID, "admin_creating_org_name", nil)
}

// HandleCreateOrganizationName обрабатывает ввод названия организации
func HandleCreateOrganizationName(b *bot.Bot, message *tgbotapi.Message) {
	if message.Text == "❌ Отмена" {
		b.ClearState(message.From.ID)
		HandleAdminMenu(b, message)
		return
	}

	orgName := strings.TrimSpace(message.Text)
	if orgName == "" {
		b.SendMessage(message.Chat.ID, "Название не может быть пустым. Попробуйте ещё раз:")
		return
	}

	b.SetState(message.From.ID, "admin_creating_org_code", map[string]interface{}{
		"org_name": orgName,
	})
	b.SendMessage(message.Chat.ID, "Введите уникальный код организации (латиницей, без пробелов):")
}

// HandleCreateOrganizationCode обрабатывает ввод кода организации
func HandleCreateOrganizationCode(b *bot.Bot, message *tgbotapi.Message) {
	ctx := context.Background()
	state := b.GetState(message.From.ID)

	if message.Text == "❌ Отмена" {
		b.ClearState(message.From.ID)
		HandleAdminMenu(b, message)
		return
	}

	orgCode := strings.ToUpper(strings.TrimSpace(message.Text))
	orgName := state.Data["org_name"].(string)

	org := &models.Organization{
		Name: orgName,
		Code: orgCode,
	}

	if err := b.DB.CreateOrganization(ctx, org); err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			b.SendMessage(message.Chat.ID, "❌ Организация с таким кодом уже существует. Введите другой код:")
			return
		}
		log.Printf("Error creating organization: %v", err)
		b.SendMessage(message.Chat.ID, "❌ Ошибка при создании организации.")
		return
	}

	b.ClearState(message.From.ID)
	b.SendMessageWithKeyboard(
		message.Chat.ID,
		fmt.Sprintf("✅ Организация *%s* (код: `%s`) успешно создана!", orgName, orgCode),
		bot.GetAdminMenuKeyboard(),
	)
}

// HandleListOrganizations показывает список организаций
func HandleListOrganizations(b *bot.Bot, message *tgbotapi.Message) {
	if !b.IsAdmin(message.From.UserName) {
		b.SendMessage(message.Chat.ID, "❌ У вас нет прав администратора.")
		return
	}

	ctx := context.Background()
	orgs, err := b.DB.GetAllOrganizations(ctx)
	if err != nil {
		log.Printf("Error getting organizations: %v", err)
		b.SendMessage(message.Chat.ID, "❌ Ошибка при получении списка организаций.")
		return
	}

	if len(orgs) == 0 {
		b.SendMessage(message.Chat.ID, "Организаций пока нет. Создайте первую!")
		return
	}

	var sb strings.Builder
	sb.WriteString("📋 *Список организаций:*\n\n")

	for i, org := range orgs {
		sb.WriteString(fmt.Sprintf("%d. *%s* (код: `%s`)\n", i+1, org.Name, org.Code))
	}

	sb.WriteString("\nДля управления организацией отправьте её номер.")

	b.SendMessage(message.Chat.ID, sb.String())
	b.SetState(message.From.ID, "admin_selecting_org", map[string]interface{}{
		"organizations": orgs,
	})
}

// HandleSelectOrganization выбор организации для управления
func HandleSelectOrganization(b *bot.Bot, message *tgbotapi.Message, idx int) {
	state := b.GetState(message.From.ID)
	orgs := state.Data["organizations"].([]*models.Organization)

	if idx < 1 || idx > len(orgs) {
		b.SendMessage(message.Chat.ID, "❌ Неверный номер. Попробуйте ещё раз.")
		return
	}

	org := orgs[idx-1]
	b.SetState(message.From.ID, "admin_managing_org", map[string]interface{}{
		"org_id":   org.ID,
		"org_name": org.Name,
	})

	b.SendMessageWithKeyboard(
		message.Chat.ID,
		fmt.Sprintf("🏢 Управление организацией *%s*\n\nВыберите действие:", org.Name),
		bot.GetOrgManageKeyboard(),
	)
}

// HandleAddManager начинает добавление менеджера
func HandleAddManager(b *bot.Bot, message *tgbotapi.Message) {
	state := b.GetState(message.From.ID)
	if state == nil || state.Data["org_id"] == nil {
		b.SendMessage(message.Chat.ID, "❌ Сначала выберите организацию.")
		return
	}

	b.SetState(message.From.ID, "admin_adding_manager", state.Data)
	b.SendMessageWithKeyboard(
		message.Chat.ID,
		"Введите @username менеджера (например: @ArtKholev):",
		bot.GetCancelKeyboard(),
	)
}

// HandleAddManagerUsername обрабатывает ввод username менеджера
func HandleAddManagerUsername(b *bot.Bot, message *tgbotapi.Message) {
	ctx := context.Background()
	state := b.GetState(message.From.ID)

	if message.Text == "❌ Отмена" {
		b.SetState(message.From.ID, "admin_managing_org", state.Data)
		orgName := state.Data["org_name"].(string)
		b.SendMessageWithKeyboard(
			message.Chat.ID,
			fmt.Sprintf("🏢 Управление организацией *%s*", orgName),
			bot.GetOrgManageKeyboard(),
		)
		return
	}

	username := database.NormalizeUsername(message.Text)
	if username == "" {
		b.SendMessage(message.Chat.ID, "❌ Некорректный username. Введите в формате @username:")
		return
	}

	orgID := state.Data["org_id"].(int64)
	orgName := state.Data["org_name"].(string)

	if err := b.DB.AddManager(ctx, orgID, username); err != nil {
		log.Printf("Error adding manager: %v", err)
		b.SendMessage(message.Chat.ID, "❌ Ошибка при добавлении менеджера.")
		return
	}

	b.SetState(message.From.ID, "admin_managing_org", state.Data)
	b.SendMessageWithKeyboard(
		message.Chat.ID,
		fmt.Sprintf("✅ Менеджер @%s добавлен в организацию *%s*", username, orgName),
		bot.GetOrgManageKeyboard(),
	)
}

// HandleListManagers показывает список менеджеров организации
func HandleListManagers(b *bot.Bot, message *tgbotapi.Message) {
	state := b.GetState(message.From.ID)
	if state == nil || state.Data["org_id"] == nil {
		b.SendMessage(message.Chat.ID, "❌ Сначала выберите организацию.")
		return
	}

	ctx := context.Background()
	orgID := state.Data["org_id"].(int64)
	orgName := state.Data["org_name"].(string)

	managers, err := b.DB.GetOrganizationManagers(ctx, orgID)
	if err != nil {
		log.Printf("Error getting managers: %v", err)
		b.SendMessage(message.Chat.ID, "❌ Ошибка при получении списка менеджеров.")
		return
	}

	if len(managers) == 0 {
		b.SendMessage(message.Chat.ID, fmt.Sprintf("В организации *%s* пока нет менеджеров.", orgName))
		return
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("👥 *Менеджеры организации %s:*\n\n", orgName))

	for i, m := range managers {
		status := "✅"
		if !m.IsActive {
			status = "❌"
		}
		sb.WriteString(fmt.Sprintf("%d. @%s %s\n", i+1, m.Username, status))
	}

	sb.WriteString("\nДля удаления менеджера отправьте его номер.")

	b.SendMessage(message.Chat.ID, sb.String())
	b.SetState(message.From.ID, "admin_removing_manager", map[string]interface{}{
		"org_id":   orgID,
		"org_name": orgName,
		"managers": managers,
	})
}

// HandleRemoveManager удаляет менеджера
func HandleRemoveManager(b *bot.Bot, message *tgbotapi.Message, idx int) {
	ctx := context.Background()
	state := b.GetState(message.From.ID)
	managers := state.Data["managers"].([]*models.OrganizationManager)
	orgID := state.Data["org_id"].(int64)
	orgName := state.Data["org_name"].(string)

	if idx < 1 || idx > len(managers) {
		b.SendMessage(message.Chat.ID, "❌ Неверный номер.")
		return
	}

	manager := managers[idx-1]
	if err := b.DB.RemoveManager(ctx, orgID, manager.Username); err != nil {
		log.Printf("Error removing manager: %v", err)
		b.SendMessage(message.Chat.ID, "❌ Ошибка при удалении менеджера.")
		return
	}

	b.SetState(message.From.ID, "admin_managing_org", map[string]interface{}{
		"org_id":   orgID,
		"org_name": orgName,
	})
	b.SendMessageWithKeyboard(
		message.Chat.ID,
		fmt.Sprintf("✅ Менеджер @%s удалён из организации *%s*", manager.Username, orgName),
		bot.GetOrgManageKeyboard(),
	)
}
