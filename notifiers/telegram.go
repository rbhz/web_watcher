package notifiers

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/rbhz/web_watcher/watcher"
)

// TelegramNotifier Sends notifications via TelegramBotAPI
type TelegramNotifier struct {
	bot         *tgbotapi.BotAPI
	users       []int64
	messageText string
}

// Notify users about update
func (n TelegramNotifier) Notify(updates []watcher.URLUpdate) {
	updates = filterFails(updates)
	if len(updates) > 0 {
		for _, user := range n.users {
			go n.SendMessage(
				user, getMessage(updates, n.messageText),
			)
		}
	}
}

// SendMessage to telegram user
func (n TelegramNotifier) SendMessage(user int64, message string) {
	msg := tgbotapi.NewMessage(user, message)
	_, err := n.bot.Send(msg)
	if err != nil {
		log.Printf("Failed to send telegram message: %v", err)
	}
}

// NewTelegramNotifier creates notifier
func NewTelegramNotifier(cfg TelegramConfig) TelegramNotifier {
	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		log.Fatalf("Failed to initialize Telegram bot: %v", err)
	}
	return TelegramNotifier{
		bot:         bot,
		users:       cfg.Users,
		messageText: cfg.MessageText,
	}
}
