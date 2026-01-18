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
	"github.com/jackc/pgx/v5"
)

func HandleAddWorkout(b *bot.Bot, message *tgbotapi.Message) {
	ctx := context.Background()

	user, err := b.DB.GetUserByTelegramID(ctx, message.From.ID)
	if err != nil {
		if err == pgx.ErrNoRows {
			b.SendMessage(message.Chat.ID, "Пожалуйста, сначала зарегистрируйтесь командой /start")
			return
		}
		log.Printf("Error getting user: %v", err)
		b.SendMessage(message.Chat.ID, "Произошла ошибка.")
		return
	}

	b.SendMessageWithKeyboard(
		message.Chat.ID,
		"Выберите группу мышц для тренировки:",
		bot.GetMuscleGroupKeyboard(),
	)
	b.SetState(message.From.ID, "awaiting_muscle_group", map[string]interface{}{
		"user_id": user.ID,
	})
}

func HandleMuscleGroupSelection(b *bot.Bot, message *tgbotapi.Message) {
	if message.Text == "❌ Отмена" {
		b.ClearState(message.From.ID)
		ctx := context.Background()
		user, _ := b.DB.GetUserByTelegramID(ctx, message.From.ID)
		isTrainer := user.Role == models.RoleTrainer
		b.SendMessageWithKeyboard(message.Chat.ID, "Отменено.", bot.GetMainMenuKeyboard(isTrainer))
		return
	}

	state := b.GetState(message.From.ID)
	userID := state.Data["user_id"].(int64)

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
		b.SendMessage(message.Chat.ID, "Пожалуйста, выберите группу мышц из списка.")
		return
	}

	ctx := context.Background()
	workout := &models.Workout{
		UserID:      userID,
		Date:        time.Now(),
		MuscleGroup: muscleGroup,
	}

	if err := b.DB.CreateWorkout(ctx, workout); err != nil {
		log.Printf("Error creating workout: %v", err)
		b.SendMessage(message.Chat.ID, "Ошибка при создании тренировки.")
		return
	}

	b.SetState(message.From.ID, "adding_exercises", map[string]interface{}{
		"workout_id": workout.ID,
		"user_id":    userID,
		"order":      1,
	})

	b.SendMessageWithKeyboard(
		message.Chat.ID,
		"Тренировка создана! ✅\n\nТеперь добавьте упражнение в формате:\n"+
			"Название\nПодходы\nПовторения\nВес (кг)\n\n"+
			"Например:\n"+
			"Жим лежа\n4\n10\n80\n\n"+
			"Отправьте '✅ Завершить' когда закончите.",
		bot.GetCancelKeyboard(),
	)
}

func HandleAddExercise(b *bot.Bot, message *tgbotapi.Message) {
	ctx := context.Background()
	state := b.GetState(message.From.ID)

	if message.Text == "❌ Отмена" || message.Text == "✅ Завершить" {
		b.ClearState(message.From.ID)
		user, _ := b.DB.GetUserByTelegramID(ctx, message.From.ID)
		isTrainer := user.Role == models.RoleTrainer
		b.SendMessageWithKeyboard(message.Chat.ID, "Тренировка сохранена! 💪", bot.GetMainMenuKeyboard(isTrainer))
		return
	}

	if len(message.Photo) > 0 {
		photos := message.Photo
		photoFileID := photos[len(photos)-1].FileID
		state.Data["photo_file_id"] = photoFileID
		b.SendMessage(message.Chat.ID, "Фото сохранено! Теперь отправьте данные упражнения.")
		return
	}

	lines := strings.Split(strings.TrimSpace(message.Text), "\n")
	if len(lines) < 4 {
		b.SendMessage(message.Chat.ID, "Неверный формат. Пожалуйста, укажите:\nНазвание\nПодходы\nПовторения\nВес")
		return
	}

	name := strings.TrimSpace(lines[0])
	sets, err1 := strconv.Atoi(strings.TrimSpace(lines[1]))
	reps, err2 := strconv.Atoi(strings.TrimSpace(lines[2]))
	weight, err3 := strconv.ParseFloat(strings.TrimSpace(lines[3]), 64)

	if err1 != nil || err2 != nil || err3 != nil {
		b.SendMessage(message.Chat.ID, "Ошибка в числовых значениях. Проверьте формат.")
		return
	}

	workoutID := state.Data["workout_id"].(int64)
	order := state.Data["order"].(int)

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

	if err := b.DB.CreateExercise(ctx, exercise); err != nil {
		log.Printf("Error creating exercise: %v", err)
		b.SendMessage(message.Chat.ID, "Ошибка при сохранении упражнения.")
		return
	}

	state.Data["order"] = order + 1
	b.SendMessage(message.Chat.ID, fmt.Sprintf("✅ Упражнение '%s' добавлено!\n\nДобавьте ещё одно или отправьте '✅ Завершить'", name))
}

func HandleMyWorkouts(b *bot.Bot, message *tgbotapi.Message) {
	ctx := context.Background()

	user, err := b.DB.GetUserByTelegramID(ctx, message.From.ID)
	if err != nil {
		b.SendMessage(message.Chat.ID, "Ошибка при получении данных.")
		return
	}

	workouts, err := b.DB.GetWorkoutsByUser(ctx, user.ID, 10)
	if err != nil {
		log.Printf("Error getting workouts: %v", err)
		b.SendMessage(message.Chat.ID, "Ошибка при получении тренировок.")
		return
	}

	if len(workouts) == 0 {
		b.SendMessage(message.Chat.ID, "У вас пока нет тренировок. Добавьте первую!")
		return
	}

	var response strings.Builder
	response.WriteString("📝 Ваши последние тренировки:\n\n")

	for _, w := range workouts {
		exercises, _ := b.DB.GetExercisesByWorkout(ctx, w.ID)
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
