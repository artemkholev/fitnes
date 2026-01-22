package main

import (
	"context"
	"fitness-bot/internal/bot"
	"fitness-bot/internal/database"
	"fitness-bot/internal/handlers"
	"fitness-bot/internal/models"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Применяем миграции
	dbURL := database.GetDatabaseURL(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)
	if err := database.RunMigrations(dbURL); err != nil {
		log.Printf("Warning: migrations failed: %v", err)
	}

	ctx := context.Background()

	db, err := database.NewDB(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Database connected successfully")

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is required")
	}

	adminUsername := os.Getenv("ADMIN_USERNAME")
	if adminUsername == "" {
		log.Fatal("ADMIN_USERNAME is required")
	}

	b, err := bot.NewBot(botToken, db, adminUsername)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.API.GetUpdatesChan(u)

	log.Println("Bot started successfully!")

	for update := range updates {
		if update.CallbackQuery != nil {
			go safeHandleCallback(b, update.CallbackQuery)
			continue
		}

		if update.Message == nil {
			continue
		}

		go safeHandleUpdate(b, update.Message)
	}
}

// safeHandleUpdate оборачивает handleUpdate с recover для защиты от panic
func safeHandleUpdate(b *bot.Bot, message *tgbotapi.Message) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("PANIC recovered: %v (user: %d, text: %s)", r, message.From.ID, message.Text)
			b.SendMessage(message.Chat.ID, "❌ Произошла ошибка. Попробуйте /start")
			b.ClearState(message.From.ID)
		}
	}()
	handleUpdate(b, message)
}

// safeHandleCallback оборачивает handleCallback с recover
func safeHandleCallback(b *bot.Bot, callback *tgbotapi.CallbackQuery) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("PANIC in callback: %v (user: %d, data: %s)", r, callback.From.ID, callback.Data)
			b.AnswerCallback(callback.ID, "Произошла ошибка")
			b.ClearState(callback.From.ID)
		}
	}()
	handleCallback(b, callback)
}

// handleCallback обрабатывает нажатия на inline-кнопки
func handleCallback(b *bot.Bot, callback *tgbotapi.CallbackQuery) {
	ctx := context.Background()

	// Отвечаем на callback чтобы убрать "часики"
	b.AnswerCallback(callback.ID, "")

	// Парсим callback data
	prefix, id, action := bot.ParseCallbackData(callback.Data)

	// Получаем информацию о доступах
	username := callback.From.UserName
	accessInfo, err := b.DB.GetUserAccessInfo(ctx, callback.From.ID, username)
	if err != nil {
		log.Printf("Error getting access info in callback: %v", err)
		return
	}
	accessInfo.IsAdmin = b.IsAdmin(username)

	chatID := callback.Message.Chat.ID
	messageID := callback.Message.MessageID

	switch prefix {
	case "org":
		handleOrgCallback(b, callback, id, action, accessInfo, chatID, messageID)
	case "muscle":
		handleMuscleCallback(b, callback, action, chatID, messageID)
	case "client":
		handleClientListCallback(b, callback, id, action, chatID, messageID)
	case "client_action":
		handleClientActionCallback(b, callback, id, action, chatID, messageID)
	case "manager":
		handleManagerListCallback(b, callback, id, action, chatID, messageID)
	case "trainer":
		handleTrainerListCallback(b, callback, id, action, chatID, messageID)
	case "exercise":
		handleExerciseCallback(b, callback, action, accessInfo, chatID, messageID)
	default:
		log.Printf("Unknown callback prefix: %s", prefix)
	}
}

// handleOrgCallback обрабатывает выбор организации
func handleOrgCallback(b *bot.Bot, callback *tgbotapi.CallbackQuery, id int64, action string, accessInfo *models.AccessInfo, chatID int64, messageID int) {
	if action == "cancel" {
		b.CleanupMessages(chatID, callback.From.ID)
		b.ClearState(callback.From.ID)
		return
	}

	state := b.GetState(callback.From.ID)
	if state == nil {
		return
	}

	// Находим организацию по ID
	switch state.State {
	case "admin_selecting_org":
		orgs, ok := state.Data["organizations"].([]*models.Organization)
		if !ok {
			return
		}
		for _, org := range orgs {
			if org.ID == id {
				// Очищаем предыдущие сообщения
				b.CleanupMessages(chatID, callback.From.ID)
				b.SetState(callback.From.ID, "admin_managing_org", map[string]interface{}{
					"org_id":   org.ID,
					"org_name": org.Name,
				})
				b.SendMessageWithKeyboard(
					chatID,
					"Управление организацией *"+bot.EscapeMarkdown(org.Name)+"*\n\nВыберите действие:",
					bot.GetOrgManageKeyboard(),
				)
				return
			}
		}
	}
}

