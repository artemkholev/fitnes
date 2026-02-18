package handlers

import (
	"fitness-bot/internal/bot"
	"fitness-bot/internal/database"
	"fitness-bot/internal/models"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleTrainerMenu показывает меню тренера
func HandleTrainerMenu(b *bot.Bot, message *tgbotapi.Message, trainerOrgs []*models.TrainerOrgInfo) {
	if len(trainerOrgs) == 0 {
		b.SendMessage(message.Chat.ID, "❌ У вас нет доступа тренера ни к одной организации.")
		return
	}

	// Фильтруем активные организации
	activeOrgs := []*models.TrainerOrgInfo{}
	for _, org := range trainerOrgs {
		if org.IsActive {
			activeOrgs = append(activeOrgs, org)
		}
	}

	if len(activeOrgs) == 0 {
		b.SendMessage(message.Chat.ID, "❌ Все ваши доступы к организациям были деактивированы.")
		return
	}

	// Если одна организация - сразу показываем управление
	if len(activeOrgs) == 1 {
		org := activeOrgs[0]
		showTrainerOrgMenu(b, message, org.TrainerID, org.Organization.ID, org.Organization.Name)
		return
	}

	// Несколько организаций - показываем выбор
	var sb strings.Builder
	sb.WriteString("🏢 *Выберите организацию:*\n\n")

	for i, org := range activeOrgs {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, org.Organization.Name))
	}

	b.SendMessage(message.Chat.ID, sb.String())
	b.SetState(message.From.ID, "trainer_selecting_org", map[string]interface{}{
		"organizations": activeOrgs,
	})
}

func showTrainerOrgMenu(b *bot.Bot, message *tgbotapi.Message, trainerID, orgID int64, orgName string) {
	// Очищаем старые сообщения
	b.CleanupMessages(message.Chat.ID, message.From.ID)

	b.SetState(message.From.ID, "trainer_managing_org", map[string]interface{}{
		"trainer_id": trainerID,
		"org_id":     orgID,
		"org_name":   orgName,
	})

	breadcrumbs := bot.GetBreadcrumbs("🏠 Главная", "🏋️ Тренер", orgName)
	text := breadcrumbs + "Выберите действие:"

	b.SendMessageWithKeyboard(
		message.Chat.ID,
		text,
		bot.GetTrainerMenuKeyboard(),
	)
}

// HandleTrainerSelectOrg выбор организации тренером
func HandleTrainerSelectOrg(b *bot.Bot, message *tgbotapi.Message, idx int) {
	state := b.GetState(message.From.ID)
	if state == nil {
		b.SendMessage(message.Chat.ID, "❌ Список организаций устарел. Попробуйте снова.")
		return
	}

	orgs, ok := state.Data["organizations"].([]*models.TrainerOrgInfo)
	if !ok || len(orgs) == 0 || idx < 1 || idx > len(orgs) {
		b.SendMessage(message.Chat.ID, "❌ Неверный номер или список устарел.")
		return
	}

	org := orgs[idx-1]
	showTrainerOrgMenu(b, message, org.TrainerID, org.Organization.ID, org.Organization.Name)
}

// HandleAddClient начинает добавление клиента
func HandleAddClient(b *bot.Bot, message *tgbotapi.Message) {
	state := b.GetState(message.From.ID)
	if state == nil {
		b.SendMessage(message.Chat.ID, "❌ Сначала выберите организацию.")
		return
	}

	_, okT := bot.GetStateInt64(state.Data, "trainer_id")
	_, okID := bot.GetStateInt64(state.Data, "org_id")
	_, okName := bot.GetStateString(state.Data, "org_name")
	if !okT || !okID || !okName {
		b.SendMessage(message.Chat.ID, "❌ Сначала выберите организацию.")
		return
	}

	b.SetState(message.From.ID, "trainer_adding_client", bot.CopyStateData(state.Data))
	b.SendMessageWithKeyboard(
		message.Chat.ID,
		"Введите @username клиента (например: @client\\_ivan):",
		bot.GetCancelKeyboard(),
	)
}

