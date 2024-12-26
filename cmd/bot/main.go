package main

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/aidosgal/lenshub/internal/config"
)

func main() {
	// Load configuration
	cfg := config.MustLoad()

	// Initialize database connection
	db, err := sqlx.Connect("postgres", fmt.Sprintf(
		"user=%s password=%s host=%s dbname=%s port=%d sslmode=%s",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Name, cfg.Database.Port, cfg.Database.SSLMode,
	))
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()

	// Initialize Telegram bot
	bot, err := tgbotapi.NewBotAPI(cfg.Telegram)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil || !update.Message.IsCommand() {
			continue
		}

		switch update.Message.Command() {
		case "start":
			analytics, err := fetchAnalytics(db)
			if err != nil {
				log.Printf("Failed to fetch analytics: %v", err)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при получении аналитики.")
				bot.Send(msg)
				continue
			}

			// Prepare and send the analytics response
			msgText := fmt.Sprintf(
				"📊 Аналитика:\n\n"+
					"👥 Количество пользователей: %d\n"+
					"🔧 Исполнители: %d\n"+
					"👔 Заказчики: %d\n"+
					"📦 Заказы: %d\n"+
					"💬 Отклики: %d",
				analytics.TotalUsers, analytics.Executors, analytics.Customers, analytics.Orders, analytics.Responses,
			)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
			bot.Send(msg)
		default:
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда.")
			bot.Send(msg)
		}
	}
}

// Analytics holds the analytics data
type Analytics struct {
	TotalUsers int
	Executors  int
	Customers  int
	Orders     int
	Responses  int
}

// fetchAnalytics queries the database and returns analytics data
func fetchAnalytics(db *sqlx.DB) (*Analytics, error) {
	var analytics Analytics

	// Query total users
	err := db.Get(&analytics.TotalUsers, "SELECT COUNT(*) FROM users")
	if err != nil {
		return nil, err
	}

	// Query executors
	err = db.Get(&analytics.Executors, "SELECT COUNT(*) FROM users WHERE role = 'Исполнитель'")
	if err != nil {
		return nil, err
	}

	// Query customers
	err = db.Get(&analytics.Customers, "SELECT COUNT(*) FROM users WHERE role = 'Заказчик'")
	if err != nil {
		return nil, err
	}

	// Query orders
	err = db.Get(&analytics.Orders, "SELECT COUNT(*) FROM orders")
	if err != nil {
		return nil, err
	}

	// Query responses
	err = db.Get(&analytics.Responses, "SELECT COUNT(*) FROM responses")
	if err != nil {
		return nil, err
	}

	return &analytics, nil
}

