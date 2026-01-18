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

// IsAdmin проверяет является ли пользователь админом
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
