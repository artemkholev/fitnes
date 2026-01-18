package handlers

import (
	"context"
	"fitness-bot/internal/bot"
	"fitness-bot/internal/models"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleFindTrainer(b *bot.Bot, message *tgbotapi.Message) {
	ctx := context.Background()

	user, err := b.DB.GetUserByTelegramID(ctx, message.From.ID)
	if err != nil {
		b.SendMessage(message.Chat.ID, "Ошибка при получении данных.")
		return
	}

	if user.OrganizationID == nil {
		b.SendMessage(message.Chat.ID, "Вы не привязаны к организации. Обратитесь к администратору.")
		return
	}

	trainers, err := b.DB.GetTrainersByOrganization(ctx, *user.OrganizationID)
	if err != nil {
		log.Printf("Error getting trainers: %v", err)
		b.SendMessage(message.Chat.ID, "Ошибка при поиске тренеров.")
		return
	}

	if len(trainers) == 0 {
		b.SendMessage(message.Chat.ID, "В вашей организации пока нет тренеров.")
		return
	}

	var response strings.Builder
	response.WriteString("👨‍🏫 Доступные тренеры:\n\n")

	for i, trainer := range trainers {
		response.WriteString(fmt.Sprintf("%d. %s (@%s)\n", i+1, trainer.FullName, trainer.Username))
	}

	response.WriteString("\nЧтобы выбрать тренера, отправьте его номер.")

	b.SendMessage(message.Chat.ID, response.String())
	b.SetState(message.From.ID, "selecting_trainer", map[string]interface{}{
		"trainers": trainers,
		"user_id":  user.ID,
	})
}

func HandleTrainerSelection(b *bot.Bot, message *tgbotapi.Message, trainerIdx int) {
	ctx := context.Background()
	state := b.GetState(message.From.ID)

	trainers := state.Data["trainers"].([]*models.User)
	if trainerIdx < 1 || trainerIdx > len(trainers) {
		b.SendMessage(message.Chat.ID, "Неверный номер. Попробуйте ещё раз.")
		return
	}

	trainer := trainers[trainerIdx-1]
	userID := state.Data["user_id"].(int64)

	if err := b.DB.UpdateUserTrainer(ctx, userID, trainer.ID); err != nil {
		log.Printf("Error updating trainer: %v", err)
		b.SendMessage(message.Chat.ID, "Ошибка при привязке к тренеру.")
		return
	}

	b.ClearState(message.From.ID)
	b.SendMessageWithKeyboard(
		message.Chat.ID,
		fmt.Sprintf("✅ Вы успешно привязались к тренеру %s!", trainer.FullName),
		bot.GetMainMenuKeyboard(false),
	)
}

func HandleMyClients(b *bot.Bot, message *tgbotapi.Message) {
	ctx := context.Background()

	user, err := b.DB.GetUserByTelegramID(ctx, message.From.ID)
	if err != nil {
		b.SendMessage(message.Chat.ID, "Ошибка при получении данных.")
		return
	}

	if user.Role != models.RoleTrainer {
		b.SendMessage(message.Chat.ID, "Эта функция доступна только тренерам.")
		return
	}

	clients, err := b.DB.GetClientsByTrainer(ctx, user.ID)
	if err != nil {
		log.Printf("Error getting clients: %v", err)
		b.SendMessage(message.Chat.ID, "Ошибка при получении клиентов.")
		return
	}

	if len(clients) == 0 {
		b.SendMessage(message.Chat.ID, "У вас пока нет клиентов.")
		return
	}

	var response strings.Builder
	response.WriteString("👥 Ваши клиенты:\n\n")

	for i, client := range clients {
		response.WriteString(fmt.Sprintf("%d. %s (@%s)\n", i+1, client.FullName, client.Username))
	}

	b.SendMessage(message.Chat.ID, response.String())
}
