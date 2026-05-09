package telegram

import (
	"context"
	"net/url"
	"strings"
	"url-shortener/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/net/publicsuffix"
)

type BotHandler struct {
	bot     *tgbotapi.BotAPI
	service *service.LinkService
	prefix  string
}

func NewBotHandler(token string) (*BotHandler, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return &BotHandler{bot: bot, service: nil}, err
	}
	return &BotHandler{bot: bot, service: nil}, nil
}

func (bot *BotHandler) StartBot(s *service.LinkService, prefix string) {
	bot.service = s
	bot.prefix = prefix

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			go bot.HandleMessage(update.Message)
		}
	}
}

func (bot *BotHandler) HandleMessage(msg *tgbotapi.Message) {
	switch msg.Command() {
	case "start":
		reply := tgbotapi.NewMessage(msg.Chat.ID, "asdfhsdaifhakdf")
		bot.bot.Send(reply)
	default:
		ctx := context.Background()
		text := msg.Text
		if !strings.HasPrefix(text, "https://") || strings.HasPrefix(text, "http://") {
			text = "https://" + text
		}

		if _, err := url.Parse(text); err != nil {
			reply := tgbotapi.NewMessage(msg.Chat.ID, err.Error())
			bot.bot.Send(reply)
			return
		}
		if _, err := publicsuffix.EffectiveTLDPlusOne(text); err != nil {
			reply := tgbotapi.NewMessage(msg.Chat.ID, err.Error())
			bot.bot.Send(reply)
			return
		}

		link, err := bot.service.CreateLink(ctx, text)
		if err != nil {
			reply := tgbotapi.NewMessage(msg.Chat.ID, err.Error())
			bot.bot.Send(reply)
		} else {
			reply := tgbotapi.NewMessage(msg.Chat.ID, bot.prefix+link.Short)
			bot.bot.Send(reply)
		}
	}
}
