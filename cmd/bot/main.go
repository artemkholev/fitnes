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

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
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
		if update.Message == nil {
			continue
		}

		go handleUpdate(b, update.Message)
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
		if idx, err := strconv.Atoi(message.Text); err == nil {
			handlers.HandleSelectOrganization(b, message, idx)
		}
	case "admin_managing_org":
		handleAdminOrgActions(b, message, accessInfo)
	case "admin_adding_manager":
		handlers.HandleAddManagerUsername(b, message)
	case "admin_removing_manager":
		if idx, err := strconv.Atoi(message.Text); err == nil {
			handlers.HandleRemoveManager(b, message, idx)
		}

	// ===== МЕНЕДЖЕР =====
	case "manager_selecting_org":
		if idx, err := strconv.Atoi(message.Text); err == nil {
			handlers.HandleManagerSelectOrg(b, message, idx)
		}
	case "manager_managing_org":
		handleManagerOrgActions(b, message, accessInfo)
	case "manager_adding_trainer":
		handlers.HandleAddTrainerUsername(b, message)
	case "manager_removing_trainer":
		if idx, err := strconv.Atoi(message.Text); err == nil {
			handlers.HandleRemoveTrainer(b, message, idx)
		}

	// ===== ТРЕНЕР =====
	case "trainer_selecting_org":
		if idx, err := strconv.Atoi(message.Text); err == nil {
			handlers.HandleTrainerSelectOrg(b, message, idx)
		}
	case "trainer_managing_org":
		handleTrainerOrgActions(b, message, accessInfo)
	case "trainer_adding_client":
		handlers.HandleAddClientUsername(b, message)
	case "trainer_viewing_clients":
		text := message.Text
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
		}
	case "trainer_client_action":
		if idx, err := strconv.Atoi(message.Text); err == nil {
			handlers.HandleClientAction(b, message, idx)
		}

	// ===== КЛИЕНТ =====
	case "client_selecting_trainer":
		if idx, err := strconv.Atoi(message.Text); err == nil {
			handlers.HandleClientSelectTrainer(b, message, idx)
		}
	case "client_with_trainer":
		handleClientActions(b, message, accessInfo)
	case "client_viewing_archive":
		if idx, err := strconv.Atoi(message.Text); err == nil {
			handlers.HandleSelectArchivedTrainer(b, message, idx)
		}

	// ===== ТРЕНИРОВКИ =====
	case "awaiting_muscle_group", "creating_workout_for_client":
		handlers.HandleMuscleGroupSelection(b, message)
	case "adding_exercises":
		handlers.HandleAddExercise(b, message)
	case "awaiting_exercise_name":
		handlers.HandleExerciseNameForStats(b, message)

	// ===== ГРУППОВЫЕ ТРЕНИРОВКИ =====
	case "joining_group_training":
		if idx, err := strconv.Atoi(message.Text); err == nil {
			handlers.HandleJoinGroupTraining(b, message, idx)
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