// HandleAddClientUsername обрабатывает ввод username клиента
func HandleAddClientUsername(b *bot.Bot, message *tgbotapi.Message) {
	state := b.GetState(message.From.ID)

	// Безопасное извлечение данных
	trainerID, okT := bot.GetStateInt64(state.Data, "trainer_id")
	orgID, okID := bot.GetStateInt64(state.Data, "org_id")
	orgName, okName := bot.GetStateString(state.Data, "org_name")
	if !okT || !okID || !okName {
		b.ClearState(message.From.ID)
		b.SendMessage(message.Chat.ID, "❌ Ошибка состояния. Попробуйте снова.")
		return
	}

	if message.Text == "❌ Отмена" {
		showTrainerOrgMenu(b, message, trainerID, orgID, orgName)
		return
	}

	username := database.NormalizeUsername(message.Text)
	if username == "" {
		b.SendWithCancel(message.Chat.ID, "❌ Некорректный username. Введите в формате @username:")
		return
	}

	if err := b.DB.AddClient( trainerID, username); err != nil {
		log.Printf("Error adding client: %v", err)
		errStr := err.Error()
		if strings.Contains(errStr, "duplicate") || strings.Contains(errStr, "unique") {
			b.SendWithCancel(message.Chat.ID, fmt.Sprintf("⚠️ @%s уже ваш клиент.", username))
		} else {
			b.SendWithCancel(message.Chat.ID, "❌ Ошибка при добавлении клиента.")
		}
		return
	}

	showTrainerOrgMenu(b, message, trainerID, orgID, orgName)
	b.SendMessageWithKeyboard(
		message.Chat.ID,
		fmt.Sprintf("✅ Клиент @%s добавлен.\n\nКогда клиент напишет боту, он получит доступ к тренировкам.", username),
		bot.GetTrainerMenuKeyboard(),
	)
}

// HandleListClients показывает список клиентов тренера
func HandleListClients(b *bot.Bot, message *tgbotapi.Message) {
	state := b.GetState(message.From.ID)
	if state == nil {
		b.SendMessage(message.Chat.ID, "❌ Сначала выберите организацию.")
		return
	}

	trainerID, okT := bot.GetStateInt64(state.Data, "trainer_id")
	orgName, okName := bot.GetStateString(state.Data, "org_name")
	if !okT || !okName {
		b.SendMessage(message.Chat.ID, "❌ Сначала выберите организацию.")
		return
	}


	clients, err := b.DB.GetTrainerClients( trainerID)
	if err != nil {
		log.Printf("Error getting clients: %v", err)
		b.SendMessage(message.Chat.ID, "❌ Ошибка при получении списка клиентов.")
		return
	}

	if len(clients) == 0 {
		b.SendMessage(message.Chat.ID, fmt.Sprintf("У вас пока нет клиентов в организации *%s*.", orgName))
		return
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("👥 *Ваши клиенты в %s:*\n\n", orgName))

	// Создаем inline-кнопки для каждого клиента
	var items []string
	var ids []int64
	for _, c := range clients {
		status := "✅"
		if !c.Client.IsActive {
			status = "❌"
		}

		name := c.Client.Username
		if c.FullName != "" {
			name = c.FullName
		}

		workoutInfo := ""
		if c.WorkoutCount > 0 {
			workoutInfo = fmt.Sprintf(" | %d 🏋️", c.WorkoutCount)
		}

		items = append(items, fmt.Sprintf("%s %s%s", name, status, workoutInfo))
		ids = append(ids, c.Client.ID)
	}

	sb.WriteString("Выберите клиента для просмотра:")

	keyboard := bot.GetInlineListKeyboard(items, ids, "client")
	b.SendInlineKeyboard(message.Chat.ID, sb.String(), keyboard)

	b.SetState(message.From.ID, "trainer_viewing_clients", map[string]interface{}{
		"trainer_id": state.Data["trainer_id"],
		"org_id":     state.Data["org_id"],
		"org_name":   state.Data["org_name"],
		"clients":    clients,
	})
}

// HandleSelectClient выбор клиента для просмотра
func HandleSelectClient(b *bot.Bot, message *tgbotapi.Message, idx int) {
	state := b.GetState(message.From.ID)
	if state == nil {
		b.SendMessage(message.Chat.ID, "❌ Список клиентов устарел. Попробуйте снова.")
		return
	}

	clients, ok := state.Data["clients"].([]*models.ClientWithInfo)
	if !ok || len(clients) == 0 || idx < 1 || idx > len(clients) {
		b.SendMessage(message.Chat.ID, "❌ Неверный номер или список устарел.")
		return
	}

	client := clients[idx-1]

	// Очищаем старые сообщения
	b.CleanupMessages(message.Chat.ID, message.From.ID)

	var sb strings.Builder
	name := client.Client.Username
	if client.FullName != "" {
		name = client.FullName
	}

	// Breadcrumbs
	orgName, _ := bot.GetStateString(state.Data, "org_name")
	breadcrumbs := bot.GetBreadcrumbs("🏠 Главная", "🏋️ Тренер", orgName, "👥 Клиенты", name)
	sb.WriteString(breadcrumbs)

	sb.WriteString(fmt.Sprintf("Username: @%s\n", client.Client.Username))
	sb.WriteString(fmt.Sprintf("Тренировок: %d\n", client.WorkoutCount))
	if client.LastWorkout != nil {
		sb.WriteString(fmt.Sprintf("Последняя: %s\n", client.LastWorkout.Format("02.01.2006")))
	}

	status := "Активен ✅"
	if !client.Client.IsActive {
		status = "Деактивирован ❌"
	}
	sb.WriteString(fmt.Sprintf("Статус: %s\n", status))

	sb.WriteString("\n*Действия:*\n")
	sb.WriteString("1. 📊 Статистика клиента\n")
	sb.WriteString("2. ➕ Создать тренировку\n")
	sb.WriteString("3. 📋 История тренировок\n")
	if client.Client.IsActive {
		sb.WriteString("4. ❌ Удалить клиента\n")
	}

	b.SendMessage(message.Chat.ID, sb.String())
	b.SetState(message.From.ID, "trainer_client_action", map[string]interface{}{
		"trainer_id": state.Data["trainer_id"],
		"org_id":     state.Data["org_id"],
		"org_name":   state.Data["org_name"],
		"client":     client,
	})
}

