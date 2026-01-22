package handlers

import (
	"context"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"fitness-bot/internal/bot"
	"fitness-bot/internal/database"
	"fitness-bot/internal/models"
)

// HandleAdminMenu — вход в админку
func HandleAdminMenu(b *bot.Bot, message *tgbotapi.Message) {
	if !b.IsAdmin(message.From.UserName) {
		b.SendError(message.Chat.ID, "У вас нет прав администратора.")
		return
	}

	b.SendMessageWithKeyboard(
		message.Chat.ID,
		"*Панель администратора*\n\nВыберите действие:",
		bot.GetAdminMenuKeyboard(),
	)
}

// HandleCreateOrganization — начало создания организации
func HandleCreateOrganization(b *bot.Bot, message *tgbotapi.Message) {
	if !b.IsAdmin(message.From.UserName) {
		b.SendError(message.Chat.ID, "У вас нет прав администратора.")
		return
	}

	b.SendWithCancel(message.Chat.ID, "Введите название новой организации:")
	b.SetState(message.From.ID, "admin_creating_org_name", nil)
}

// HandleCreateOrganizationName — ввод названия
func HandleCreateOrganizationName(b *bot.Bot, message *tgbotapi.Message) {
	if message.Text == "❌ Отмена" {
		b.ClearState(message.From.ID)
		HandleAdminMenu(b, message)
		return
	}

	orgName := strings.TrimSpace(message.Text)
	if orgName == "" {
		b.SendWithCancel(message.Chat.ID, "Название не может быть пустым. Попробуйте ещё раз:")
		return
	}

	b.SetState(message.From.ID, "admin_creating_org_code", map[string]interface{}{
		"org_name": orgName,
	})
	b.SendWithCancel(message.Chat.ID, "Введите уникальный код организации (латиницей, без пробелов):")
}

// HandleCreateOrganizationCode — ввод кода
func HandleCreateOrganizationCode(b *bot.Bot, message *tgbotapi.Message) {
	ctx := context.Background()
	state := b.GetState(message.From.ID)

	if message.Text == "❌ Отмена" {
		b.ClearState(message.From.ID)
		HandleAdminMenu(b, message)
		return
	}

	orgCode := strings.ToUpper(strings.TrimSpace(message.Text))
	if orgCode == "" {
		b.SendWithCancel(message.Chat.ID, "Код не может быть пустым. Введите заново:")
		return
	}

	orgName, ok := bot.GetStateString(state.Data, "org_name")
	if !ok {
		b.ClearState(message.From.ID)
		b.SendError(message.Chat.ID, "Ошибка состояния. Попробуйте снова.")
		HandleAdminMenu(b, message)
		return
	}
	escapedName := bot.EscapeMarkdown(orgName)

	org := &models.Organization{
		Name: orgName,
		Code: orgCode,
	}

	if err := b.DB.CreateOrganization(ctx, org); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") || strings.Contains(err.Error(), "unique") {
			b.SendWithCancel(message.Chat.ID, "Организация с таким кодом уже существует. Введите другой код:")
			return
		}
		log.Printf("Error creating organization (admin: %s): %v", message.From.UserName, err)
		b.SendError(message.Chat.ID, "Ошибка при создания организации.")
		return
	}

	b.ClearState(message.From.ID)
	b.SendSuccess(message.Chat.ID,
		fmt.Sprintf("Организация *%s* (код: `%s`) успешно создана!", escapedName, orgCode),
		bot.GetAdminMenuKeyboard())
}

