package watcher

import (
	"database/sql"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// Notifier interface
type Notifier interface {
	Notify(URLUpdate)
}

// Watcher check if urls changed
type Watcher struct {
	urls        []*URL
	period      time.Duration
	errorPeriod time.Duration
	dbPath      string
	db          *sql.DB
}

// Start watcher as daemon
func (w *Watcher) Start(notifiers []Notifier) {
	checking := make(map[int]bool)
	updates := make(chan URLUpdate)

	for {
		select {
		case <-time.Tick(100 * time.Millisecond):
			for _, url := range w.urls {
				var period time.Duration
				if url.Good() {
					period = w.period
				} else {
					period = w.errorPeriod
				}
				if url.lastCheck.Add(period).Before(time.Now()) {
					if _, ok := checking[url.id]; ok {
						continue
					}
					checking[url.id] = true
					log.Debug().Str("url", url.Link).Msg("Found url to check")
					go w.check(url, updates)
				}
			}
		case update := <-updates:
			log.Debug().Str("url", update.New.Link).Msg("Checked")
			delete(checking, update.Old.id)
			if len(update.Changed) > 0 {
				for _, n := range notifiers {
					go n.Notify(update)
				}
			}
		}
	}
}

func (w *Watcher) check(url *URL, out chan<- URLUpdate) {
	log.Debug().Str("url", url.Link).Msg("Got url to check")
	update := url.Update()
	update.New.save(w.db)
	out <- update
}

func (w *Watcher) initDB() {
	db, err := sql.Open("sqlite3", w.dbPath)
	log.Info().Msg("Initializing Database")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to open db")
	}
	_, err = db.Exec(
		`CREATE TABLE IF NOT EXISTS urls (
			link VARCHAR(200) PRIMARY KEY,
			last_change DATE NOT NULL,
			hash BLOB,
			status INT NOT NULL,
			error VARCHAR(500) NOT NULL
		);`,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create table")
	}
	w.db = db
}

// GetUrls return watchers urls slice
func (w Watcher) GetUrls() []*URL {
	return w.urls
}

// NewWatcher returns watcher
func NewWatcher(urls []string, cfg Config) Watcher {
	watcher := Watcher{
		period:      cfg.Period * time.Second,
		errorPeriod: cfg.ErrorPeriod * time.Second,
		dbPath:      cfg.DBPath}
	watcher.initDB()
	var wg sync.WaitGroup
	wg.Add(len(urls))
	for idx, url := range urls {
		go func(idx int, url string) {
			defer wg.Done()
			watcher.urls = append(watcher.urls, getURL(idx, url, watcher.db))
		}(idx, url)
	}
	wg.Wait()
	return watcher
}