// handleMuscleCallback обрабатывает выбор группы мышц
func handleMuscleCallback(b *bot.Bot, callback *tgbotapi.CallbackQuery, action string, chatID int64, messageID int) {
	ctx := context.Background()

	if action == "cancel" {
		b.CleanupMessages(chatID, callback.From.ID)
		b.ClearState(callback.From.ID)
		accessInfo, _ := b.DB.GetUserAccessInfo(ctx, callback.From.ID, callback.From.UserName)
		accessInfo.IsAdmin = b.IsAdmin(callback.From.UserName)
		b.SendMessageWithKeyboard(chatID, "Отменено.", bot.GetStartMenuKeyboard(accessInfo))
		return
	}

	state := b.GetState(callback.From.ID)
	if state == nil {
		return
	}

	muscleMap := map[string]models.MuscleGroup{
		"chest":     models.MuscleChest,
		"back":      models.MuscleBack,
		"legs":      models.MuscleLegs,
		"shoulders": models.MuscleShoulders,
		"biceps":    models.MuscleBiceps,
		"triceps":   models.MuscleTriceps,
		"abs":       models.MuscleAbs,
		"cardio":    models.MuscleCardio,
	}

	muscleGroup, ok := muscleMap[action]
	if !ok {
		return
	}

	// Извлекаем trainer_client_id если есть
	var trainerClientID *int64
	if state.Data != nil {
		if tcID, ok := bot.GetStateInt64(state.Data, "trainer_client_id"); ok && tcID > 0 {
			trainerClientID = &tcID
		}
	}

	workout := &models.Workout{
		TrainerClientID:  trainerClientID,
		ClientTelegramID: callback.From.ID,
		Date:             time.Now(),
		MuscleGroup:      muscleGroup,
	}

	if err := b.DB.CreateWorkout(ctx, workout); err != nil {
		log.Printf("Error creating workout: %v", err)
		b.EditMessageText(chatID, messageID, "❌ Ошибка при создании тренировки.", nil)
		return
	}

	b.SetState(callback.From.ID, "adding_exercises", map[string]interface{}{
		"workout_id":  workout.ID,
		"telegram_id": callback.From.ID,
		"order":       1,
	})

	// Очищаем предыдущие сообщения
	b.CleanupMessages(chatID, callback.From.ID)

	msgID := b.SendInlineKeyboard(
		chatID,
		"✅ Тренировка создана!\n\n*Добавьте упражнение:*\nОтправьте данные в формате:\n```\nНазвание\nПодходы\nПовторения\nВес (кг)\n```\n\nПример:\n```\nЖим лежа\n4\n10\n80\n```",
		bot.GetInlineFinishKeyboard(),
	)
	// Сохраняем ID нового сообщения
	b.StoreMessageID(callback.From.ID, msgID)
}

