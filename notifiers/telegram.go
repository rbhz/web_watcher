package notifiers

import (
	"log"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/rbhz/web_watcher/watcher"
)

// TelegramNotifier Sends notifications via TelegramBotAPI
type TelegramNotifier struct {
	bot           *tgbotapi.BotAPI
	users         []int64
	messagePeriod time.Duration
	updates       []watcher.URLUpdate
	mux           sync.Mutex
}

// Notify users about update
func (n *TelegramNotifier) Notify(update watcher.URLUpdate) {
	if checkStatusChange(update) {
		n.mux.Lock()
		n.updates = append(n.updates, update)
		n.mux.Unlock()
	}
}

// Run sends message each n seconds
func (n *TelegramNotifier) Run() {
	for range time.Tick(n.messagePeriod * time.Second) {
		n.mux.Lock()
		if len(n.updates) == 0 {
			n.mux.Unlock()
		} else {
			message := getMessage(n.updates)
			n.updates = make([]watcher.URLUpdate, 0)
			n.mux.Unlock()
			wg := sync.WaitGroup{}
			for _, user := range n.users {
				wg.Add(1)
				go func(user int64) {
					defer wg.Done()
					n.sendMessage(user, message)
				}(user)
			}
			wg.Wait()
		}
	}
}

// SendMessage to telegram user
func (n *TelegramNotifier) sendMessage(user int64, message string) {
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
		bot:           bot,
		users:         cfg.Users,
		messagePeriod: cfg.MessagePeriod,
	}
}
