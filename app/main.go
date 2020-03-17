package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jinzhu/configor"
	"github.com/rbhz/web_watcher/notifiers"
	"github.com/rbhz/web_watcher/watcher"

	_ "github.com/mattn/go-sqlite3"
)

type arguments struct {
	filePath string
	confPath string
}

type runnable interface {
	Run()
}

func getArguments() (args arguments) {
	flag.Usage = func() {
		fmt.Printf("Usage: %s [OPTIONS] path_to_file \n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.StringVar(&args.confPath, "conf", "./config.yaml", "Path to config")
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}
	args.filePath = flag.Arg(0)
	return
}

func readFile(path string) (lines []string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return
}

func main() {
	args := getArguments()
	conf := &Config{}
	err := configor.Load(conf, args.confPath)
	if err != nil {
		log.Fatal(err)
	}
	watcherInstance := watcher.NewWatcher(
		readFile(args.filePath),
		conf.App)

	var ns []watcher.Notifier
	if conf.Web.Active {
		notifier := notifiers.NewWebNotifier(conf.Web, watcherInstance)
		ns = append(ns, &notifier)

	}
	if conf.PostMark.Active {
		notifier := notifiers.NewPostMarkNotifier(conf.PostMark)
		ns = append(ns, &notifier)
	}
	if conf.Telegram.Active {
		notifier := notifiers.NewTelegramNotifier(conf.Telegram)
		ns = append(ns, &notifier)
	}
	for _, notifier := range ns {
		if notifier, ok := notifier.(runnable); ok {
			go notifier.Run()
		}
	}
	watcherInstance.Start(ns)
}