// handleClientListCallback обрабатывает выбор клиента из списка
func handleClientListCallback(b *bot.Bot, callback *tgbotapi.CallbackQuery, id int64, action string, chatID int64, messageID int) {
	if action == "cancel" {
		b.CleanupMessages(chatID, callback.From.ID)
		state := b.GetState(callback.From.ID)
		if state != nil {
			trainerID, okT := bot.GetStateInt64(state.Data, "trainer_id")
			orgID, okID := bot.GetStateInt64(state.Data, "org_id")
			orgName, okName := bot.GetStateString(state.Data, "org_name")
			if okT && okID && okName {
				b.SetState(callback.From.ID, "trainer_managing_org", map[string]interface{}{
					"trainer_id": trainerID,
					"org_id":     orgID,
					"org_name":   orgName,
				})
				b.SendMessageWithKeyboard(chatID, "🏋️ *Панель тренера - "+bot.EscapeMarkdown(orgName)+"*", bot.GetTrainerMenuKeyboard())
				return
			}
		}
		b.ClearState(callback.From.ID)
		return
	}

	state := b.GetState(callback.From.ID)
	if state == nil {
		return
	}

	clients, ok := state.Data["clients"].([]*models.ClientWithInfo)
	if !ok {
		return
	}

	// Находим клиента по ID
	for _, client := range clients {
		if client.Client.ID == id {
			// Показываем информацию о клиенте
			var sb strings.Builder
			name := client.Client.Username
			if client.FullName != "" {
				name = client.FullName
			}

			sb.WriteString("👤 *Клиент: " + bot.EscapeMarkdown(name) + "*\n")
			sb.WriteString("Username: @" + client.Client.Username + "\n")
			sb.WriteString("Тренировок: " + strconv.Itoa(client.WorkoutCount) + "\n")
			if client.LastWorkout != nil {
				sb.WriteString("Последняя: " + client.LastWorkout.Format("02.01.2006") + "\n")
			}

			status := "Активен ✅"
			if !client.Client.IsActive {
				status = "Деактивирован ❌"
			}
			sb.WriteString("Статус: " + status)

			b.SetState(callback.From.ID, "trainer_client_action", map[string]interface{}{
				"trainer_id": state.Data["trainer_id"],
				"org_id":     state.Data["org_id"],
				"org_name":   state.Data["org_name"],
				"client":     client,
			})

			keyboard := bot.GetInlineClientActionsKeyboard(client.Client.ID, client.Client.IsActive)
			b.EditMessageText(chatID, messageID, sb.String(), &keyboard)
			return
		}
	}
}

// handleClientActionCallback обрабатывает действия с клиентом
func handleClientActionCallback(b *bot.Bot, callback *tgbotapi.CallbackQuery, id int64, action string, chatID int64, messageID int) {
	ctx := context.Background()
	state := b.GetState(callback.From.ID)
	if state == nil {
		return
	}

	if action == "back" {
		// Возвращаемся к списку клиентов
		b.CleanupMessages(chatID, callback.From.ID)
		handlers.HandleListClients(b, &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: chatID},
			From: callback.From,
		})
		return
	}

	client, ok := state.Data["client"].(*models.ClientWithInfo)
	if !ok || client == nil {
		return
	}

	trainerID, okT := bot.GetStateInt64(state.Data, "trainer_id")
	orgID, okID := bot.GetStateInt64(state.Data, "org_id")
	orgName, okName := bot.GetStateString(state.Data, "org_name")
	if !okT || !okID || !okName {
		return
	}

	switch action {
	case "stats":
		b.EditMessageText(chatID, messageID, "📊 Статистика клиента @"+client.Client.Username+" будет добавлена позже.", nil)

	case "workout":
		b.CleanupMessages(chatID, callback.From.ID)
		b.SetState(callback.From.ID, "awaiting_muscle_group", map[string]interface{}{
			"trainer_id":        trainerID,
			"org_id":            orgID,
			"org_name":          orgName,
			"client":            client,
			"trainer_client_id": client.Client.ID,
			"telegram_id":       callback.From.ID,
		})
		keyboard := bot.GetInlineMuscleGroupKeyboard()
		msgID := b.SendInlineKeyboard(chatID, "➕ *Создание тренировки для @"+client.Client.Username+"*\n\nВыберите группу мышц:", keyboard)
		b.StoreMessageID(callback.From.ID, msgID)

	case "history":
		b.EditMessageText(chatID, messageID, "📋 История тренировок @"+client.Client.Username+" будет добавлена позже.", nil)

	case "delete":
		if !client.Client.IsActive {
			b.AnswerCallback(callback.ID, "Клиент уже деактивирован")
			return
		}
		if err := b.DB.RemoveClient(ctx, trainerID, client.Client.Username); err != nil {
			log.Printf("Error removing client: %v", err)
			b.AnswerCallback(callback.ID, "Ошибка удаления")
			return
		}
		b.CleanupMessages(chatID, callback.From.ID)
		b.SetState(callback.From.ID, "trainer_managing_org", map[string]interface{}{
			"trainer_id": trainerID,
			"org_id":     orgID,
			"org_name":   orgName,
		})
		b.SendMessageWithKeyboard(
			chatID,
			"✅ Клиент @"+client.Client.Username+" удалён.",
			bot.GetTrainerMenuKeyboard(),
		)
	}
}

