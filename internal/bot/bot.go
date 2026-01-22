package bot

import (
	"fitness-bot/internal/database"
	"fitness-bot/internal/models"
	"log"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	API           *tgbotapi.BotAPI
	DB            *database.DB
	AdminUsername string
	States        map[int64]*models.UserState
	mu            sync.RWMutex
}

func NewBot(token string, db *database.DB, adminUsername string) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	log.Printf("Authorized on account %s", api.Self.UserName)
	log.Printf("Admin username: @%s", adminUsername)

	return &Bot{
		API:           api,
		DB:            db,
		AdminUsername: database.NormalizeUsername(adminUsername),
		States:        make(map[int64]*models.UserState),
	}, nil
}

func (b *Bot) IsAdmin(username string) bool {
	normalized := strings.TrimPrefix(strings.TrimSpace(username), "@")
	return strings.EqualFold(normalized, b.AdminUsername)
}

func (b *Bot) GetState(telegramID int64) *models.UserState {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.States[telegramID]
}

func (b *Bot) SetState(telegramID int64, state string, data map[string]interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.States[telegramID] = &models.UserState{
		TelegramID: telegramID,
		State:      state,
		Data:       data,
	}
}

func (b *Bot) ClearState(telegramID int64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.States, telegramID)
}

func (b *Bot) SendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	b.API.Send(msg)
}

func (b *Bot) SendMessageWithKeyboard(chatID int64, text string, keyboard interface{}) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	msg.ParseMode = "Markdown"
	b.API.Send(msg)
}

func (b *Bot) SendMessageMarkdown(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	b.API.Send(msg)
}

func (b *Bot) SendWithCancel(chatID int64, text string) {
	b.SendMessageWithKeyboard(chatID, text, GetCancelKeyboard())
}

func (b *Bot) SendSuccess(chatID int64, text string, keyboard interface{}) {
	b.SendMessageWithKeyboard(chatID, "✅ "+text, keyboard)
}

func (b *Bot) SendError(chatID int64, text string) {
	b.SendMessage(chatID, "❌ "+text)
}

// SendInlineKeyboard отправляет сообщение с inline-клавиатурой и сохраняет его ID для очистки
func (b *Bot) SendInlineKeyboard(chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) int {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	msg.ParseMode = "Markdown"
	sent, err := b.API.Send(msg)
	if err != nil {
		return 0
	}
	// Сохраняем ID сообщения для последующей очистки (в личных чатах chatID == telegramID)
	b.StoreMessageID(chatID, sent.MessageID)
	return sent.MessageID
}

// EditMessageText редактирует текст сообщения
func (b *Bot) EditMessageText(chatID int64, messageID int, text string, keyboard *tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	msg.ParseMode = "Markdown"
	if keyboard != nil {
		msg.ReplyMarkup = keyboard
	}
	b.API.Send(msg)
}

// DeleteMessage удаляет сообщение
func (b *Bot) DeleteMessage(chatID int64, messageID int) {
	del := tgbotapi.NewDeleteMessage(chatID, messageID)
	b.API.Send(del)
}

// AnswerCallback отвечает на callback query
func (b *Bot) AnswerCallback(callbackID string, text string) {
	callback := tgbotapi.NewCallback(callbackID, text)
	b.API.Send(callback)
}

// StoreMessageID сохраняет ID сообщения для последующего удаления
func (b *Bot) StoreMessageID(telegramID int64, messageID int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.States[telegramID] != nil {
		if b.States[telegramID].Data == nil {
			b.States[telegramID].Data = make(map[string]interface{})
		}
		// Сохраняем список сообщений для удаления
		var msgIDs []int
		if existing, ok := b.States[telegramID].Data["_message_ids"].([]int); ok {
			msgIDs = existing
		}
		msgIDs = append(msgIDs, messageID)
		b.States[telegramID].Data["_message_ids"] = msgIDs
	}
}

// CleanupMessages удаляет все сохранённые сообщения
func (b *Bot) CleanupMessages(chatID int64, telegramID int64) {
	b.mu.RLock()
	var msgIDs []int
	if state := b.States[telegramID]; state != nil && state.Data != nil {
		if ids, ok := state.Data["_message_ids"].([]int); ok {
			msgIDs = ids
		}
	}
	b.mu.RUnlock()

	for _, msgID := range msgIDs {
		b.DeleteMessage(chatID, msgID)
	}

	// Очищаем список
	b.mu.Lock()
	if state := b.States[telegramID]; state != nil && state.Data != nil {
		delete(state.Data, "_message_ids")
	}
	b.mu.Unlock()
}
