package handlers

import (
	"fitness-bot/internal/bot"
	"fitness-bot/internal/models"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleGroupTrainings показывает групповые тренировки
func HandleGroupTrainings(b *bot.Bot, message *tgbotapi.Message) {

	// Получаем доступы пользователя
	accessInfo, err := b.DB.GetUserAccessInfo( message.From.ID, message.From.UserName)
	if err != nil {
		b.SendMessage(message.Chat.ID, "❌ Ошибка при получении данных.")
		return
	}

	// Определяем организацию из текущего состояния
	state := b.GetState(message.From.ID)
	var orgID int64

	if state != nil && state.Data["org_id"] != nil {
		orgID = state.Data["org_id"].(int64)
	} else if len(accessInfo.TrainerOrgs) > 0 {
		orgID = accessInfo.TrainerOrgs[0].Organization.ID
	} else if len(accessInfo.ClientAccess) > 0 {
		orgID = accessInfo.ClientAccess[0].OrganizationID
	} else {
		b.SendMessage(message.Chat.ID, "❌ У вас нет доступа к организациям.")
		return
	}

	trainings, err := b.DB.GetUpcomingGroupTrainings(orgID)
	if err != nil {
		log.Printf("Error getting group trainings: %v", err)
		b.SendMessage(message.Chat.ID, "❌ Ошибка при получении тренировок.")
		return
	}

	if len(trainings) == 0 {
		isTrainer := len(accessInfo.TrainerOrgs) > 0
		if isTrainer {
			b.SendMessage(message.Chat.ID, "Пока нет запланированных групповых тренировок.\n\nЧтобы создать, используйте команду в панели тренера.")
		} else {
			b.SendMessage(message.Chat.ID, "Пока нет запланированных групповых тренировок.")
		}
		return
	}

	var response strings.Builder
	response.WriteString("📅 *Предстоящие групповые тренировки:*\n\n")

	for i, training := range trainings {
		count, _ := b.DB.GetParticipantCount(training.ID)
		response.WriteString(fmt.Sprintf("%d. *%s*\n", i+1, training.Name))
		response.WriteString(fmt.Sprintf("   📝 %s\n", training.Description))
		response.WriteString(fmt.Sprintf("   📅 %s\n", training.ScheduledAt.Format("02.01.2006 15:04")))
		response.WriteString(fmt.Sprintf("   👥 %d/%d участников\n\n", count, training.MaxParticipants))
	}

	// Клиенты могут записываться
	if len(accessInfo.ClientAccess) > 0 {
		response.WriteString("Чтобы записаться, отправьте номер тренировки.")

		user, _ := b.DB.GetUserByTelegramID(message.From.ID)
		userID := int64(0)
		if user != nil {
			userID = user.ID
		}

		b.SetState(message.From.ID, "joining_group_training", map[string]interface{}{
			"trainings": trainings,
			"user_id":   userID,
		})
	}

	b.SendMessage(message.Chat.ID, response.String())
}

// HandleJoinGroupTraining записывает пользователя на групповую тренировку
func HandleJoinGroupTraining(b *bot.Bot, message *tgbotapi.Message, trainingIdx int) {
	state := b.GetState(message.From.ID)

	trainings := state.Data["trainings"].([]*models.GroupTraining)
	if trainingIdx < 1 || trainingIdx > len(trainings) {
		b.SendMessage(message.Chat.ID, "❌ Неверный номер. Попробуйте ещё раз.")
		return
	}

	training := trainings[trainingIdx-1]
	userID := state.Data["user_id"].(int64)

	count, _ := b.DB.GetParticipantCount(training.ID)
	if count >= training.MaxParticipants {
		b.SendMessage(message.Chat.ID, "❌ К сожалению, все места заняты.")
		b.ClearState(message.From.ID)
		return
	}

	if err := b.DB.JoinGroupTraining(training.ID, userID); err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			b.SendMessage(message.Chat.ID, "Вы уже записаны на эту тренировку.")
		} else {
			log.Printf("Error joining training: %v", err)
			b.SendMessage(message.Chat.ID, "❌ Ошибка при записи.")
		}
		return
	}

	b.ClearState(message.From.ID)
	b.SendMessageWithKeyboard(
		message.Chat.ID,
		fmt.Sprintf("✅ Вы записаны на тренировку '%s'!", training.Name),
		bot.GetClientMenuKeyboard(),
	)
}