// handleManagerListCallback обрабатывает выбор менеджера из списка
func handleManagerListCallback(b *bot.Bot, callback *tgbotapi.CallbackQuery, id int64, action string, chatID int64, messageID int) {
	ctx := context.Background()

	if action == "cancel" {
		b.CleanupMessages(chatID, callback.From.ID)
		state := b.GetState(callback.From.ID)
		if state != nil {
			orgID, okID := bot.GetStateInt64(state.Data, "org_id")
			orgName, okName := bot.GetStateString(state.Data, "org_name")
			if okID && okName {
				b.SetState(callback.From.ID, "admin_managing_org", map[string]interface{}{
					"org_id":   orgID,
					"org_name": orgName,
				})
				b.SendMessageWithKeyboard(chatID, "Управление организацией *"+bot.EscapeMarkdown(orgName)+"*", bot.GetOrgManageKeyboard())
				return
			}
		}
		b.ClearState(callback.From.ID)
		return
	}

	state := b.GetState(callback.From.ID)
	if state == nil {
		return
	}

	managers, ok := state.Data["managers"].([]*models.OrganizationManager)
	if !ok {
		return
	}

	orgID, okID := bot.GetStateInt64(state.Data, "org_id")
	orgName, okName := bot.GetStateString(state.Data, "org_name")
	if !okID || !okName {
		return
	}

	// Находим менеджера по ID
	for _, manager := range managers {
		if manager.ID == id {
			if err := b.DB.RemoveManager(ctx, orgID, manager.Username); err != nil {
				log.Printf("Error removing manager: %v", err)
				b.AnswerCallback(callback.ID, "Ошибка удаления")
				return
			}

			b.CleanupMessages(chatID, callback.From.ID)
			b.SetState(callback.From.ID, "admin_managing_org", map[string]interface{}{
				"org_id":   orgID,
				"org_name": orgName,
			})
			b.SendMessageWithKeyboard(
				chatID,
				"✅ Менеджер @"+manager.Username+" удалён из организации *"+bot.EscapeMarkdown(orgName)+"*",
				bot.GetOrgManageKeyboard(),
			)
			return
		}
	}
}

// handleTrainerListCallback обрабатывает выбор тренера из списка
func handleTrainerListCallback(b *bot.Bot, callback *tgbotapi.CallbackQuery, id int64, action string, chatID int64, messageID int) {
	ctx := context.Background()

	if action == "cancel" {
		b.CleanupMessages(chatID, callback.From.ID)
		state := b.GetState(callback.From.ID)
		if state != nil {
			orgID, okID := bot.GetStateInt64(state.Data, "org_id")
			orgName, okName := bot.GetStateString(state.Data, "org_name")
			if okID && okName {
				b.SetState(callback.From.ID, "manager_managing_org", map[string]interface{}{
					"org_id":   orgID,
					"org_name": orgName,
				})
				b.SendMessageWithKeyboard(chatID, "Управление организацией *"+bot.EscapeMarkdown(orgName)+"*", bot.GetManagerMenuKeyboard())
				return
			}
		}
		b.ClearState(callback.From.ID)
		return
	}

	state := b.GetState(callback.From.ID)
	if state == nil {
		return
	}

	trainers, ok := state.Data["trainers"].([]*models.OrganizationTrainer)
	if !ok {
		return
	}

	orgID, okID := bot.GetStateInt64(state.Data, "org_id")
	orgName, okName := bot.GetStateString(state.Data, "org_name")
	if !okID || !okName {
		return
	}

	// Находим тренера по ID
	for _, trainer := range trainers {
		if trainer.ID == id {
			if err := b.DB.RemoveTrainer(ctx, orgID, trainer.Username); err != nil {
				log.Printf("Error removing trainer: %v", err)
				b.AnswerCallback(callback.ID, "Ошибка удаления")
				return
			}

			b.CleanupMessages(chatID, callback.From.ID)
			b.SetState(callback.From.ID, "manager_managing_org", map[string]interface{}{
				"org_id":   orgID,
				"org_name": orgName,
			})
			b.SendMessageWithKeyboard(
				chatID,
				"✅ Тренер @"+trainer.Username+" удалён из организации *"+bot.EscapeMarkdown(orgName)+"*",
				bot.GetManagerMenuKeyboard(),
			)
			return
		}
	}
}

