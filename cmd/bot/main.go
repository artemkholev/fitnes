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

// safeHandleUpdate запускает handleUpdate с recover — сбой одного пользователя не роняет весь процесс.
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

func handleCallback(b *bot.Bot, callback *tgbotapi.CallbackQuery) {
	b.AnswerCallback(callback.ID, "")

	prefix, id, action := bot.ParseCallbackData(callback.Data)

	username := callback.From.UserName
	accessInfo, err := b.DB.GetUserAccessInfo(callback.From.ID, username)
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
	case "ex_sets":
		handleExSetsCallback(b, callback, id, action, chatID, messageID)
	case "ex_reps":
		handleExRepsCallback(b, callback, id, action, chatID, messageID)
	case "ex_weight":
		handleExWeightCallback(b, callback, id, action, chatID, messageID)
	case "date":
		handleDateCallback(b, callback, action, accessInfo, chatID, messageID)
	case "workout":
		handleWorkoutCallback(b, callback, id, action, chatID, messageID)
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

func handleMuscleCallback(b *bot.Bot, callback *tgbotapi.CallbackQuery, action string, chatID int64, messageID int) {
	if action == "cancel" {
		b.CleanupMessages(chatID, callback.From.ID)
		b.ClearState(callback.From.ID)
		accessInfo, _ := b.DB.GetUserAccessInfo(callback.From.ID, callback.From.UserName)
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

	var trainerClientID *int64
	if state.Data != nil {
		if tcID, ok := bot.GetStateInt64(state.Data, "trainer_client_id"); ok && tcID > 0 {
			trainerClientID = &tcID
		}
	}

	// Берём telegram_id клиента из состояния (для тренера, создающего тренировку клиенту).
	// Если клиент ещё не запускал бота — его telegram_id неизвестен, используем 0.
	var clientTelegramID int64
	if trainerClientID != nil {
		if tcTID, ok := state.Data["client_telegram_id"].(*int64); ok && tcTID != nil {
			clientTelegramID = *tcTID
		}
	} else {
		clientTelegramID = callback.From.ID
	}

	workoutDate := time.Now()
	if state.Data != nil {
		if d, ok := state.Data["workout_date"].(time.Time); ok {
			workoutDate = d
		}
	}

	workout := &models.Workout{
		TrainerClientID:  trainerClientID,
		ClientTelegramID: clientTelegramID,
		Date:             workoutDate,
		MuscleGroup:      muscleGroup,
	}

	if err := b.DB.CreateWorkout(workout); err != nil {
		log.Printf("Error creating workout: %v", err)
		b.EditMessageText(chatID, messageID, "❌ Ошибка при создании тренировки.", nil)
		return
	}

	b.SetState(callback.From.ID, "adding_exercises", map[string]interface{}{
		"workout_id": workout.ID,
		"order":      1,
		"step":       "name",
	})

	b.CleanupMessages(chatID, callback.From.ID)
	b.SendMessageWithKeyboard(chatID, "🏋️ Тренировка создана!\n\nВведите название первого упражнения:", bot.GetExerciseControlKeyboard())
}

// handleExSetsCallback обрабатывает выбор количества подходов из inline-клавиатуры.
// id содержит выбранное значение; action == "other" переводит в текстовый ввод.
func handleExSetsCallback(b *bot.Bot, callback *tgbotapi.CallbackQuery, id int64, action string, chatID int64, messageID int) {
	state := b.GetState(callback.From.ID)
	if state == nil {
		return
	}

	if action == "cancel" {
		b.CleanupMessages(chatID, callback.From.ID)
		b.ClearState(callback.From.ID)
		accessInfo, _ := b.DB.GetUserAccessInfo(callback.From.ID, callback.From.UserName)
		accessInfo.IsAdmin = b.IsAdmin(callback.From.UserName)
		b.SendMessageWithKeyboard(chatID, "❌ Тренировка отменена.", bot.GetStartMenuKeyboard(accessInfo))
		return
	}

	if action == "other" {
		state.Data["step"] = "sets_custom"
		b.EditMessageText(chatID, messageID, "Введите количество подходов:", nil)
		return
	}

	sets := int(id)
	if sets <= 0 {
		return
	}

	name, _ := bot.GetStateString(state.Data, "exercise_name")
	state.Data["exercise_sets"] = sets
	state.Data["step"] = "reps"

	keyboard := bot.GetInlineRepsKeyboard()
	b.EditMessageText(chatID, messageID,
		"*"+bot.EscapeMarkdown(name)+"* | Подходы: "+strconv.Itoa(sets)+"\n\nВыберите количество повторений:",
		&keyboard,
	)
}

// handleExRepsCallback обрабатывает выбор количества повторений.
func handleExRepsCallback(b *bot.Bot, callback *tgbotapi.CallbackQuery, id int64, action string, chatID int64, messageID int) {
	state := b.GetState(callback.From.ID)
	if state == nil {
		return
	}

	if action == "cancel" {
		b.CleanupMessages(chatID, callback.From.ID)
		b.ClearState(callback.From.ID)
		accessInfo, _ := b.DB.GetUserAccessInfo(callback.From.ID, callback.From.UserName)
		accessInfo.IsAdmin = b.IsAdmin(callback.From.UserName)
		b.SendMessageWithKeyboard(chatID, "❌ Тренировка отменена.", bot.GetStartMenuKeyboard(accessInfo))
		return
	}

	if action == "other" {
		state.Data["step"] = "reps_custom"
		b.EditMessageText(chatID, messageID, "Введите количество повторений:", nil)
		return
	}

	reps := int(id)
	if reps <= 0 {
		return
	}

	name, _ := bot.GetStateString(state.Data, "exercise_name")
	sets, _ := bot.GetStateInt64(state.Data, "exercise_sets")
	state.Data["exercise_reps"] = reps
	state.Data["step"] = "weight"

	keyboard := bot.GetInlineWeightKeyboard()
	b.EditMessageText(chatID, messageID,
		"*"+bot.EscapeMarkdown(name)+"* | Подходы: "+strconv.FormatInt(sets, 10)+" | Повт.: "+strconv.Itoa(reps)+"\n\nВыберите вес (кг):",
		&keyboard,
	)
}

// handleExWeightCallback обрабатывает выбор веса и сохраняет упражнение.
func handleExWeightCallback(b *bot.Bot, callback *tgbotapi.CallbackQuery, id int64, action string, chatID int64, messageID int) {
	state := b.GetState(callback.From.ID)
	if state == nil {
		return
	}

	if action == "cancel" {
		b.CleanupMessages(chatID, callback.From.ID)
		b.ClearState(callback.From.ID)
		accessInfo, _ := b.DB.GetUserAccessInfo(callback.From.ID, callback.From.UserName)
		accessInfo.IsAdmin = b.IsAdmin(callback.From.UserName)
		b.SendMessageWithKeyboard(chatID, "❌ Тренировка отменена.", bot.GetStartMenuKeyboard(accessInfo))
		return
	}

	if action == "other" {
		state.Data["step"] = "weight_custom"
		b.EditMessageText(chatID, messageID, "Введите вес в кг (например: 80 или 72.5):", nil)
		return
	}

	if action == "bodyweight" {
		handlers.SaveExerciseStep(b, callback.From.ID, chatID, messageID, 0, true)
		return
	}

	handlers.SaveExerciseStep(b, callback.From.ID, chatID, messageID, float64(id), true)
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
		b.SendMessage(chatID, "❌ Список устарел. Нажмите «👥 Мои клиенты» ещё раз.")
		return
	}

	clients, ok := state.Data["clients"].([]*models.ClientWithInfo)
	if !ok || len(clients) == 0 {
		b.SendMessage(chatID, "❌ Список клиентов устарел. Нажмите «👥 Мои клиенты» ещё раз.")
		return
	}

	var found *models.ClientWithInfo
	for _, client := range clients {
		if client.Client.ID == id {
			found = client
			break
		}
	}
	if found == nil {
		b.SendMessage(chatID, "❌ Клиент не найден. Обновите список.")
		return
	}

	// Удаляем список, показываем карточку клиента новым сообщением внизу чата
	b.DeleteMessage(chatID, messageID)
	showClientCard(b, callback.From.ID, chatID, found, state.Data)
}

// showClientCard отправляет карточку клиента с inline-кнопками действий.
func showClientCard(b *bot.Bot, userID, chatID int64, client *models.ClientWithInfo, parentData map[string]interface{}) {
	var sb strings.Builder
	name := client.Client.Username
	if client.FullName != "" {
		name = client.FullName
	}
	sb.WriteString("👤 *" + bot.EscapeMarkdown(name) + "*\n")
	sb.WriteString("@" + bot.EscapeMarkdown(client.Client.Username) + "\n")
	sb.WriteString("Тренировок: " + strconv.Itoa(client.WorkoutCount) + "\n")
	if client.LastWorkout != nil {
		sb.WriteString("Последняя: " + client.LastWorkout.Format("02.01.2006") + "\n")
	}
	if client.Client.IsActive {
		sb.WriteString("Статус: Активен ✅")
	} else {
		sb.WriteString("Статус: Деактивирован ❌")
	}

	b.SetState(userID, "trainer_client_action", map[string]interface{}{
		"trainer_id": parentData["trainer_id"],
		"org_id":     parentData["org_id"],
		"org_name":   parentData["org_name"],
		"client":     client,
	})

	keyboard := bot.GetInlineClientActionsKeyboard(client.Client.ID, client.Client.IsActive)
	b.SendInlineKeyboard(chatID, sb.String(), keyboard)
}

// handleClientActionCallback обрабатывает действия с клиентом
func handleClientActionCallback(b *bot.Bot, callback *tgbotapi.CallbackQuery, id int64, action string, chatID int64, messageID int) {
	state := b.GetState(callback.From.ID)
	if state == nil {
		b.AnswerCallback(callback.ID, "Сессия истекла, начните заново")
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
		b.AnswerCallback(callback.ID, "Ошибка: данные клиента не найдены")
		return
	}

	trainerID, okT := bot.GetStateInt64(state.Data, "trainer_id")
	orgID, okID := bot.GetStateInt64(state.Data, "org_id")
	orgName, okName := bot.GetStateString(state.Data, "org_name")
	if !okT || !okID || !okName {
		b.AnswerCallback(callback.ID, "Ошибка: данные сессии повреждены")
		return
	}

	switch action {
	case "stats":
		b.EditMessageText(chatID, messageID, "📊 Статистика клиента @"+bot.EscapeMarkdown(client.Client.Username)+" будет добавлена позже.", nil)

	case "workout":
		b.CleanupMessages(chatID, callback.From.ID)
		b.SetState(callback.From.ID, "awaiting_workout_date", map[string]interface{}{
			"trainer_id":         trainerID,
			"org_id":             orgID,
			"org_name":           orgName,
			"client":             client,
			"trainer_client_id":  client.Client.ID,
			"client_telegram_id": client.Client.TelegramID,
		})
		keyboard := bot.GetInlineDateKeyboard()
		msgID := b.SendInlineKeyboard(chatID, "➕ *Создание тренировки для @"+bot.EscapeMarkdown(client.Client.Username)+"*\n\nВыберите дату тренировки:", keyboard)
		b.StoreMessageID(callback.From.ID, msgID)

	case "history":
		workouts, err := b.DB.GetWorkoutsByTrainerClient(client.Client.ID, 20)
		if err != nil {
			log.Printf("Error getting client workouts: %v", err)
			b.EditMessageText(chatID, messageID, "❌ Ошибка при получении истории тренировок.", nil)
			return
		}
		if len(workouts) == 0 {
			b.EditMessageText(chatID, messageID, "📋 У @"+bot.EscapeMarkdown(client.Client.Username)+" пока нет тренировок.", nil)
			return
		}
		exerciseCounts := make(map[int64]int)
		for _, w := range workouts {
			exercises, _ := b.DB.GetExercisesByWorkout(w.ID)
			exerciseCounts[w.ID] = len(exercises)
		}
		b.SetState(callback.From.ID, "viewing_client_workouts", map[string]interface{}{
			"trainer_id":      trainerID,
			"org_id":          orgID,
			"org_name":        orgName,
			"client":          client,
			"workouts":        workouts,
			"exercise_counts": exerciseCounts,
			"is_trainer_view": true,
		})
		listKeyboard := bot.GetInlineWorkoutsKeyboard(workouts, exerciseCounts)
		b.EditMessageText(chatID, messageID, "📋 *История тренировок @"+bot.EscapeMarkdown(client.Client.Username)+"*", &listKeyboard)

	case "delete":
		if !client.Client.IsActive {
			b.AnswerCallback(callback.ID, "Клиент уже деактивирован")
			return
		}
		if err := b.DB.RemoveClient( trainerID, client.Client.Username); err != nil {
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
			"✅ Клиент @"+bot.EscapeMarkdown(client.Client.Username)+" удалён.",
			bot.GetTrainerMenuKeyboard(),
		)
	}
}

// handleManagerListCallback обрабатывает выбор менеджера из списка
func handleManagerListCallback(b *bot.Bot, callback *tgbotapi.CallbackQuery, id int64, action string, chatID int64, messageID int) {

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
			if err := b.DB.RemoveManager( orgID, manager.Username); err != nil {
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
			if err := b.DB.RemoveTrainer( orgID, trainer.Username); err != nil {
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

func handleExerciseCallback(b *bot.Bot, callback *tgbotapi.CallbackQuery, action string, accessInfo *models.AccessInfo, chatID int64, messageID int) {
	switch action {
	case "finish":
		b.CleanupMessages(chatID, callback.From.ID)
		b.ClearState(callback.From.ID)
		b.SendMessageWithKeyboard(chatID, "✅ Тренировка сохранена! 💪", bot.GetStartMenuKeyboard(accessInfo))
	case "cancel":
		b.CleanupMessages(chatID, callback.From.ID)
		b.ClearState(callback.From.ID)
		b.SendMessageWithKeyboard(chatID, "❌ Тренировка отменена.", bot.GetStartMenuKeyboard(accessInfo))
	case "more":
		// Удаляем сообщение с кнопками и просим имя следующего упражнения
		b.DeleteMessage(chatID, messageID)
		state := b.GetState(callback.From.ID)
		if state != nil {
			state.Data["step"] = "name"
		}
		b.SendMessage(chatID, "➕ Введите название следующего упражнения:")
	}
}

func handleUpdate(b *bot.Bot, message *tgbotapi.Message) {

	// Связываем telegram_id с username при каждом сообщении
	if message.From.UserName != "" {
		if err := b.DB.LinkTelegramID( message.From.ID, message.From.UserName); err != nil {
			log.Printf("Error linking telegram ID: %v", err)
		}
	}

	// Получаем информацию о доступах пользователя
	username := message.From.UserName
	accessInfo, err := b.DB.GetUserAccessInfo( message.From.ID, username)
	if err != nil {
		log.Printf("Error getting access info: %v", err)
		b.SendMessage(message.Chat.ID, "❌ Ошибка при проверке доступов.")
		return
	}

	// Проверяем админа
	accessInfo.IsAdmin = b.IsAdmin(username)

	// Обеспечиваем/обновляем запись пользователя
	b.DB.EnsureUser(message.From.ID, username, message.From.FirstName+" "+message.From.LastName)

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
	b.CleanupMessages(message.Chat.ID, message.From.ID)
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
	if message.Text == "🔙 Главное меню" {
		// Сначала удаляем inline-сообщения, потом чистим состояние.
		// Обратный порядок сломает очистку — ClearState удалит список ID.
		b.CleanupMessages(message.Chat.ID, message.From.ID)
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
			b.CleanupMessages(message.Chat.ID, message.From.ID)
			b.ClearState(message.From.ID)
			handlers.HandleAdminMenu(b, message)
			return
		}
		if idx, err := strconv.Atoi(message.Text); err == nil {
			handlers.HandleSelectOrganization(b, message, idx)
		} else {
			b.CleanupMessages(message.Chat.ID, message.From.ID)
			handlers.HandleAdminMenu(b, message)
		}
	case "admin_managing_org":
		handleAdminOrgActions(b, message, accessInfo)
	case "admin_adding_manager":
		handlers.HandleAddManagerUsername(b, message)
	case "admin_removing_manager":
		if message.Text == "❌ Отмена" {
			orgID, okID := bot.GetStateInt64(state.Data, "org_id")
			orgName, okName := bot.GetStateString(state.Data, "org_name")
			b.CleanupMessages(message.Chat.ID, message.From.ID)
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
		b.CleanupMessages(message.Chat.ID, message.From.ID)
		handleAdminOrgActions(b, message, accessInfo)

	// ===== МЕНЕДЖЕР =====
	case "manager_selecting_org":
		if message.Text == "❌ Отмена" {
			b.CleanupMessages(message.Chat.ID, message.From.ID)
			b.ClearState(message.From.ID)
			handleStartCommand(b, message, accessInfo)
			return
		}
		if idx, err := strconv.Atoi(message.Text); err == nil {
			handlers.HandleManagerSelectOrg(b, message, idx)
		} else {
			b.CleanupMessages(message.Chat.ID, message.From.ID)
			handleStartCommand(b, message, accessInfo)
		}
	case "manager_managing_org":
		handleManagerOrgActions(b, message, accessInfo)
	case "manager_adding_trainer":
		handlers.HandleAddTrainerUsername(b, message)
	case "manager_removing_trainer":
		if message.Text == "❌ Отмена" {
			orgID, okID := bot.GetStateInt64(state.Data, "org_id")
			orgName, okName := bot.GetStateString(state.Data, "org_name")
			b.CleanupMessages(message.Chat.ID, message.From.ID)
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
		b.CleanupMessages(message.Chat.ID, message.From.ID)
		handleManagerOrgActions(b, message, accessInfo)

	// ===== ТРЕНЕР =====
	case "trainer_selecting_org":
		if message.Text == "❌ Отмена" {
			b.CleanupMessages(message.Chat.ID, message.From.ID)
			b.ClearState(message.From.ID)
			handleStartCommand(b, message, accessInfo)
			return
		}
		if idx, err := strconv.Atoi(message.Text); err == nil {
			handlers.HandleTrainerSelectOrg(b, message, idx)
		} else {
			b.CleanupMessages(message.Chat.ID, message.From.ID)
			handleStartCommand(b, message, accessInfo)
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
			b.CleanupMessages(message.Chat.ID, message.From.ID)
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
			b.CleanupMessages(message.Chat.ID, message.From.ID)
			handleTrainerOrgActions(b, message, accessInfo)
		}
	case "trainer_client_action":
		// Все действия выполняются через inline-кнопки. При вводе текста
		// повторно показываем карточку клиента с кнопками.
		client, ok := state.Data["client"].(*models.ClientWithInfo)
		if ok && client != nil {
			showClientCard(b, message.From.ID, message.Chat.ID, client, state.Data)
		} else {
			b.ClearState(message.From.ID)
			handleStartCommand(b, message, accessInfo)
		}

	// ===== КЛИЕНТ =====
	case "client_selecting_trainer":
		if message.Text == "❌ Отмена" {
			b.CleanupMessages(message.Chat.ID, message.From.ID)
			b.ClearState(message.From.ID)
			handleStartCommand(b, message, accessInfo)
			return
		}
		if idx, err := strconv.Atoi(message.Text); err == nil {
			handlers.HandleClientSelectTrainer(b, message, idx)
		} else {
			b.CleanupMessages(message.Chat.ID, message.From.ID)
			handleStartCommand(b, message, accessInfo)
		}
	case "client_with_trainer":
		handleClientActions(b, message, accessInfo)
	case "client_viewing_archive":
		if message.Text == "❌ Отмена" {
			b.CleanupMessages(message.Chat.ID, message.From.ID)
			b.ClearState(message.From.ID)
			handleStartCommand(b, message, accessInfo)
			return
		}
		if idx, err := strconv.Atoi(message.Text); err == nil {
			handlers.HandleSelectArchivedTrainer(b, message, idx)
		} else {
			b.CleanupMessages(message.Chat.ID, message.From.ID)
			handleStartCommand(b, message, accessInfo)
		}

	// ===== ТРЕНИРОВКИ =====
	case "awaiting_workout_date":
		if message.Text == "❌ Отмена" {
			b.CleanupMessages(message.Chat.ID, message.From.ID)
			b.ClearState(message.From.ID)
			b.SendMessageWithKeyboard(message.Chat.ID, "Отменено.", bot.GetStartMenuKeyboard(accessInfo))
			return
		}
		step, _ := bot.GetStateString(state.Data, "step")
		if step == "date_custom" {
			handlers.HandleWorkoutDateCustom(b, message)
		}
	case "viewing_workouts", "viewing_client_workouts":
		// Всё управление через inline-кнопки
	case "awaiting_muscle_group":
		handlers.HandleMuscleGroupSelection(b, message)
	case "adding_exercises":
		// Отмена и завершение обрабатываются здесь независимо от шага
		if message.Text == "❌ Отмена" {
			b.CleanupMessages(message.Chat.ID, message.From.ID)
			b.ClearState(message.From.ID)
			accessInfo2, _ := b.DB.GetUserAccessInfo(message.From.ID, message.From.UserName)
			accessInfo2.IsAdmin = b.IsAdmin(message.From.UserName)
			b.SendMessageWithKeyboard(message.Chat.ID, "❌ Тренировка отменена.", bot.GetStartMenuKeyboard(accessInfo2))
			return
		}
		if message.Text == "✅ Завершить" {
			b.CleanupMessages(message.Chat.ID, message.From.ID)
			b.ClearState(message.From.ID)
			accessInfo2, _ := b.DB.GetUserAccessInfo(message.From.ID, message.From.UserName)
			accessInfo2.IsAdmin = b.IsAdmin(message.From.UserName)
			b.SendMessageWithKeyboard(message.Chat.ID, "✅ Тренировка сохранена! 💪", bot.GetStartMenuKeyboard(accessInfo2))
			return
		}
		step, _ := bot.GetStateString(state.Data, "step")
		switch step {
		case "", "name":
			handlers.HandleExerciseName(b, message)
		case "sets_custom":
			handlers.HandleExerciseSetsCustom(b, message)
		case "reps_custom":
			handlers.HandleExerciseRepsCustom(b, message)
		case "weight_custom":
			handlers.HandleExerciseWeightCustom(b, message)
		default:
			b.SendMessage(message.Chat.ID, "Используйте кнопки для выбора значения.")
		}
	case "awaiting_exercise_name":
		handlers.HandleExerciseNameForStats(b, message)

	// ===== ГРУППОВЫЕ ТРЕНИРОВКИ =====
	case "joining_group_training":
		if message.Text == "❌ Отмена" {
			b.CleanupMessages(message.Chat.ID, message.From.ID)
			b.ClearState(message.From.ID)
			handleStartCommand(b, message, accessInfo)
			return
		}
		if idx, err := strconv.Atoi(message.Text); err == nil {
			handlers.HandleJoinGroupTraining(b, message, idx)
		} else {
			b.CleanupMessages(message.Chat.ID, message.From.ID)
			handleStartCommand(b, message, accessInfo)
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
		handleStartCommand(b, message, accessInfo)
	default:
		b.SendMessage(message.Chat.ID, "Выберите действие из меню.")
	}
}

func handleTrainerOrgActions(b *bot.Bot, message *tgbotapi.Message, accessInfo *models.AccessInfo) {
	switch message.Text {
	case "➕ Добавить клиента":
		handlers.HandleAddClient(b, message)
	case "👥 Мои клиенты":
		handlers.HandleListClients(b, message)
	case "📅 Групповые тренировки":
		handlers.HandleGroupTrainings(b, message)
	case "📅 Создать групповую":
		handlers.HandleCreateGroupTraining(b, message)
	case "📊 Статистика":
		handlers.HandleStats(b, message)
	case "🔙 Главное меню":
		handleStartCommand(b, message, accessInfo)
	default:
		b.SendMessage(message.Chat.ID, "Выберите действие из меню.")
	}
}

// handleDateCallback обрабатывает выбор даты тренировки.
func handleDateCallback(b *bot.Bot, callback *tgbotapi.CallbackQuery, action string, accessInfo *models.AccessInfo, chatID int64, messageID int) {
	state := b.GetState(callback.From.ID)
	if state == nil {
		return
	}

	switch action {
	case "cancel":
		b.CleanupMessages(chatID, callback.From.ID)
		b.ClearState(callback.From.ID)
		b.SendMessageWithKeyboard(chatID, "Отменено.", bot.GetStartMenuKeyboard(accessInfo))
	case "today":
		state.Data["workout_date"] = time.Now()
		b.CleanupMessages(chatID, callback.From.ID)
		b.SetState(callback.From.ID, "awaiting_muscle_group", state.Data)
		keyboard := bot.GetInlineMuscleGroupKeyboard()
		msgID := b.SendInlineKeyboard(chatID, "🏋️ Выберите группу мышц:", keyboard)
		b.StoreMessageID(callback.From.ID, msgID)
	case "other":
		state.Data["step"] = "date_custom"
		b.EditMessageText(chatID, messageID, "📅 Введите дату тренировки в формате ДД.ММ.ГГГГ\n(например: 15.01.2025):", nil)
	}
}

// handleWorkoutCallback обрабатывает просмотр и удаление тренировок.
func handleWorkoutCallback(b *bot.Bot, callback *tgbotapi.CallbackQuery, id int64, action string, chatID int64, messageID int) {
	state := b.GetState(callback.From.ID)

	switch action {
	case "close":
		b.DeleteMessage(chatID, messageID)
		b.ClearState(callback.From.ID)

	case "back":
		if state == nil {
			b.DeleteMessage(chatID, messageID)
			return
		}
		workouts, _ := state.Data["workouts"].([]*models.Workout)
		exerciseCounts, _ := state.Data["exercise_counts"].(map[int64]int)
		if len(workouts) == 0 {
			b.EditMessageText(chatID, messageID, "У вас пока нет тренировок.", nil)
			return
		}
		keyboard := bot.GetInlineWorkoutsKeyboard(workouts, exerciseCounts)
		title := "📝 *Ваши тренировки:*"
		if isTrainer, _ := state.Data["is_trainer_view"].(bool); isTrainer {
			if client, ok := state.Data["client"].(*models.ClientWithInfo); ok && client != nil {
				title = "📋 *История тренировок @" + bot.EscapeMarkdown(client.Client.Username) + "*"
			}
		}
		b.EditMessageText(chatID, messageID, title, &keyboard)

	case "detail":
		if state == nil {
			return
		}
		workout, exercises := getWorkoutWithExercises(b, id, state)
		if workout == nil {
			b.AnswerCallback(callback.ID, "Тренировка не найдена")
			return
		}
		text := handlers.FormatWorkoutDetail(workout, exercises)
		canDelete := true
		if isTrainer, _ := state.Data["is_trainer_view"].(bool); isTrainer {
			canDelete = false
		}
		keyboard := bot.GetInlineWorkoutActionsKeyboard(id, canDelete)
		b.EditMessageText(chatID, messageID, text, &keyboard)

	case "delete_ask":
		keyboard := bot.GetInlineWorkoutDeleteConfirmKeyboard(id)
		b.EditMessageText(chatID, messageID, "⚠️ Удалить эту тренировку? Это действие необратимо.", &keyboard)

	case "delete":
		if err := b.DB.DeleteWorkout(id); err != nil {
			log.Printf("Error deleting workout %d: %v", id, err)
			b.AnswerCallback(callback.ID, "Ошибка при удалении")
			return
		}
		// Убираем удалённую тренировку из state
		if state != nil {
			if workouts, ok := state.Data["workouts"].([]*models.Workout); ok {
				updated := make([]*models.Workout, 0, len(workouts))
				for _, w := range workouts {
					if w.ID != id {
						updated = append(updated, w)
					}
				}
				state.Data["workouts"] = updated
				if ec, ok := state.Data["exercise_counts"].(map[int64]int); ok {
					delete(ec, id)
				}
				if len(updated) == 0 {
					b.EditMessageText(chatID, messageID, "✅ Тренировка удалена. Тренировок больше нет.", nil)
					b.ClearState(callback.From.ID)
					return
				}
				exerciseCounts, _ := state.Data["exercise_counts"].(map[int64]int)
				keyboard := bot.GetInlineWorkoutsKeyboard(updated, exerciseCounts)
				b.EditMessageText(chatID, messageID, "✅ Тренировка удалена.\n\n📝 *Ваши тренировки:*", &keyboard)
			}
		}
	}
}

// getWorkoutWithExercises ищет тренировку по ID в state.Data и загружает упражнения.
func getWorkoutWithExercises(b *bot.Bot, id int64, state *models.UserState) (*models.Workout, []*models.Exercise) {
	var found *models.Workout
	if workouts, ok := state.Data["workouts"].([]*models.Workout); ok {
		for _, w := range workouts {
			if w.ID == id {
				found = w
				break
			}
		}
	}
	if found == nil {
		var err error
		found, err = b.DB.GetWorkoutByID(id)
		if err != nil {
			return nil, nil
		}
	}
	exercises, _ := b.DB.GetExercisesByWorkout(id)
	return found, exercises
}

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
		handleStartCommand(b, message, accessInfo)
	default:
		b.SendMessage(message.Chat.ID, "Выберите действие из меню.")
	}
}
