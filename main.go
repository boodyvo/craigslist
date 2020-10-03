package main

import (
	"log"
	"time"

	notification_manager "github.com/boodyvo/craigslist/notification-manager"

	"github.com/caarlos0/env"

	_ "github.com/go-sql-driver/mysql"

	"github.com/boodyvo/craigslist/scrapper"
)

type Config struct {
	User       string   `env:"CRAIGSLIST_DB_USER,required"`
	Password   string   `env:"CRAIGSLIST_DB_PASSWORD,required"`
	Host       string   `env:"CRAIGSLIST_DB_HOST,required"`
	Database   string   `env:"CRAIGSLIST_DB_DATABASE,required"`
	Table      string   `env:"CRAIGSLIST_DB_TABLE,required"`
	FieldName  string   `env:"CRAIGSLIST_DB_FIELD_NAME,required"`
	TGToken    string   `env:"CRAIGSLIST_TELEGRAM_BOT_TOKEN,required"`
	TGChannels []string `env:"CRAIGSLIST_TELEGRAM_CHANNEL,required"`
	StartTime  string   `env:"CRAIGSLIST_START_TIME,required"`
	StopTime   string   `env:"CRAIGSLIST_STOP_TIME,required"`
	ProxyURL   string   `env:"CRAIGSLIST_PROXY_URL,required"`
}

func stringToTime(str string) (time.Time, error) {
	tm, err := time.Parse(time.Kitchen, str)
	if err != nil {
		return time.Time{}, err
	}

	return tm, nil
}

func main() {
	config := Config{}
	err := env.Parse(&config)
	if err != nil {
		log.Fatalf("cannot parse config: %s", err)
	}
	db, err := notification_manager.NewMySQL(
		config.User,
		config.Password,
		config.Host,
		config.Database,
		config.Table,
		config.FieldName,
	)
	if err != nil {
		log.Fatalf("cannot connect to db: %s", err)
	}
	defer db.Stop()

	start, err := stringToTime(config.StartTime)
	if err != nil {
		log.Fatalf("canont parse start time: %s", err)
	}
	stop, err := stringToTime(config.StopTime)
	if err != nil {
		log.Fatalf("canont parse stop time: %s", err)
	}
	if stop.Before(start) {
		stop = stop.Add(24 * time.Hour)
	}
	tgbot, err := notification_manager.NewTelegram(
		config.TGToken,
		config.TGChannels,
		start,
		stop,
	)
	if err != nil {
		log.Fatalf("cannot create telegram bot: %s", err)
	}
	defer tgbot.Stop()

	craigslistScrapper := scrapper.New(config.ProxyURL)
	log.Println(craigslistScrapper.GetLastIndex())

	notificationChanDB := craigslistScrapper.SubscriptionChan()
	db.Subscribe(notificationChanDB)

	notificationChanTelegram := craigslistScrapper.SubscriptionChan()
	tgbot.Subscribe(notificationChanTelegram)

	log.Printf("finish scrapper with error: %s", craigslistScrapper.Start())
}
