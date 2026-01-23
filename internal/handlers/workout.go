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

func HandleAddWorkout(b *bot.Bot, message *tgbotapi.Message) {
	// Получаем текущее состояние - там может быть trainer_client_id
	state := b.GetState(message.From.ID)
	var trainerClientID int64

	if state != nil && state.Data != nil {
		if tcID, ok := bot.GetStateInt64(state.Data, "trainer_client_id"); ok {
			trainerClientID = tcID
		}
	}

	b.SendMessageWithKeyboard(
		message.Chat.ID,
		"🏋️ *Новая тренировка*\n\nВыберите группу мышц:",
		bot.GetMuscleGroupKeyboard(),
	)
	b.SetState(message.From.ID, "awaiting_muscle_group", map[string]interface{}{
		"telegram_id":       message.From.ID,
		"trainer_client_id": trainerClientID,
	})
}

func HandleMuscleGroupSelection(b *bot.Bot, message *tgbotapi.Message) {

	if message.Text == "❌ Отмена" {
		b.CleanupMessages(message.Chat.ID, message.From.ID)
		b.ClearState(message.From.ID)
		accessInfo, _ := b.DB.GetUserAccessInfo( message.From.ID, message.From.UserName)
		b.SendMessageWithKeyboard(message.Chat.ID, "Отменено.", bot.GetStartMenuKeyboard(accessInfo))
		return
	}

	state := b.GetState(message.From.ID)

	muscleGroupMap := map[string]models.MuscleGroup{
		"💪 Грудь":   models.MuscleChest,
		"🦾 Спина":   models.MuscleBack,
		"🦵 Ноги":    models.MuscleLegs,
		"🏋️ Плечи":  models.MuscleShoulders,
		"💪 Бицепс":  models.MuscleBiceps,
		"💪 Трицепс": models.MuscleTriceps,
		"🎯 Пресс":   models.MuscleAbs,
		"🏃 Кардио":  models.MuscleCardio,
	}

	muscleGroup, ok := muscleGroupMap[message.Text]
	if !ok {
		b.SendMessageWithKeyboard(message.Chat.ID, "⚠️ Пожалуйста, выберите группу мышц из кнопок:", bot.GetMuscleGroupKeyboard())
		return
	}

	// Безопасное извлечение trainer_client_id (может быть int64 или *int64)
	var trainerClientID *int64
	if state != nil && state.Data != nil {
		// Пробуем как int64
		if tcID, ok := bot.GetStateInt64(state.Data, "trainer_client_id"); ok && tcID > 0 {
			trainerClientID = &tcID
		}
		// Пробуем как *int64
		if tcID, ok := state.Data["trainer_client_id"].(*int64); ok && tcID != nil {
			trainerClientID = tcID
		}
	}

	workout := &models.Workout{
		TrainerClientID:  trainerClientID,
		ClientTelegramID: message.From.ID,
		Date:             time.Now(),
		MuscleGroup:      muscleGroup,
	}

	if err := b.DB.CreateWorkout(workout); err != nil {
		log.Printf("Error creating workout (trainer_client_id=%v, telegram_id=%d): %v", trainerClientID, message.From.ID, err)
		b.SendMessage(message.Chat.ID, "❌ Ошибка при создании тренировки. Попробуйте позже.")
		return
	}

	b.SetState(message.From.ID, "adding_exercises", map[string]interface{}{
		"workout_id":  workout.ID,
		"telegram_id": message.From.ID,
		"order":       1,
	})

	breadcrumbs := bot.GetBreadcrumbs("🏠 Главная", "🏋️ Тренировки", "➕ Новая тренировка")
	text := breadcrumbs + "Добавьте упражнение в формате:\n"+
		"```\nНазвание\nПодходы\nПовторения\nВес (кг)\n```\n\n"+
		"Например:\n"+
		"```\nЖим лежа\n4\n10\n80\n```\n\n"+
		"Отправьте '✅ Завершить' когда закончите."

	b.SendMessageWithKeyboard(
		message.Chat.ID,
		text,
		bot.GetCancelKeyboard(),
	)
}