// handleExerciseCallback обрабатывает завершение/отмену добавления упражнений
func handleExerciseCallback(b *bot.Bot, callback *tgbotapi.CallbackQuery, action string, accessInfo *models.AccessInfo, chatID int64, messageID int) {
	b.CleanupMessages(chatID, callback.From.ID)
	b.ClearState(callback.From.ID)

	if action == "finish" {
		b.SendMessageWithKeyboard(chatID, "✅ Тренировка сохранена! 💪", bot.GetStartMenuKeyboard(accessInfo))
	} else {
		b.SendMessageWithKeyboard(chatID, "❌ Тренировка отменена.", bot.GetStartMenuKeyboard(accessInfo))
	}
}

func handleUpdate(b *bot.Bot, message *tgbotapi.Message) {
	ctx := context.Background()

	// Связываем telegram_id с username при каждом сообщении
	if message.From.UserName != "" {
		if err := b.DB.LinkTelegramID(ctx, message.From.ID, message.From.UserName); err != nil {
			log.Printf("Error linking telegram ID: %v", err)
		}
	}

	// Получаем информацию о доступах пользователя
	username := message.From.UserName
	accessInfo, err := b.DB.GetUserAccessInfo(ctx, message.From.ID, username)
	if err != nil {
		log.Printf("Error getting access info: %v", err)
		b.SendMessage(message.Chat.ID, "❌ Ошибка при проверке доступов.")
		return
	}

	// Проверяем админа
	accessInfo.IsAdmin = b.IsAdmin(username)

	// Обеспечиваем/обновляем запись пользователя
	b.DB.EnsureUser(ctx, message.From.ID, username, message.From.FirstName+" "+message.From.LastName)

	state := b.GetState(message.From.ID)

	// Обработка команд
	if message.IsCommand() {
		switch message.Command() {
		case "start":
			handleStartCommand(b, message, accessInfo)
		default:
			b.SendMessage(message.Chat.ID, "Неизвестная команда. Используйте /start")
		}
		return
	}

	// Обработка состояний
	if state != nil {
		handleState(b, message, state, accessInfo)
		return
	}

	// Обработка кнопок меню
	handleMenuButtons(b, message, accessInfo)
}

func handleStartCommand(b *bot.Bot, message *tgbotapi.Message, accessInfo *models.AccessInfo) {
	b.ClearState(message.From.ID)

	// Формируем приветствие
	var sb strings.Builder
	sb.WriteString("👋 *Добро пожаловать в FitBot!*\n\n")

	hasAccess := false

	if accessInfo.IsAdmin {
		sb.WriteString("👑 Вы администратор системы\n")
		hasAccess = true
	}

	if len(accessInfo.ManagerOrgs) > 0 {
		activeCount := 0
		for _, org := range accessInfo.ManagerOrgs {
			if org.IsActive {
				activeCount++
			}
		}
		if activeCount > 0 {
			sb.WriteString("🏢 Менеджер организаций: " + strconv.Itoa(activeCount) + "\n")
			hasAccess = true
		}
	}

	if len(accessInfo.TrainerOrgs) > 0 {
		activeCount := 0
		for _, org := range accessInfo.TrainerOrgs {
			if org.IsActive {
				activeCount++
			}
		}
		if activeCount > 0 {
			sb.WriteString("🏋️ Тренер в организациях: " + strconv.Itoa(activeCount) + "\n")
			hasAccess = true
		}
	}

	if len(accessInfo.ClientAccess) > 0 {
		sb.WriteString("📝 Активных тренеров: " + strconv.Itoa(len(accessInfo.ClientAccess)) + "\n")
		hasAccess = true
	}

	if len(accessInfo.ArchivedAccess) > 0 {
		sb.WriteString("📚 Архивных записей: " + strconv.Itoa(len(accessInfo.ArchivedAccess)) + "\n")
	}

	if !hasAccess && len(accessInfo.ArchivedAccess) == 0 {
		handlers.HandleNoAccess(b, message)
		return
	}

	sb.WriteString("\nВыберите действие:")

	b.SendMessageWithKeyboard(message.Chat.ID, sb.String(), bot.GetStartMenuKeyboard(accessInfo))
}

