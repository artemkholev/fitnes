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

	b, err := bot.NewBot(botToken, db)
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
	state := b.GetState(message.From.ID)

	user, _ := b.DB.GetUserByTelegramID(ctx, message.From.ID)

	if message.IsCommand() {
		switch message.Command() {
		case "start":
			handlers.HandleStart(b, message)
		case "creategroup":
			handlers.HandleCreateGroupTraining(b, message)
		default:
			b.SendMessage(message.Chat.ID, "Неизвестная команда")
		}
		return
	}

	if state != nil {
		switch state.State {
		case "awaiting_role":
			handlers.HandleRoleSelection(b, message)
		case "awaiting_org_code":
			handlers.HandleOrgCode(b, message)
		case "awaiting_muscle_group":
			handlers.HandleMuscleGroupSelection(b, message)
		case "adding_exercises":
			handlers.HandleAddExercise(b, message)
		case "selecting_trainer":
			if idx, err := strconv.Atoi(message.Text); err == nil {
				handlers.HandleTrainerSelection(b, message, idx)
			}
		case "joining_group_training":
			if idx, err := strconv.Atoi(message.Text); err == nil {
				handlers.HandleJoinGroupTraining(b, message, idx)
			}
		case "creating_group_training":
			handlers.HandleCreateGroupTrainingData(b, message)
		case "awaiting_exercise_name":
			handlers.HandleExerciseNameForStats(b, message)
		}
		return
	}

	if user == nil {
		handlers.HandleStart(b, message)
		return
	}

	switch message.Text {
	case "➕ Добавить тренировку", "➕ Создать тренировку":
		handlers.HandleAddWorkout(b, message)
	case "📝 Мои тренировки":
		handlers.HandleMyWorkouts(b, message)
	case "🔍 Найти тренера":
		handlers.HandleFindTrainer(b, message)
	case "👥 Мои клиенты":
		handlers.HandleMyClients(b, message)
	case "📅 Групповые тренировки":
		handlers.HandleGroupTrainings(b, message)
	case "📊 Моя статистика", "📊 Статистика":
		handlers.HandleStats(b, message)
	default:
		isTrainer := user.Role == models.RoleTrainer
		b.SendMessageWithKeyboard(
			message.Chat.ID,
			"Выберите действие из меню:",
			bot.GetMainMenuKeyboard(isTrainer),
		)
	}
}
