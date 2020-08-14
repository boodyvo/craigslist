package main

import (
	"log"

	"github.com/caarlos0/env"

	_ "github.com/go-sql-driver/mysql"

	notification_manager "github.com/boodyvo/craigslist/notification-manager"
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

	tgbot, err := notification_manager.NewTelegram(config.TGToken, config.TGChannels)
	if err != nil {
		log.Fatalf("cannot create telegram bot: %s", err)
	}
	defer tgbot.Stop()

	craigslistScrapper := scrapper.New()
	log.Println(craigslistScrapper.GetLastIndex())

	notificationChanDB := craigslistScrapper.SubscriptionChan()
	db.Subscribe(notificationChanDB)

	notificationChanTelegram := craigslistScrapper.SubscriptionChan()
	tgbot.Subscribe(notificationChanTelegram)

	log.Printf("finish scrapper with error: %s", craigslistScrapper.Start())
}
