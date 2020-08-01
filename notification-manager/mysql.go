package notification_manager

import (
	"database/sql"
	"fmt"
	"log"
)

type MySQLNotificationManager struct {
	table      string
	fieldName  string
	connection *sql.DB
}

func NewMySQL(user, password, host, database, table, fieldName string) (NotificationManager, error) {
	db, err := sql.Open(
		"mysql",
		fmt.Sprintf(
			"%s:%s@tcp(%s)/%s",
			user,
			password,
			host,
			database,
		),
	)
	if err != nil {
		return nil, err
	}

	return &MySQLNotificationManager{
		table:      table,
		fieldName:  fieldName,
		connection: db,
	}, nil
}

func (nm *MySQLNotificationManager) Stop() {
	_ = nm.connection.Close()
}

func (nm *MySQLNotificationManager) Subscribe(notificationChan <-chan string) {
	go func() {
		for url := range notificationChan {
			nm.insert(url)
		}
	}()
}

func (nm *MySQLNotificationManager) insert(url string) {
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES ( '%s' )",
		nm.table,
		nm.fieldName,
		url,
	)
	insert, err := nm.connection.Query(query)
	if err != nil {
		log.Printf("cannot insert url into db: %s", err)

		return
	}
	defer insert.Close()
}
