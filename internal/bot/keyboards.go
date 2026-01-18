package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func GetMainMenuKeyboard(isTrainer bool) tgbotapi.ReplyKeyboardMarkup {
	var rows [][]tgbotapi.KeyboardButton

	if isTrainer {
		rows = [][]tgbotapi.KeyboardButton{
			{
				tgbotapi.NewKeyboardButton("➕ Создать тренировку"),
				tgbotapi.NewKeyboardButton("👥 Мои клиенты"),
			},
			{
				tgbotapi.NewKeyboardButton("📅 Групповые тренировки"),
				tgbotapi.NewKeyboardButton("📊 Статистика"),
			},
		}
	} else {
		rows = [][]tgbotapi.KeyboardButton{
			{
				tgbotapi.NewKeyboardButton("➕ Добавить тренировку"),
				tgbotapi.NewKeyboardButton("📝 Мои тренировки"),
			},
			{
				tgbotapi.NewKeyboardButton("🔍 Найти тренера"),
				tgbotapi.NewKeyboardButton("📅 Групповые тренировки"),
			},
			{
				tgbotapi.NewKeyboardButton("📊 Моя статистика"),
			},
		}
	}

	return tgbotapi.NewReplyKeyboard(rows...)
}

func GetMuscleGroupKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("💪 Грудь"),
			tgbotapi.NewKeyboardButton("🦾 Спина"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🦵 Ноги"),
			tgbotapi.NewKeyboardButton("🏋️ Плечи"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("💪 Бицепс"),
			tgbotapi.NewKeyboardButton("💪 Трицепс"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🎯 Пресс"),
			tgbotapi.NewKeyboardButton("🏃 Кардио"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("❌ Отмена"),
		),
	)
}

func GetRoleKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("👤 Клиент"),
			tgbotapi.NewKeyboardButton("💼 Тренер"),
		),
	)
}

func GetCancelKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("❌ Отмена"),
		),
	)
}
