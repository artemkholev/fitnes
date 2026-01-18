package handlers

import (
	"context"
	"fitness-bot/internal/bot"
	"fitness-bot/internal/models"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleGroupTrainings(b *bot.Bot, message *tgbotapi.Message) {
	ctx := context.Background()

	user, err := b.DB.GetUserByTelegramID(ctx, message.From.ID)
	if err != nil {
		b.SendMessage(message.Chat.ID, "Ошибка при получении данных.")
		return
	}

	if user.OrganizationID == nil {
		b.SendMessage(message.Chat.ID, "Вы не привязаны к организации.")
		return
	}

	trainings, err := b.DB.GetUpcomingGroupTrainings(ctx, *user.OrganizationID)
	if err != nil {
		log.Printf("Error getting group trainings: %v", err)
		b.SendMessage(message.Chat.ID, "Ошибка при получении тренировок.")
		return
	}

	if len(trainings) == 0 {
		if user.Role == models.RoleTrainer {
			b.SendMessage(message.Chat.ID, "Пока нет запланированных групповых тренировок.\n\nЧтобы создать, отправьте:\n/creategroup")
		} else {
			b.SendMessage(message.Chat.ID, "Пока нет запланированных групповых тренировок.")
		}
		return
	}

	var response strings.Builder
	response.WriteString("📅 Предстоящие групповые тренировки:\n\n")

	for i, training := range trainings {
		count, _ := b.DB.GetParticipantCount(ctx, training.ID)
		response.WriteString(fmt.Sprintf("%d. %s\n", i+1, training.Name))
		response.WriteString(fmt.Sprintf("   📝 %s\n", training.Description))
		response.WriteString(fmt.Sprintf("   📅 %s\n", training.ScheduledAt.Format("02.01.2006 15:04")))
		response.WriteString(fmt.Sprintf("   👥 %d/%d участников\n\n", count, training.MaxParticipants))
	}

	if user.Role == models.RoleClient {
		response.WriteString("Чтобы записаться, отправьте номер тренировки.")
		b.SetState(message.From.ID, "joining_group_training", map[string]interface{}{
			"trainings": trainings,
			"user_id":   user.ID,
		})
	}

	b.SendMessage(message.Chat.ID, response.String())
}

func HandleJoinGroupTraining(b *bot.Bot, message *tgbotapi.Message, trainingIdx int) {
	ctx := context.Background()
	state := b.GetState(message.From.ID)

	trainings := state.Data["trainings"].([]*models.GroupTraining)
	if trainingIdx < 1 || trainingIdx > len(trainings) {
		b.SendMessage(message.Chat.ID, "Неверный номер. Попробуйте ещё раз.")
		return
	}

	training := trainings[trainingIdx-1]
	userID := state.Data["user_id"].(int64)

	count, _ := b.DB.GetParticipantCount(ctx, training.ID)
	if count >= training.MaxParticipants {
		b.SendMessage(message.Chat.ID, "К сожалению, все места заняты.")
		b.ClearState(message.From.ID)
		return
	}

	if err := b.DB.JoinGroupTraining(ctx, training.ID, userID); err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			b.SendMessage(message.Chat.ID, "Вы уже записаны на эту тренировку.")
		} else {
			log.Printf("Error joining training: %v", err)
			b.SendMessage(message.Chat.ID, "Ошибка при записи.")
		}
		return
	}

	b.ClearState(message.From.ID)
	user, _ := b.DB.GetUserByTelegramID(ctx, message.From.ID)
	b.SendMessageWithKeyboard(
		message.Chat.ID,
		fmt.Sprintf("✅ Вы записаны на тренировку '%s'!", training.Name),
		bot.GetMainMenuKeyboard(false),
	)

	trainer, _ := b.DB.GetUserByTelegramID(ctx, 0)
	if trainer != nil {
		notif := fmt.Sprintf("🔔 Новый участник %s записался на '%s'", user.FullName, training.Name)
		b.SendMessage(trainer.TelegramID, notif)
	}
}

func HandleCreateGroupTraining(b *bot.Bot, message *tgbotapi.Message) {
	ctx := context.Background()

	user, err := b.DB.GetUserByTelegramID(ctx, message.From.ID)
	if err != nil || user.Role != models.RoleTrainer {
		b.SendMessage(message.Chat.ID, "Только тренеры могут создавать групповые тренировки.")
		return
	}

	if user.OrganizationID == nil {
		b.SendMessage(message.Chat.ID, "Вы не привязаны к организации.")
		return
	}

	b.SendMessageWithKeyboard(
		message.Chat.ID,
		"Создание групповой тренировки.\n\nОтправьте данные в формате:\n"+
			"Название\nОписание\nДата и время (ДД.ММ.ГГГГ ЧЧ:ММ)\nМакс. участников\n\n"+
			"Например:\nФункциональный тренинг\nИнтенсивная тренировка\n25.01.2026 18:00\n15",
		bot.GetCancelKeyboard(),
	)
	b.SetState(message.From.ID, "creating_group_training", map[string]interface{}{
		"org_id":     *user.OrganizationID,
		"trainer_id": user.ID,
	})
}

func HandleCreateGroupTrainingData(b *bot.Bot, message *tgbotapi.Message) {
	ctx := context.Background()
	state := b.GetState(message.From.ID)

	if message.Text == "❌ Отмена" {
		b.ClearState(message.From.ID)
		user, _ := b.DB.GetUserByTelegramID(ctx, message.From.ID)
		b.SendMessageWithKeyboard(message.Chat.ID, "Отменено.", bot.GetMainMenuKeyboard(user.Role == models.RoleTrainer))
		return
	}

	lines := strings.Split(strings.TrimSpace(message.Text), "\n")
	if len(lines) < 4 {
		b.SendMessage(message.Chat.ID, "Неверный формат. Проверьте данные.")
		return
	}

	name := strings.TrimSpace(lines[0])
	description := strings.TrimSpace(lines[1])
	dateStr := strings.TrimSpace(lines[2])
	maxParticipants, err := strconv.Atoi(strings.TrimSpace(lines[3]))

	if err != nil {
		b.SendMessage(message.Chat.ID, "Ошибка в количестве участников.")
		return
	}

	scheduledAt, err := time.Parse("02.01.2006 15:04", dateStr)
	if err != nil {
		b.SendMessage(message.Chat.ID, "Ошибка в формате даты. Используйте ДД.ММ.ГГГГ ЧЧ:ММ")
		return
	}

	training := &models.GroupTraining{
		OrganizationID:  state.Data["org_id"].(int64),
		TrainerID:       state.Data["trainer_id"].(int64),
		Name:            name,
		Description:     description,
		ScheduledAt:     scheduledAt,
		MaxParticipants: maxParticipants,
	}

	if err := b.DB.CreateGroupTraining(ctx, training); err != nil {
		log.Printf("Error creating group training: %v", err)
		b.SendMessage(message.Chat.ID, "Ошибка при создании тренировки.")
		return
	}

	b.ClearState(message.From.ID)
	user, _ := b.DB.GetUserByTelegramID(ctx, message.From.ID)
	b.SendMessageWithKeyboard(
		message.Chat.ID,
		fmt.Sprintf("✅ Групповая тренировка '%s' создана!", name),
		bot.GetMainMenuKeyboard(user.Role == models.RoleTrainer),
	)
}
