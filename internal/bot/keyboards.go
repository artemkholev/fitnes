package bot

import (
	"fitness-bot/internal/models"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func GetStartMenuKeyboard(accessInfo *models.AccessInfo) tgbotapi.ReplyKeyboardMarkup {
	var rows [][]tgbotapi.KeyboardButton

	if accessInfo.IsAdmin {
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("👑 Админ-панель"),
		))
	}

	hasActiveManager := false
	for _, org := range accessInfo.ManagerOrgs {
		if org.IsActive {
			hasActiveManager = true
			break
		}
	}
	if hasActiveManager {
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🏢 Управление организацией"),
		))
	}

	hasActiveTrainer := false
	for _, org := range accessInfo.TrainerOrgs {
		if org.IsActive {
			hasActiveTrainer = true
			break
		}
	}
	if hasActiveTrainer {
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🏋️ Панель тренера"),
		))
	}

	if len(accessInfo.ClientAccess) > 0 {
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📝 Мои тренировки"),
		))
	}

	if len(accessInfo.ArchivedAccess) > 0 {
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📚 Архив тренировок"),
		))
	}

	if len(rows) == 0 {
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("ℹ️ О боте"),
		))
	}

	return tgbotapi.NewReplyKeyboard(rows...)
}

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

func GetTrainerMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("➕ Добавить клиента"),
			tgbotapi.NewKeyboardButton("👥 Мои клиенты"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📅 Групповые тренировки"),
			tgbotapi.NewKeyboardButton("📅 Создать групповую"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📊 Статистика"),
			tgbotapi.NewKeyboardButton("🔙 Главное меню"),
		),
	)
}

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

// GetMainMenuKeyboard устарела, оставлена для совместимости
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

// GetExerciseControlKeyboard показывается во время добавления упражнений.
// Позволяет завершить тренировку или отменить её целиком.
func GetExerciseControlKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("✅ Завершить"),
			tgbotapi.NewKeyboardButton("❌ Отмена"),
		),
	)
}

// ====== INLINE KEYBOARDS ======

// GetInlineOrganizationsKeyboard строит список организаций для выбора.
// Каждая кнопка передаёт ID организации как int через formatCallbackData.
func GetInlineOrganizationsKeyboard(orgs []*models.Organization, prefix string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, org := range orgs {
		btn := tgbotapi.NewInlineKeyboardButtonData(
			org.Name,
			formatCallbackData(prefix, org.ID),
		)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", prefix+":cancel"),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func GetInlineListKeyboard(items []string, ids []int64, prefix string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for i, item := range items {
		var id int64
		if i < len(ids) {
			id = ids[i]
		} else {
			id = int64(i + 1)
		}
		btn := tgbotapi.NewInlineKeyboardButtonData(
			item,
			formatCallbackData(prefix, id),
		)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", prefix+":cancel"),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func GetInlineMuscleGroupKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💪 Грудь", "muscle:chest"),
			tgbotapi.NewInlineKeyboardButtonData("🦾 Спина", "muscle:back"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🦵 Ноги", "muscle:legs"),
			tgbotapi.NewInlineKeyboardButtonData("🏋️ Плечи", "muscle:shoulders"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💪 Бицепс", "muscle:biceps"),
			tgbotapi.NewInlineKeyboardButtonData("💪 Трицепс", "muscle:triceps"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎯 Пресс", "muscle:abs"),
			tgbotapi.NewInlineKeyboardButtonData("🏃 Кардио", "muscle:cardio"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "muscle:cancel"),
		),
	)
}

func GetInlineClientActionsKeyboard(clientID int64, isActive bool) tgbotapi.InlineKeyboardMarkup {
	rows := [][]tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 Статистика", formatCallbackData("client_action", clientID)+":stats"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ Создать тренировку", formatCallbackData("client_action", clientID)+":workout"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 История тренировок", formatCallbackData("client_action", clientID)+":history"),
		),
	}
	if isActive {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Удалить клиента", formatCallbackData("client_action", clientID)+":delete"),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", "client_action:back"),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func GetInlineConfirmKeyboard(prefix string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Да", prefix+":confirm"),
			tgbotapi.NewInlineKeyboardButtonData("❌ Нет", prefix+":cancel"),
		),
	)
}