func handleState(b *bot.Bot, message *tgbotapi.Message, state *models.UserState, accessInfo *models.AccessInfo) {
	// Проверка на "Главное меню" или "Отмена"
	if message.Text == "🔙 Главное меню" {
		b.ClearState(message.From.ID)
		handleStartCommand(b, message, accessInfo)
		return
	}

	switch state.State {
	// ===== АДМИН =====
	case "admin_creating_org_name":
		handlers.HandleCreateOrganizationName(b, message)
	case "admin_creating_org_code":
		handlers.HandleCreateOrganizationCode(b, message)
	case "admin_selecting_org":
		if message.Text == "❌ Отмена" {
			b.ClearState(message.From.ID)
			handlers.HandleAdminMenu(b, message)
			return
		}
		if idx, err := strconv.Atoi(message.Text); err == nil {
			handlers.HandleSelectOrganization(b, message, idx)
		} else {
			b.SendMessage(message.Chat.ID, "⚠️ Введите номер организации или нажмите «❌ Отмена»")
		}
	case "admin_managing_org":
		handleAdminOrgActions(b, message, accessInfo)
	case "admin_adding_manager":
		handlers.HandleAddManagerUsername(b, message)
	case "admin_removing_manager":
		if message.Text == "❌ Отмена" {
			orgID, okID := bot.GetStateInt64(state.Data, "org_id")
			orgName, okName := bot.GetStateString(state.Data, "org_name")
			if !okID || !okName {
				b.ClearState(message.From.ID)
				handlers.HandleAdminMenu(b, message)
				return
			}
			b.SetState(message.From.ID, "admin_managing_org", map[string]interface{}{
				"org_id":   orgID,
				"org_name": orgName,
			})
			b.SendMessageWithKeyboard(message.Chat.ID,
				"Управление организацией *"+bot.EscapeMarkdown(orgName)+"*",
				bot.GetOrgManageKeyboard())
			return
		}
		if idx, err := strconv.Atoi(message.Text); err == nil {
			handlers.HandleRemoveManager(b, message, idx)
		} else {
			b.SendMessage(message.Chat.ID, "⚠️ Введите номер менеджера для удаления или нажмите «❌ Отмена»")
		}

	// ===== МЕНЕДЖЕР =====
	case "manager_selecting_org":
		if message.Text == "❌ Отмена" {
			b.ClearState(message.From.ID)
			handleStartCommand(b, message, accessInfo)
			return
		}
		if idx, err := strconv.Atoi(message.Text); err == nil {
			handlers.HandleManagerSelectOrg(b, message, idx)
		} else {
			b.SendMessage(message.Chat.ID, "⚠️ Введите номер организации или нажмите «❌ Отмена»")
		}
	case "manager_managing_org":
		handleManagerOrgActions(b, message, accessInfo)
	case "manager_adding_trainer":
		handlers.HandleAddTrainerUsername(b, message)
	case "manager_removing_trainer":
		if message.Text == "❌ Отмена" {
			orgID, okID := bot.GetStateInt64(state.Data, "org_id")
			orgName, okName := bot.GetStateString(state.Data, "org_name")
			if !okID || !okName {
				b.ClearState(message.From.ID)
				handleStartCommand(b, message, accessInfo)
				return
			}
			b.SetState(message.From.ID, "manager_managing_org", map[string]interface{}{
				"org_id":   orgID,
				"org_name": orgName,
			})
			b.SendMessageWithKeyboard(message.Chat.ID,
				"Управление организацией *"+bot.EscapeMarkdown(orgName)+"*",
				bot.GetManagerMenuKeyboard())
			return
		}
		if idx, err := strconv.Atoi(message.Text); err == nil {
			handlers.HandleRemoveTrainer(b, message, idx)
		} else {
			b.SendMessage(message.Chat.ID, "⚠️ Введите номер тренера для удаления или нажмите «❌ Отмена»")
		}

	// ===== ТРЕНЕР =====
	case "trainer_selecting_org":
		if message.Text == "❌ Отмена" {
			b.ClearState(message.From.ID)
			handleStartCommand(b, message, accessInfo)
			return
		}
		if idx, err := strconv.Atoi(message.Text); err == nil {
			handlers.HandleTrainerSelectOrg(b, message, idx)
		} else {
			b.SendMessage(message.Chat.ID, "⚠️ Введите номер организации или нажмите «❌ Отмена»")
		}
	case "trainer_managing_org":
		handleTrainerOrgActions(b, message, accessInfo)
	case "trainer_adding_client":
		handlers.HandleAddClientUsername(b, message)
	case "trainer_viewing_clients":
		text := message.Text
		if text == "❌ Отмена" {
			trainerID, okT := bot.GetStateInt64(state.Data, "trainer_id")
			orgID, okID := bot.GetStateInt64(state.Data, "org_id")
			orgName, okName := bot.GetStateString(state.Data, "org_name")
			if !okT || !okID || !okName {
				b.ClearState(message.From.ID)
				handleStartCommand(b, message, accessInfo)
				return
			}
			b.SetState(message.From.ID, "trainer_managing_org", map[string]interface{}{
				"trainer_id": trainerID,
				"org_id":     orgID,
				"org_name":   orgName,
			})
			b.SendMessageWithKeyboard(message.Chat.ID,
				"🏋️ *Панель тренера - "+bot.EscapeMarkdown(orgName)+"*",
				bot.GetTrainerMenuKeyboard())
			return
		}
		if strings.HasPrefix(strings.ToLower(text), "удалить ") {
			parts := strings.Fields(text)
			if len(parts) >= 2 {
				if idx, err := strconv.Atoi(parts[1]); err == nil {
					handlers.HandleRemoveClientByIndex(b, message, idx)
					return
				}
			}
		}
		if idx, err := strconv.Atoi(text); err == nil {
			handlers.HandleSelectClient(b, message, idx)
		} else {
			b.SendMessage(message.Chat.ID, "⚠️ Введите номер клиента, «удалить [номер]» или «❌ Отмена»")
		}
	case "trainer_client_action":
		if idx, err := strconv.Atoi(message.Text); err == nil {
			handlers.HandleClientAction(b, message, idx)
		} else {
			b.SendMessage(message.Chat.ID, "⚠️ Введите номер действия (1-4)")
		}

	// ===== КЛИЕНТ =====
	case "client_selecting_trainer":
		if message.Text == "❌ Отмена" {
			b.ClearState(message.From.ID)
			handleStartCommand(b, message, accessInfo)
			return
		}
		if idx, err := strconv.Atoi(message.Text); err == nil {
			handlers.HandleClientSelectTrainer(b, message, idx)
		} else {
			b.SendMessage(message.Chat.ID, "⚠️ Введите номер тренера или нажмите «❌ Отмена»")
		}
	case "client_with_trainer":
		handleClientActions(b, message, accessInfo)
	case "client_viewing_archive":
		if message.Text == "❌ Отмена" {
			b.ClearState(message.From.ID)
			handleStartCommand(b, message, accessInfo)
			return
		}
		if idx, err := strconv.Atoi(message.Text); err == nil {
			handlers.HandleSelectArchivedTrainer(b, message, idx)
		} else {
			b.SendMessage(message.Chat.ID, "⚠️ Введите номер записи или нажмите «❌ Отмена»")
		}

	// ===== ТРЕНИРОВКИ =====
	case "awaiting_muscle_group":
		handlers.HandleMuscleGroupSelection(b, message)
	case "adding_exercises":
		handlers.HandleAddExercise(b, message)
	case "awaiting_exercise_name":
		handlers.HandleExerciseNameForStats(b, message)

	// ===== ГРУППОВЫЕ ТРЕНИРОВКИ =====
	case "joining_group_training":
		if message.Text == "❌ Отмена" {
			b.ClearState(message.From.ID)
			handleStartCommand(b, message, accessInfo)
			return
		}
		if idx, err := strconv.Atoi(message.Text); err == nil {
			handlers.HandleJoinGroupTraining(b, message, idx)
		} else {
			b.SendMessage(message.Chat.ID, "⚠️ Введите номер тренировки или нажмите «❌ Отмена»")
		}
	case "creating_group_training":
		handlers.HandleCreateGroupTrainingData(b, message)

	default:
		b.ClearState(message.From.ID)
		handleStartCommand(b, message, accessInfo)
	}
}