func HandleAddExercise(b *bot.Bot, message *tgbotapi.Message) {
	state := b.GetState(message.From.ID)

	if message.Text == "❌ Отмена" || message.Text == "✅ Завершить" {
		b.CleanupMessages(message.Chat.ID, message.From.ID)
		b.ClearState(message.From.ID)
		accessInfo, _ := b.DB.GetUserAccessInfo( message.From.ID, message.From.UserName)
		b.SendMessageWithKeyboard(message.Chat.ID, "✅ Тренировка сохранена! 💪", bot.GetStartMenuKeyboard(accessInfo))
		return
	}

	if state == nil || state.Data == nil {
		b.SendMessage(message.Chat.ID, "❌ Ошибка. Начните тренировку заново.")
		return
	}

	if len(message.Photo) > 0 {
		photos := message.Photo
		photoFileID := photos[len(photos)-1].FileID
		state.Data["photo_file_id"] = photoFileID
		b.SendMessage(message.Chat.ID, "📷 Фото сохранено! Теперь отправьте данные упражнения.")
		return
	}

	lines := strings.Split(strings.TrimSpace(message.Text), "\n")
	if len(lines) < 4 {
		b.SendMessage(message.Chat.ID, "❌ Неверный формат. Укажите:\n\nНазвание\nПодходы\nПовторения\nВес (кг)")
		return
	}

	name := strings.TrimSpace(lines[0])
	sets, err1 := strconv.Atoi(strings.TrimSpace(lines[1]))
	reps, err2 := strconv.Atoi(strings.TrimSpace(lines[2]))
	weight, err3 := strconv.ParseFloat(strings.TrimSpace(lines[3]), 64)

	if err1 != nil || err2 != nil || err3 != nil {
		b.SendMessage(message.Chat.ID, "❌ Ошибка в числовых значениях. Проверьте формат:\n\nНазвание\n4\n10\n80")
		return
	}

	workoutID, okW := bot.GetStateInt64(state.Data, "workout_id")
	if !okW {
		b.SendMessage(message.Chat.ID, "❌ Ошибка. Начните тренировку заново.")
		return
	}

	order := 1
	if o, ok := state.Data["order"].(int); ok {
		order = o
	}

	photoFileID := ""
	if photo, ok := state.Data["photo_file_id"].(string); ok {
		photoFileID = photo
		delete(state.Data, "photo_file_id")
	}

	exercise := &models.Exercise{
		WorkoutID:   workoutID,
		Name:        name,
		Sets:        sets,
		Reps:        reps,
		Weight:      weight,
		PhotoFileID: photoFileID,
		Order:       order,
	}

	if err := b.DB.CreateExercise(exercise); err != nil {
		log.Printf("Error creating exercise: %v", err)
		b.SendMessage(message.Chat.ID, "❌ Ошибка при сохранении упражнения.")
		return
	}

	state.Data["order"] = order + 1
	b.SendMessage(message.Chat.ID, fmt.Sprintf("✅ Упражнение '%s' добавлено!\n\nДобавьте ещё одно или отправьте '✅ Завершить'", name))
}

func HandleMyWorkouts(b *bot.Bot, message *tgbotapi.Message) {
	workouts, err := b.DB.GetWorkoutsByClientTelegramID(message.From.ID, 10)
	if err != nil {
		log.Printf("Error getting workouts: %v", err)
		b.SendMessage(message.Chat.ID, "❌ Ошибка при получении тренировок.")
		return
	}

	if len(workouts) == 0 {
		b.SendMessage(message.Chat.ID, "У вас пока нет тренировок. Добавьте первую!")
		return
	}

	var response strings.Builder
	response.WriteString("📝 *Ваши последние тренировки:*\n\n")

	for _, w := range workouts {
		exercises, _ := b.DB.GetExercisesByWorkout(w.ID)
		response.WriteString(fmt.Sprintf("📅 %s - %s\n", w.Date.Format("02.01.2006"), w.MuscleGroup))

		if len(exercises) > 0 {
			for _, ex := range exercises {
				response.WriteString(fmt.Sprintf("  • %s: %d x %d (%.1f кг)\n",
					ex.Name, ex.Sets, ex.Reps, ex.Weight))
			}
		}
		response.WriteString("\n")
	}

	b.SendMessage(message.Chat.ID, response.String())
}