// HandleListOrganizations — список организаций
func HandleListOrganizations(b *bot.Bot, message *tgbotapi.Message) {
	if !b.IsAdmin(message.From.UserName) {
		b.SendError(message.Chat.ID, "У вас нет прав администратора.")
		return
	}

	ctx := context.Background()
	orgs, err := b.DB.GetAllOrganizations(ctx)
	if err != nil {
		log.Printf("Error getting organizations: %v", err)
		b.SendError(message.Chat.ID, "Ошибка при получении списка организаций.")
		return
	}

	if len(orgs) == 0 {
		b.SendWithCancel(message.Chat.ID, "Организаций пока нет. Создайте первую!")
		return
	}

	var sb strings.Builder
	sb.WriteString("*Список организаций:*\n\n")

	for i, org := range orgs {
		sb.WriteString(fmt.Sprintf("%d. *%s* (код: `%s`)\n", i+1, bot.EscapeMarkdown(org.Name), org.Code))
	}
	sb.WriteString("\nОтправьте номер организации для управления или нажмите «Отмена».")

	b.SendWithCancel(message.Chat.ID, sb.String())

	b.SetState(message.From.ID, "admin_selecting_org", map[string]interface{}{
		"organizations": orgs,
	})
}

// HandleSelectOrganization — выбор организации
func HandleSelectOrganization(b *bot.Bot, message *tgbotapi.Message, idx int) {
	state := b.GetState(message.From.ID)
	orgs, ok := state.Data["organizations"].([]*models.Organization)

	if !ok || len(orgs) == 0 || idx < 1 || idx > len(orgs) {
		b.SendWithCancel(message.Chat.ID, "Неверный номер или список устарел. Запросите список заново.")
		return
	}

	org := orgs[idx-1]
	b.SetState(message.From.ID, "admin_managing_org", map[string]interface{}{
		"org_id":   org.ID,
		"org_name": org.Name,
	})

	b.SendMessageWithKeyboard(
		message.Chat.ID,
		fmt.Sprintf("Управление организацией *%s*\n\nВыберите действие:", bot.EscapeMarkdown(org.Name)),
		bot.GetOrgManageKeyboard(),
	)
}

// HandleAddManager — начало добавления менеджера
func HandleAddManager(b *bot.Bot, message *tgbotapi.Message) {
	state := b.GetState(message.From.ID)
	if state == nil {
		b.SendError(message.Chat.ID, "Сначала выберите организацию.")
		return
	}

	_, okID := bot.GetStateInt64(state.Data, "org_id")
	_, okName := bot.GetStateString(state.Data, "org_name")
	if !okID || !okName {
		b.SendError(message.Chat.ID, "Сначала выберите организацию.")
		return
	}

	// Копируем состояние — критически важно!
	b.SetState(message.From.ID, "admin_adding_manager", bot.CopyStateData(state.Data))

	b.SendWithCancel(message.Chat.ID, "Введите @username менеджера (например: @ivan\\_manager):")
}

// HandleAddManagerUsername — обработка ввода username
func HandleAddManagerUsername(b *bot.Bot, message *tgbotapi.Message) {
	ctx := context.Background()
	state := b.GetState(message.From.ID)

	// Безопасное извлечение данных
	orgID, okID := bot.GetStateInt64(state.Data, "org_id")
	orgName, okName := bot.GetStateString(state.Data, "org_name")
	if !okID || !okName {
		b.ClearState(message.From.ID)
		b.SendError(message.Chat.ID, "Ошибка состояния. Попробуйте снова.")
		HandleAdminMenu(b, message)
		return
	}

	if message.Text == "❌ Отмена" {
		b.SetState(message.From.ID, "admin_managing_org", bot.CopyStateData(state.Data))
		b.SendMessageWithKeyboard(
			message.Chat.ID,
			fmt.Sprintf("Управление организацией *%s*", bot.EscapeMarkdown(orgName)),
			bot.GetOrgManageKeyboard(),
		)
		return
	}

	username := database.NormalizeUsername(message.Text)
	if username == "" {
		b.SendWithCancel(message.Chat.ID, "Некорректный username. Введите в формате @username:")
		return
	}

	if err := b.DB.AddManager(ctx, orgID, username); err != nil {
		log.Printf("Error adding manager @%s (admin: %s): %v", username, message.From.UserName, err)

		// Более понятные ошибки
		errStr := err.Error()
		if strings.Contains(errStr, "not found") {
			b.SendWithCancel(message.Chat.ID, fmt.Sprintf("Пользователь @%s не найден в системе. Пусть сначала запустит бота.", username))
		} else if strings.Contains(errStr, "already") {
			b.SendWithCancel(message.Chat.ID, fmt.Sprintf("Пользователь @%s уже является менеджером этой организации.", username))
		} else {
			b.SendError(message.Chat.ID, "Ошибка при добавлении менеджера.")
		}
		return
	}

	b.SetState(message.From.ID, "admin_managing_org", map[string]interface{}{
		"org_id":   orgID,
		"org_name": orgName,
	})

	b.SendSuccess(message.Chat.ID,
		fmt.Sprintf("Менеджер @%s добавлен в организацию *%s*", username, bot.EscapeMarkdown(orgName)),
		bot.GetOrgManageKeyboard())
}