// HandleRemoveClientByIndex удаляет клиента по индексу из списка
func HandleRemoveClientByIndex(b *bot.Bot, message *tgbotapi.Message, idx int) {
	state := b.GetState(message.From.ID)
	if state == nil {
		b.SendMessage(message.Chat.ID, "❌ Список клиентов устарел.")
		return
	}

	clients, ok := state.Data["clients"].([]*models.ClientWithInfo)
	if !ok || len(clients) == 0 || idx < 1 || idx > len(clients) {
		b.SendMessage(message.Chat.ID, "❌ Неверный номер или список устарел.")
		return
	}

	trainerID, okT := bot.GetStateInt64(state.Data, "trainer_id")
	orgID, okID := bot.GetStateInt64(state.Data, "org_id")
	orgName, okName := bot.GetStateString(state.Data, "org_name")
	if !okT || !okID || !okName {
		b.SendMessage(message.Chat.ID, "❌ Ошибка состояния.")
		return
	}

	client := clients[idx-1]
	if err := b.DB.RemoveClient( trainerID, client.Client.Username); err != nil {
		log.Printf("Error removing client: %v", err)
		b.SendMessage(message.Chat.ID, "❌ Ошибка при удалении клиента.")
		return
	}

	showTrainerOrgMenu(b, message, trainerID, orgID, orgName)
	b.SendMessageWithKeyboard(
		message.Chat.ID,
		fmt.Sprintf("✅ Клиент @%s удалён.\n\n⚠️ Клиент сможет просматривать историю тренировок.", client.Client.Username),
		bot.GetTrainerMenuKeyboard(),
	)
}

// HandleClientAction обрабатывает действие с клиентом
func HandleClientAction(b *bot.Bot, message *tgbotapi.Message, action int) {
	state := b.GetState(message.From.ID)
	if state == nil {
		b.SendMessage(message.Chat.ID, "❌ Ошибка состояния. Попробуйте снова.")
		return
	}

	client, ok := state.Data["client"].(*models.ClientWithInfo)
	if !ok || client == nil {
		b.SendMessage(message.Chat.ID, "❌ Клиент не выбран.")
		return
	}

	trainerID, okT := bot.GetStateInt64(state.Data, "trainer_id")
	orgID, okID := bot.GetStateInt64(state.Data, "org_id")
	orgName, okName := bot.GetStateString(state.Data, "org_name")
	if !okT || !okID || !okName {
		b.SendMessage(message.Chat.ID, "❌ Ошибка состояния.")
		return
	}

	switch action {
	case 1: // Статистика
		b.SendMessage(message.Chat.ID, fmt.Sprintf("📊 Статистика клиента @%s будет добавлена позже.", client.Client.Username))

	case 2: // Создать тренировку
		b.SetState(message.From.ID, "awaiting_muscle_group", map[string]interface{}{
			"trainer_id":         trainerID,
			"org_id":             orgID,
			"org_name":           orgName,
			"client":             client,
			"trainer_client_id":  client.Client.ID,
			"client_telegram_id": client.Client.TelegramID, // *int64, может быть nil если клиент не запустил бота
		})
		keyboard := bot.GetInlineMuscleGroupKeyboard()
		msgID := b.SendInlineKeyboard(
			message.Chat.ID,
			fmt.Sprintf("➕ *Создание тренировки для @%s*\n\nВыберите группу мышц:", client.Client.Username),
			keyboard,
		)
		b.StoreMessageID(message.From.ID, msgID)

	case 3: // История тренировок
		b.SendMessage(message.Chat.ID, fmt.Sprintf("📋 История тренировок @%s будет добавлена позже.", client.Client.Username))

	case 4: // Удалить клиента
		if !client.Client.IsActive {
			b.SendMessage(message.Chat.ID, "❌ Клиент уже деактивирован.")
			return
		}
		if err := b.DB.RemoveClient( trainerID, client.Client.Username); err != nil {
			log.Printf("Error removing client: %v", err)
			b.SendMessage(message.Chat.ID, "❌ Ошибка при удалении клиента.")
			return
		}
		showTrainerOrgMenu(b, message, trainerID, orgID, orgName)
		b.SendMessageWithKeyboard(
			message.Chat.ID,
			fmt.Sprintf("✅ Клиент @%s удалён.", client.Client.Username),
			bot.GetTrainerMenuKeyboard(),
		)

	default:
		b.SendMessage(message.Chat.ID, "❌ Неверный номер действия.")
	}
}