// GetInlineFinishKeyboard показывается после добавления упражнения.
func GetInlineFinishKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ Ещё упражнение", "exercise:more"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Завершить тренировку", "exercise:finish"),
		),
	)
}

// GetInlineSetsKeyboard для выбора количества подходов.
// Числовые кнопки передают значение через id; "other" — переход в текстовый ввод.
func GetInlineSetsKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("1", "ex_sets:1"),
			tgbotapi.NewInlineKeyboardButtonData("2", "ex_sets:2"),
			tgbotapi.NewInlineKeyboardButtonData("3", "ex_sets:3"),
			tgbotapi.NewInlineKeyboardButtonData("4", "ex_sets:4"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("5", "ex_sets:5"),
			tgbotapi.NewInlineKeyboardButtonData("6", "ex_sets:6"),
			tgbotapi.NewInlineKeyboardButtonData("✏️ Своё", "ex_sets:other"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "ex_sets:cancel"),
		),
	)
}

// GetInlineRepsKeyboard для выбора количества повторений.
func GetInlineRepsKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("6", "ex_reps:6"),
			tgbotapi.NewInlineKeyboardButtonData("8", "ex_reps:8"),
			tgbotapi.NewInlineKeyboardButtonData("10", "ex_reps:10"),
			tgbotapi.NewInlineKeyboardButtonData("12", "ex_reps:12"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("15", "ex_reps:15"),
			tgbotapi.NewInlineKeyboardButtonData("20", "ex_reps:20"),
			tgbotapi.NewInlineKeyboardButtonData("25", "ex_reps:25"),
			tgbotapi.NewInlineKeyboardButtonData("✏️ Своё", "ex_reps:other"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "ex_reps:cancel"),
		),
	)
}

// GetInlineWeightKeyboard для выбора веса в кг.
func GetInlineWeightKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("10", "ex_weight:10"),
			tgbotapi.NewInlineKeyboardButtonData("15", "ex_weight:15"),
			tgbotapi.NewInlineKeyboardButtonData("20", "ex_weight:20"),
			tgbotapi.NewInlineKeyboardButtonData("25", "ex_weight:25"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("30", "ex_weight:30"),
			tgbotapi.NewInlineKeyboardButtonData("40", "ex_weight:40"),
			tgbotapi.NewInlineKeyboardButtonData("50", "ex_weight:50"),
			tgbotapi.NewInlineKeyboardButtonData("60", "ex_weight:60"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("70", "ex_weight:70"),
			tgbotapi.NewInlineKeyboardButtonData("80", "ex_weight:80"),
			tgbotapi.NewInlineKeyboardButtonData("100", "ex_weight:100"),
			tgbotapi.NewInlineKeyboardButtonData("✏️ Своё", "ex_weight:other"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "ex_weight:cancel"),
		),
	)
}

// formatCallbackData форматирует callback data с целочисленным ID.
func formatCallbackData(prefix string, id int64) string {
	return prefix + ":" + strconv.FormatInt(id, 10)
}

// ParseCallbackData разбирает callback data формата "prefix:id[:action]".
// Если второй сегмент не является числом, он становится action, а id=0.
func ParseCallbackData(data string) (prefix string, id int64, action string) {
	parts := strings.Split(data, ":")
	if len(parts) < 2 {
		return data, 0, ""
	}
	prefix = parts[0]
	if parsed, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
		id = parsed
		if len(parts) >= 3 {
			action = parts[2]
		}
	} else {
		action = parts[1]
	}
	return
}
