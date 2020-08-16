package notification_manager

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type TelegramNotificationManager struct {
	bot *tgbotapi.BotAPI
	// channels represent all telegram channels that should be notified
	channels []string
}

func NewTelegram(token string, channels []string) (NotificationManager, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &TelegramNotificationManager{
		bot:      bot,
		channels: channels,
	}, nil
}

func (nm *TelegramNotificationManager) Stop() {
	nm.bot.Client.CloseIdleConnections()
}

func (nm *TelegramNotificationManager) Subscribe(notificationChan <-chan string) {
	go func() {
		for url := range notificationChan {
			nm.send(url)
		}
	}()
}

func (nm *TelegramNotificationManager) send(url string) {
	for _, channel := range nm.channels {
		msg := tgbotapi.NewMessageToChannel(channel, url)
		_, err := nm.bot.Send(msg)
		if err != nil {
			log.Printf("could not send message for url %s to channel %s, error %s\n", url, channel, err)
		}
	}
}