// HandleListManagers — список менеджеров
func HandleListManagers(b *bot.Bot, message *tgbotapi.Message) {
	state := b.GetState(message.From.ID)
	if state == nil {
		b.SendError(message.Chat.ID, "Сначала выберите организацию.")
		return
	}

	orgID, okID := bot.GetStateInt64(state.Data, "org_id")
	orgName, okName := bot.GetStateString(state.Data, "org_name")
	if !okID || !okName {
		b.SendError(message.Chat.ID, "Сначала выберите организацию.")
		return
	}

	ctx := context.Background()

	managers, err := b.DB.GetOrganizationManagers(ctx, orgID)
	if err != nil {
		log.Printf("Error getting managers (org %d): %v", orgID, err)
		b.SendError(message.Chat.ID, "Ошибка при получении списка менеджеров.")
		return
	}

	if len(managers) == 0 {
		b.SendWithCancel(message.Chat.ID, fmt.Sprintf("В организации *%s* пока нет менеджеров.", bot.EscapeMarkdown(orgName)))
		return
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("*Менеджеры организации %s:*\n\n", bot.EscapeMarkdown(orgName)))

	for i, m := range managers {
		status := "неактивен"
		if m.IsActive {
			status = "активен"
		}
		sb.WriteString(fmt.Sprintf("%d. @%s — %s\n", i+1, m.Username, status))
	}
	sb.WriteString("\nОтправьте номер для удаления или нажмите «Отмена».")

	b.SendWithCancel(message.Chat.ID, sb.String())

	newData := bot.CopyStateData(state.Data)
	newData["managers"] = managers
	b.SetState(message.From.ID, "admin_removing_manager", newData)
}

// HandleRemoveManager — удаление менеджера
func HandleRemoveManager(b *bot.Bot, message *tgbotapi.Message, idx int) {
	state := b.GetState(message.From.ID)
	if state == nil {
		b.SendError(message.Chat.ID, "Сначала выберите организацию.")
		return
	}

	managers, ok := state.Data["managers"].([]*models.OrganizationManager)
	if !ok || len(managers) == 0 || idx < 1 || idx > len(managers) {
		b.SendWithCancel(message.Chat.ID, "Неверный номер или список устарел. Запросите список заново.")
		return
	}

	orgID, okID := bot.GetStateInt64(state.Data, "org_id")
	orgName, okName := bot.GetStateString(state.Data, "org_name")
	if !okID || !okName {
		b.SendError(message.Chat.ID, "Ошибка состояния. Попробуйте снова.")
		return
	}
	manager := managers[idx-1]

	ctx := context.Background()
	if err := b.DB.RemoveManager(ctx, orgID, manager.Username); err != nil {
		log.Printf("Error removing manager @%s: %v", manager.Username, err)
		b.SendError(message.Chat.ID, "Ошибка при удалении менеджера.")
		return
	}

	b.SetState(message.From.ID, "admin_managing_org", map[string]interface{}{
		"org_id":   orgID,
		"org_name": orgName,
	})

	b.SendSuccess(message.Chat.ID,
		fmt.Sprintf("Менеджер @%s удалён из организации *%s*", manager.Username, bot.EscapeMarkdown(orgName)),
		bot.GetOrgManageKeyboard())
}
