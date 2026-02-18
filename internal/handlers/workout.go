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
		accessInfo, _ := b.DB.GetUserAccessInfo(message.From.ID, message.From.UserName)
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
		b.SendMessage(message.Chat.ID, "⚠️ Выберите группу мышц из кнопок.")
		return
	}

	var trainerClientID *int64
	if state != nil && state.Data != nil {
		if tcID, ok := bot.GetStateInt64(state.Data, "trainer_client_id"); ok && tcID > 0 {
			trainerClientID = &tcID
		}
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
		"workout_id": workout.ID,
		"order":      1,
		"step":       "name",
	})

	b.SendMessageWithKeyboard(
		message.Chat.ID,
		"🏋️ Тренировка создана!\n\nВведите название первого упражнения:",
		bot.GetExerciseControlKeyboard(),
	)
}

// HandleExerciseName обрабатывает ввод названия упражнения.
// После получения названия показывает inline-клавиатуру для выбора подходов.
func HandleExerciseName(b *bot.Bot, message *tgbotapi.Message) {
	state := b.GetState(message.From.ID)
	if state == nil || state.Data == nil {
		b.SendMessage(message.Chat.ID, "❌ Ошибка. Начните тренировку заново.")
		return
	}

	if len(message.Photo) > 0 {
		b.SendMessage(message.Chat.ID, "📷 Сначала введите название упражнения текстом.")
		return
	}

	if message.Text == "✅ Завершить" {
		b.CleanupMessages(message.Chat.ID, message.From.ID)
		b.ClearState(message.From.ID)
		accessInfo, _ := b.DB.GetUserAccessInfo(message.From.ID, message.From.UserName)
		b.SendMessageWithKeyboard(message.Chat.ID, "✅ Тренировка сохранена! 💪", bot.GetStartMenuKeyboard(accessInfo))
		return
	}

	name := strings.TrimSpace(message.Text)
	if name == "" {
		b.SendMessage(message.Chat.ID, "Введите название упражнения:")
		return
	}

	state.Data["exercise_name"] = name
	state.Data["step"] = "sets"

	keyboard := bot.GetInlineSetsKeyboard()
	msgID := b.SendInlineKeyboard(
		message.Chat.ID,
		fmt.Sprintf("*%s*\n\nВыберите количество подходов:", bot.EscapeMarkdown(name)),
		keyboard,
	)
	b.StoreMessageID(message.From.ID, msgID)
}

// HandleExerciseSetsCustom обрабатывает ручной ввод количества подходов.
func HandleExerciseSetsCustom(b *bot.Bot, message *tgbotapi.Message) {
	state := b.GetState(message.From.ID)
	if state == nil {
		return
	}

	sets, err := strconv.Atoi(strings.TrimSpace(message.Text))
	if err != nil || sets <= 0 {
		b.SendMessage(message.Chat.ID, "⚠️ Введите целое положительное число:")
		return
	}

	name, _ := bot.GetStateString(state.Data, "exercise_name")
	state.Data["exercise_sets"] = sets
	state.Data["step"] = "reps"

	keyboard := bot.GetInlineRepsKeyboard()
	msgID := b.SendInlineKeyboard(
		message.Chat.ID,
		fmt.Sprintf("*%s* | Подходы: %d\n\nВыберите количество повторений:", bot.EscapeMarkdown(name), sets),
		keyboard,
	)
	b.StoreMessageID(message.From.ID, msgID)
}

