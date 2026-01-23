package handlers

import (
	"fitness-bot/internal/bot"
	"fitness-bot/internal/charts"
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleStats(b *bot.Bot, message *tgbotapi.Message) {
	b.SendMessageWithKeyboard(
		message.Chat.ID,
		"Введите название упражнения для просмотра прогресса:\n\nНапример: Жим лежа",
		bot.GetCancelKeyboard(),
	)
	b.SetState(message.From.ID, "awaiting_exercise_name", map[string]interface{}{
		"telegram_id": message.From.ID,
	})
}

func HandleExerciseNameForStats(b *bot.Bot, message *tgbotapi.Message) {
	state := b.GetState(message.From.ID)

	if message.Text == "❌ Отмена" {
		b.ClearState(message.From.ID)
		// Определяем клавиатуру на основе доступов
		accessInfo, _ := b.DB.GetUserAccessInfo( message.From.ID, message.From.UserName)
		b.SendMessageWithKeyboard(message.Chat.ID, "Отменено.", bot.GetStartMenuKeyboard(accessInfo))
		return
	}

	exerciseName := message.Text
	telegramID := state.Data["telegram_id"].(int64)

	from := time.Now().AddDate(0, -3, 0)
	to := time.Now()

	exercises, err := b.DB.GetExerciseStats(telegramID, exerciseName, from, to)
	if err != nil {
		log.Printf("Error getting exercise stats: %v", err)
		b.SendMessage(message.Chat.ID, "❌ Ошибка при получении статистики.")
		return
	}

	if len(exercises) == 0 {
		b.SendMessage(message.Chat.ID, fmt.Sprintf("Упражнение '%s' не найдено в ваших тренировках.", exerciseName))
		b.ClearState(message.From.ID)
		return
	}

	chartData, err := charts.GenerateProgressChart(exercises, exerciseName)
	if err != nil {
		log.Printf("Error generating chart: %v", err)
		b.SendMessage(message.Chat.ID, "❌ Ошибка при создании графика.")
		return
	}

	if chartData != nil {
		photoBytes := tgbotapi.FileBytes{
			Name:  "progress.png",
			Bytes: chartData,
		}
		photo := tgbotapi.NewPhoto(message.Chat.ID, photoBytes)
		photo.Caption = fmt.Sprintf("📊 Прогресс по упражнению '%s' за последние 3 месяца", exerciseName)
		b.API.Send(photo)
	}

	var statsText string
	if len(exercises) > 0 {
		latest := exercises[0]
		statsText = fmt.Sprintf("\n📈 *Последний результат:*\n"+
			"Вес: %.1f кг\n"+
			"Подходы: %d\n"+
			"Повторения: %d\n"+
			"Дата: %s",
			latest.Weight,
			latest.Sets,
			latest.Reps,
			latest.CreatedAt.Format("02.01.2006"),
		)
	}

	b.ClearState(message.From.ID)
	accessInfo, _ := b.DB.GetUserAccessInfo( message.From.ID, message.From.UserName)
	b.SendMessageWithKeyboard(message.Chat.ID, statsText, bot.GetStartMenuKeyboard(accessInfo))
}
