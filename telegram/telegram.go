package telegram

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/OloloevReal/go-simple-baby-bot/store"
	log "github.com/OloloevReal/go-simple-log"

	tgbotapiv5 "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/net/proxy"
)

const (
	commandStart = "/start"
	commandHelp  = "/help"
)

var commandDescription = map[string]string{
	commandStart: "Start bot",
	commandHelp:  "Available commands",
}

type TelegramConfig struct {
	Token   string
	Proxy   proxy.Dialer
	Timeout time.Duration
	Store   store.Store
	Debug   bool
}

type TelegramBot struct {
	config   *TelegramConfig
	botAPI   *tgbotapiv5.BotAPI
	handlers map[string]handler
}

type handler func(update *tgbotapiv5.Update)

func NewTelegram(config *TelegramConfig) (bot *TelegramBot, err error) {

	bot = &TelegramBot{
		config:   config,
		handlers: make(map[string]handler),
	}

	bot.makeHandlers()

	if config.Proxy != nil {
		tr := &http.Transport{}
		tr.Dial = config.Proxy.Dial
		client := &http.Client{Transport: tr}
		bot.botAPI, err = tgbotapiv5.NewBotAPIWithClient(config.Token, client)
	} else {
		bot.botAPI, err = tgbotapiv5.NewBotAPI(config.Token)
	}
	bot.botAPI.Debug = config.Debug

	return
}

func (b *TelegramBot) Run(ctx context.Context) {
	defer log.Println("[INFO] telegram bot finished")

	bot := b.botAPI
	bot.Request(tgbotapiv5.RemoveWebhookConfig{})
	updateConfig := tgbotapiv5.NewUpdate(0)
	updateConfig.Timeout = 60
	updates := bot.GetUpdatesChan(updateConfig)
	updates.Clear()

	time.Sleep(time.Millisecond * 500)
	for {
		select {
		case update := <-updates:
			{
				if (update.Message != nil && update.Message.IsCommand()) ||
					(update.CallbackQuery != nil && strings.HasPrefix(update.CallbackQuery.Data, "/")) {

					var userID int
					userID, err := b.getUserID(&update)
					if err != nil {
						log.Printf("[ERROR] can't get UserID, %s", err)
					}
					var cmd string
					if update.Message != nil {
						cmd = update.Message.Text
					} else {
						cmd = update.CallbackQuery.Data
					}
					log.Printf("[DEBUG] User ID=%d, received command=%s", userID, cmd)
					handler, err := b.GetHandler(cmd)
					if err != nil {
						log.Printf("[ERROR] can't find handler for command=\"%s\"", cmd)
						continue
					}
					_ = handler

				} else if update.Message != nil && len(update.Message.Text) > 0 {
					var userID int
					userID, err := b.getUserID(&update)
					if err != nil {
						log.Printf("[ERROR] can't get UserID, %s", err)
					}
					log.Printf("[DEBUG] received text message from user=%d: \"%s\"", userID, update.Message.Text)

					value := new(store.BValue)
					value.Timestamp = time.Now()
					value.UserID = userID

					if err := value.ParseValue(update.Message.Text); err != nil {
						log.Printf("[ERROR] %s", err)
					} else {

						ctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)
						defer cancel()
						lastValue, err := b.config.Store.GetLast(ctx, userID)
						if err != nil {
							log.Printf("[ERROR] failed to get last value, %s", err)
						} else {
							ctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)
							defer cancel()

							err = b.config.Store.Put(ctx, value)
							if err != nil {
								log.Printf("[ERROR] failed to put data into database")
							}
						}

						msgConfig := tgbotapiv5.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Прошлое: %d г.\r\nТекущее: %d г.\r\n---------\r\nРазница: %d г.", lastValue, value.Value, value.Value-lastValue))
						b.botAPI.Request(msgConfig)

					}

				}
			}
		case <-ctx.Done():
			{
				return
			}
		}
	}
}

func (b *TelegramBot) makeHandlers() {
	b.AddHandler(commandHelp, nil)
	b.AddHandler(commandStart, nil)
}

func (b *TelegramBot) AddHandler(cmd string, fn handler) {
	b.handlers[cmd] = fn
}

func (b *TelegramBot) GetHandler(cmd string) (fn handler, err error) {
	if b.handlers == nil {
		return fn, errors.New("handlers is nil")
	}
	fn, ok := b.handlers[cmd]
	if !ok {
		return fn, fmt.Errorf("%s didn't find", cmd)
	}
	return
}

func (b *TelegramBot) getChatID(update *tgbotapiv5.Update) (chatID int64, err error) {
	if update != nil {
		if update.Message != nil {
			chatID = update.Message.Chat.ID
		} else if update.CallbackQuery != nil {
			chatID = update.CallbackQuery.Message.Chat.ID
		} else {
			err = fmt.Errorf("can't to determine ChatID")
		}
		return
	}
	err = fmt.Errorf("update object is nil")
	return
}

func (b *TelegramBot) getUserID(update *tgbotapiv5.Update) (userID int, err error) {
	if update != nil {
		if update.Message != nil {
			userID = update.Message.From.ID
		} else if update.CallbackQuery != nil {
			userID = update.CallbackQuery.From.ID
		} else {
			err = fmt.Errorf("can't to determine UserID")
		}
		return
	}
	err = fmt.Errorf("update object is nil")
	return
}
