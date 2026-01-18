package handlers

import (
	"context"
	"fitness-bot/internal/bot"
	"fitness-bot/internal/database"
	"fitness-bot/internal/models"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5"
)

func HandleStart(b *bot.Bot, message *tgbotapi.Message) {
	ctx := context.Background()

	user, err := b.DB.GetUserByTelegramID(ctx, message.From.ID)
	if err != nil && err != pgx.ErrNoRows {
		log.Printf("Error getting user: %v", err)
		b.SendMessage(message.Chat.ID, "Произошла ошибка. Попробуйте позже.")
		return
	}

	if user != nil {
		isTrainer := user.Role == models.RoleTrainer
		b.SendMessageWithKeyboard(
			message.Chat.ID,
			"Добро пожаловать! Выберите действие:",
			bot.GetMainMenuKeyboard(isTrainer),
		)
		return
	}

	b.SendMessageWithKeyboard(
		message.Chat.ID,
		"Привет! Я помогу тебе отслеживать тренировки.\n\nВыбери свою роль:",
		bot.GetRoleKeyboard(),
	)
	b.SetState(message.From.ID, "awaiting_role", nil)
}

func HandleRoleSelection(b *bot.Bot, message *tgbotapi.Message) {
	ctx := context.Background()

	var role models.UserRole
	switch message.Text {
	case "👤 Клиент":
		role = models.RoleClient
	case "💼 Тренер":
		role = models.RoleTrainer
	default:
		b.SendMessage(message.Chat.ID, "Пожалуйста, выберите роль из предложенных вариантов.")
		return
	}

	b.SendMessageWithKeyboard(
		message.Chat.ID,
		"Введите код организации (если есть) или отправьте '-' чтобы пропустить:",
		bot.GetCancelKeyboard(),
	)
	b.SetState(message.From.ID, "awaiting_org_code", map[string]interface{}{
		"role": role,
	})
}

func HandleOrgCode(b *bot.Bot, message *tgbotapi.Message) {
	ctx := context.Background()
	state := b.GetState(message.From.ID)

	if message.Text == "❌ Отмена" {
		b.ClearState(message.From.ID)
		HandleStart(b, message)
		return
	}

	role := state.Data["role"].(models.UserRole)
	var orgID *int64

	if message.Text != "-" {
		org, err := b.DB.GetOrganizationByCode(ctx, message.Text)
		if err != nil {
			b.SendMessage(message.Chat.ID, "Организация с таким кодом не найдена. Попробуйте ещё раз или отправьте '-'")
			return
		}
		orgID = &org.ID
	}

	username := message.From.UserName
	if username == "" {
		username = "user_" + string(rune(message.From.ID))
	}

	fullName := message.From.FirstName
	if message.From.LastName != "" {
		fullName += " " + message.From.LastName
	}

	user := &models.User{
		TelegramID:     message.From.ID,
		Username:       username,
		FullName:       fullName,
		Role:           role,
		OrganizationID: orgID,
	}

	if err := b.DB.CreateUser(ctx, user); err != nil {
		log.Printf("Error creating user: %v", err)
		b.SendMessage(message.Chat.ID, "Ошибка при регистрации. Попробуйте позже.")
		return
	}

	b.ClearState(message.From.ID)
	isTrainer := role == models.RoleTrainer

	var welcomeMsg string
	if isTrainer {
		welcomeMsg = "Регистрация успешна! Вы зарегистрированы как тренер."
	} else {
		welcomeMsg = "Регистрация успешна! Вы зарегистрированы как клиент."
	}

	b.SendMessageWithKeyboard(
		message.Chat.ID,
		welcomeMsg,
		bot.GetMainMenuKeyboard(isTrainer),
	)
}
