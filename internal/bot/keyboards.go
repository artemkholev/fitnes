package bot

import (
	"fitness-bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// GetStartMenuKeyboard возвращает начальное меню на основе доступов пользователя
func GetStartMenuKeyboard(accessInfo *models.AccessInfo) tgbotapi.ReplyKeyboardMarkup {
	var rows [][]tgbotapi.KeyboardButton

	// Админ
	if accessInfo.IsAdmin {
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("👑 Админ-панель"),
		))
	}

	// Менеджер
	if len(accessInfo.ManagerOrgs) > 0 {
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🏢 Управление организацией"),
		))
	}

	// Тренер
	if len(accessInfo.TrainerOrgs) > 0 {
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🏋️ Панель тренера"),
		))
	}

	// Клиент (активные доступы)
	if len(accessInfo.ClientAccess) > 0 {
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📝 Мои тренировки"),
		))
	}

	// Архивные доступы (только просмотр)
	if len(accessInfo.ArchivedAccess) > 0 {
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📚 Архив тренировок"),
		))
	}

	// Если нет никаких доступов
	if len(rows) == 0 {
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("ℹ️ О боте"),
		))
	}

	return tgbotapi.NewReplyKeyboard(rows...)
}

// GetAdminMenuKeyboard возвращает клавиатуру админ-панели
func GetAdminMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🏢 Создать организацию"),
			tgbotapi.NewKeyboardButton("📋 Список организаций"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔙 Главное меню"),
		),
	)
}

// GetOrgManageKeyboard возвращает клавиатуру управления организацией (для админа)
func GetOrgManageKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("➕ Добавить менеджера"),
			tgbotapi.NewKeyboardButton("📋 Список менеджеров"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔙 К списку организаций"),
		),
	)
}

// GetManagerMenuKeyboard возвращает клавиатуру менеджера
func GetManagerMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("➕ Добавить тренера"),
			tgbotapi.NewKeyboardButton("📋 Список тренеров"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔙 Главное меню"),
		),
	)
}

// GetTrainerMenuKeyboard возвращает клавиатуру тренера
func GetTrainerMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("➕ Добавить клиента"),
			tgbotapi.NewKeyboardButton("👥 Мои клиенты"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📅 Групповые тренировки"),
			tgbotapi.NewKeyboardButton("📊 Статистика"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔙 Главное меню"),
		),
	)
}

// GetClientMenuKeyboard возвращает клавиатуру клиента
func GetClientMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("➕ Добавить тренировку"),
			tgbotapi.NewKeyboardButton("📝 Мои тренировки"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📊 Моя статистика"),
			tgbotapi.NewKeyboardButton("📅 Групповые тренировки"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔙 Главное меню"),
		),
	)
}

// GetMainMenuKeyboard - устаревшая функция для совместимости
func GetMainMenuKeyboard(isTrainer bool) tgbotapi.ReplyKeyboardMarkup {
	if isTrainer {
		return GetTrainerMenuKeyboard()
	}
	return GetClientMenuKeyboard()
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
