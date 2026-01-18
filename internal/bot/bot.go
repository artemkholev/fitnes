package bot

import (
	"fitness-bot/internal/database"
	"fitness-bot/internal/models"
	"log"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	API    *tgbotapi.BotAPI
	DB     *database.DB
	States map[int64]*models.UserState
	mu     sync.RWMutex
}

func NewBot(token string, db *database.DB) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	log.Printf("Authorized on account %s", api.Self.UserName)

	return &Bot{
		API:    api,
		DB:     db,
		States: make(map[int64]*models.UserState),
	}, nil
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
	b.API.Send(msg)
}