func handleMenuButtons(b *bot.Bot, message *tgbotapi.Message, accessInfo *models.AccessInfo) {
	switch message.Text {
	// ===== АДМИН =====
	case "👑 Админ-панель":
		if accessInfo.IsAdmin {
			handlers.HandleAdminMenu(b, message)
		} else {
			b.SendMessage(message.Chat.ID, "❌ Доступ запрещён.")
		}
	case "🏢 Создать организацию":
		if accessInfo.IsAdmin {
			handlers.HandleCreateOrganization(b, message)
		}
	case "📋 Список организаций":
		if accessInfo.IsAdmin {
			handlers.HandleListOrganizations(b, message)
		}

	// ===== МЕНЕДЖЕР =====
	case "🏢 Управление организацией":
		handlers.HandleManagerMenu(b, message, accessInfo.ManagerOrgs)

	// ===== ТРЕНЕР =====
	case "🏋️ Панель тренера":
		handlers.HandleTrainerMenu(b, message, accessInfo.TrainerOrgs)

	// ===== КЛИЕНТ =====
	case "📝 Мои тренировки":
		if len(accessInfo.ClientAccess) > 0 {
			handlers.HandleClientMenu(b, message, accessInfo.ClientAccess)
		} else {
			b.SendMessage(message.Chat.ID, "❌ У вас нет активных доступов к тренерам.")
		}

	case "📚 Архив тренировок":
		handlers.HandleArchivedAccess(b, message, accessInfo.ArchivedAccess)

	case "ℹ️ О боте":
		handlers.HandleNoAccess(b, message)

	default:
		b.SendMessageWithKeyboard(
			message.Chat.ID,
			"Выберите действие из меню:",
			bot.GetStartMenuKeyboard(accessInfo),
		)
	}
}

