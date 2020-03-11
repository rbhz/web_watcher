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
		webNotifier := notifiers.NewWebNotifier(conf.Web, watcherInstance)
		go webNotifier.Server.Run()
		ns = append(ns, webNotifier)

	}
	// if conf.PostMark.Active {
	// 	ns = append(ns, notifiers.NewPostMarkNotifier(conf.PostMark))
	// }
	// if conf.Telegram.Active {
	// 	ns = append(ns, notifiers.NewTelegramNotifier(conf.Telegram))
	// }
	watcherInstance.Start(ns)
}