// HandleCreateGroupTraining начинает создание групповой тренировки (для тренеров)
func HandleCreateGroupTraining(b *bot.Bot, message *tgbotapi.Message) {

	// Проверяем что пользователь тренер
	accessInfo, err := b.DB.GetUserAccessInfo( message.From.ID, message.From.UserName)
	if err != nil || len(accessInfo.TrainerOrgs) == 0 {
		b.SendMessage(message.Chat.ID, "❌ Только тренеры могут создавать групповые тренировки.")
		return
	}

	state := b.GetState(message.From.ID)
	var orgID, trainerID int64

	if state != nil && state.Data["org_id"] != nil {
		orgID = state.Data["org_id"].(int64)
		trainerID = state.Data["trainer_id"].(int64)
	} else {
		// Берём первую активную организацию
		for _, org := range accessInfo.TrainerOrgs {
			if org.IsActive {
				orgID = org.Organization.ID
				trainerID = org.TrainerID
				break
			}
		}
	}

	if orgID == 0 {
		b.SendMessage(message.Chat.ID, "❌ Нет доступных организаций.")
		return
	}

	b.SendMessageWithKeyboard(
		message.Chat.ID,
		"*Создание групповой тренировки*\n\nОтправьте данные в формате:\n"+
			"Название\nОписание\nДата и время (ДД.ММ.ГГГГ ЧЧ:ММ)\nМакс. участников\n\n"+
			"Например:\nФункциональный тренинг\nИнтенсивная тренировка\n25.01.2026 18:00\n15",
		bot.GetCancelKeyboard(),
	)
	b.SetState(message.From.ID, "creating_group_training", map[string]interface{}{
		"org_id":     orgID,
		"trainer_id": trainerID,
	})
}

// HandleCreateGroupTrainingData обрабатывает данные для создания групповой тренировки
func HandleCreateGroupTrainingData(b *bot.Bot, message *tgbotapi.Message) {
	state := b.GetState(message.From.ID)

	if message.Text == "❌ Отмена" {
		b.ClearState(message.From.ID)
		b.SendMessageWithKeyboard(message.Chat.ID, "Отменено.", bot.GetTrainerMenuKeyboard())
		return
	}

	lines := strings.Split(strings.TrimSpace(message.Text), "\n")
	if len(lines) < 4 {
		b.SendMessage(message.Chat.ID, "❌ Неверный формат. Проверьте данные.")
		return
	}

	name := strings.TrimSpace(lines[0])
	description := strings.TrimSpace(lines[1])
	dateStr := strings.TrimSpace(lines[2])
	maxParticipants, err := strconv.Atoi(strings.TrimSpace(lines[3]))

	if err != nil {
		b.SendMessage(message.Chat.ID, "❌ Ошибка в количестве участников.")
		return
	}

	scheduledAt, err := time.Parse("02.01.2006 15:04", dateStr)
	if err != nil {
		b.SendMessage(message.Chat.ID, "❌ Ошибка в формате даты. Используйте ДД.ММ.ГГГГ ЧЧ:ММ")
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

	if err := b.DB.CreateGroupTraining(training); err != nil {
		log.Printf("Error creating group training: %v", err)
		b.SendMessage(message.Chat.ID, "❌ Ошибка при создании тренировки.")
		return
	}

	b.ClearState(message.From.ID)
	b.SendMessageWithKeyboard(
		message.Chat.ID,
		fmt.Sprintf("✅ Групповая тренировка '%s' создана!", name),
		bot.GetTrainerMenuKeyboard(),
	)
}