// handleAdminOrgActions обрабатывает действия в управлении организацией (админ)
func handleAdminOrgActions(b *bot.Bot, message *tgbotapi.Message, accessInfo *models.AccessInfo) {
	switch message.Text {
	case "➕ Добавить менеджера":
		handlers.HandleAddManager(b, message)
	case "📋 Список менеджеров":
		handlers.HandleListManagers(b, message)
	case "🔙 К списку организаций":
		handlers.HandleListOrganizations(b, message)
	default:
		b.SendMessage(message.Chat.ID, "Выберите действие из меню.")
	}
}

// handleManagerOrgActions обрабатывает действия в панели менеджера
func handleManagerOrgActions(b *bot.Bot, message *tgbotapi.Message, accessInfo *models.AccessInfo) {
	switch message.Text {
	case "➕ Добавить тренера":
		handlers.HandleAddTrainer(b, message)
	case "📋 Список тренеров":
		handlers.HandleListTrainers(b, message)
	case "🔙 Главное меню":
		b.ClearState(message.From.ID)
		handleStartCommand(b, message, accessInfo)
	default:
		b.SendMessage(message.Chat.ID, "Выберите действие из меню.")
	}
}

// handleTrainerOrgActions обрабатывает действия в панели тренера
func handleTrainerOrgActions(b *bot.Bot, message *tgbotapi.Message, accessInfo *models.AccessInfo) {
	switch message.Text {
	case "➕ Добавить клиента":
		handlers.HandleAddClient(b, message)
	case "👥 Мои клиенты":
		handlers.HandleListClients(b, message)
	case "📅 Групповые тренировки":
		handlers.HandleGroupTrainings(b, message)
	case "📊 Статистика":
		handlers.HandleStats(b, message)
	case "🔙 Главное меню":
		b.ClearState(message.From.ID)
		handleStartCommand(b, message, accessInfo)
	default:
		b.SendMessage(message.Chat.ID, "Выберите действие из меню.")
	}
}

// handleClientActions обрабатывает действия в панели клиента
func handleClientActions(b *bot.Bot, message *tgbotapi.Message, accessInfo *models.AccessInfo) {
	switch message.Text {
	case "➕ Добавить тренировку":
		handlers.HandleAddWorkout(b, message)
	case "📝 Мои тренировки":
		handlers.HandleMyWorkouts(b, message)
	case "📊 Моя статистика":
		handlers.HandleStats(b, message)
	case "📅 Групповые тренировки":
		handlers.HandleGroupTrainings(b, message)
	case "🔙 Главное меню":
		b.ClearState(message.From.ID)
		handleStartCommand(b, message, accessInfo)
	default:
		b.SendMessage(message.Chat.ID, "Выберите действие из меню.")
	}
}
