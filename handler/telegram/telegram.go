package telegram

import (
	"context"
	"url-shortener/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotHandler struct {
	bot     *tgbotapi.BotAPI
	service *service.LinkService
}

func NewBotHandler(token string) (*BotHandler, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return &BotHandler{bot: bot, service: nil}, err
	}
	return &BotHandler{bot: bot, service: nil}, nil
}

func (bot *BotHandler) StartBot(s *service.LinkService) {
	bot.service = s

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
		link, err := bot.service.CreateLink(ctx, msg.Text)
		if err != nil {
			reply := tgbotapi.NewMessage(msg.Chat.ID, msg.Text)
			bot.bot.Send(reply)
		} else {
			reply := tgbotapi.NewMessage(msg.Chat.ID, "localhost:8080/"+link.Short)
			bot.bot.Send(reply)
		}
	}
}