// HandleExerciseRepsCustom обрабатывает ручной ввод количества повторений.
func HandleExerciseRepsCustom(b *bot.Bot, message *tgbotapi.Message) {
	state := b.GetState(message.From.ID)
	if state == nil {
		return
	}

	reps, err := strconv.Atoi(strings.TrimSpace(message.Text))
	if err != nil || reps <= 0 {
		b.SendMessage(message.Chat.ID, "⚠️ Введите целое положительное число:")
		return
	}

	name, _ := bot.GetStateString(state.Data, "exercise_name")
	sets, _ := bot.GetStateInt64(state.Data, "exercise_sets")
	state.Data["exercise_reps"] = reps
	state.Data["step"] = "weight"

	keyboard := bot.GetInlineWeightKeyboard()
	msgID := b.SendInlineKeyboard(
		message.Chat.ID,
		fmt.Sprintf("*%s* | Подходы: %d | Повт.: %d\n\nВыберите вес (кг):", bot.EscapeMarkdown(name), sets, reps),
		keyboard,
	)
	b.StoreMessageID(message.From.ID, msgID)
}

// HandleExerciseWeightCustom обрабатывает ручной ввод веса.
func HandleExerciseWeightCustom(b *bot.Bot, message *tgbotapi.Message) {
	state := b.GetState(message.From.ID)
	if state == nil {
		return
	}

	weight, err := strconv.ParseFloat(strings.TrimSpace(message.Text), 64)
	if err != nil || weight < 0 {
		b.SendMessage(message.Chat.ID, "⚠️ Введите вес числом (например: 80 или 72.5):")
		return
	}

	SaveExerciseStep(b, message.From.ID, message.Chat.ID, 0, weight, false)
}

// SaveExerciseStep сохраняет упражнение и показывает кнопки «Ещё» / «Завершить».
// Если editMsgID > 0, редактирует существующее сообщение; иначе отправляет новое.
func SaveExerciseStep(b *bot.Bot, userID, chatID int64, editMsgID int, weight float64, editMsg bool) {
	state := b.GetState(userID)
	if state == nil {
		return
	}

	name, _ := bot.GetStateString(state.Data, "exercise_name")
	setsVal, _ := bot.GetStateInt64(state.Data, "exercise_sets")
	repsVal, _ := bot.GetStateInt64(state.Data, "exercise_reps")
	workoutID, okW := bot.GetStateInt64(state.Data, "workout_id")
	if !okW {
		b.SendMessage(chatID, "❌ Ошибка. Начните тренировку заново.")
		return
	}

	order := 1
	if o, ok := state.Data["order"].(int); ok {
		order = o
	}

	exercise := &models.Exercise{
		WorkoutID: workoutID,
		Name:      name,
		Sets:      int(setsVal),
		Reps:      int(repsVal),
		Weight:    weight,
		Order:     order,
	}

	if err := b.DB.CreateExercise(exercise); err != nil {
		log.Printf("Error creating exercise: %v", err)
		b.SendMessage(chatID, "❌ Ошибка при сохранении упражнения.")
		return
	}

	state.Data["order"] = order + 1
	state.Data["step"] = "name"
	delete(state.Data, "exercise_name")
	delete(state.Data, "exercise_sets")
	delete(state.Data, "exercise_reps")

	text := fmt.Sprintf("✅ *%s* — %d×%d (%.1f кг)\n\nДобавить ещё упражнение?",
		bot.EscapeMarkdown(name), int(setsVal), int(repsVal), weight)
	keyboard := bot.GetInlineFinishKeyboard()

	if editMsg && editMsgID > 0 {
		b.EditMessageText(chatID, editMsgID, text, &keyboard)
	} else {
		msgID := b.SendInlineKeyboard(chatID, text, keyboard)
		b.StoreMessageID(userID, msgID)
	}
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
		response.WriteString(fmt.Sprintf("📅 %s — %s\n", w.Date.Format("02.01.2006"), w.MuscleGroup))

		if len(exercises) > 0 {
			for _, ex := range exercises {
				response.WriteString(fmt.Sprintf("  • %s: %d×%d (%.1f кг)\n",
					ex.Name, ex.Sets, ex.Reps, ex.Weight))
			}
		}
		response.WriteString("\n")
	}

	b.SendMessage(message.Chat.ID, response.String())
}
