package notification_manager

import (
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type TelegramNotificationManager struct {
	bot *tgbotapi.BotAPI
	// channels represent all telegram channels that should be notified
	channels []string
	start    time.Time
	stop     time.Time
}

func NewTelegram(token string, channels []string, start, stop time.Time) (NotificationManager, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &TelegramNotificationManager{
		bot:      bot,
		channels: channels,
		start:    start,
		stop:     stop,
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
	// if current time is not in provided range we don't send a notification
	if !nm.checkTime() {
		return
	}

	for _, channel := range nm.channels {
		msg := tgbotapi.NewMessageToChannel(channel, url)
		_, err := nm.bot.Send(msg)
		if err != nil {
			log.Printf("could not send message for url %s to channel %s, error %s\n", url, channel, err)
		}
	}
}

func (nm *TelegramNotificationManager) checkTime() bool {
	t := time.Now()
	timeNow, err := time.Parse(time.Kitchen, t.Format(time.Kitchen))
	// if we have an error better to send a notification, than to skip it
	if err != nil {
		return true
	}

	if timeNow.Before(nm.start) {
		return false
	}

	if timeNow.Before(nm.stop) {
		return true
	}

	return false
}
