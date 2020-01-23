package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jinzhu/configor"
	"github.com/rbhz/http_checker/notifiers"
	"github.com/rbhz/http_checker/watcher"
	"github.com/rbhz/http_checker/web"

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
	configor.Load(conf, args.confPath)
	watcherInstance := watcher.NewWatcher(
		readFile(args.filePath),
		conf.Period,
		conf.DBPath,
	)

	var ns []watcher.Notifier
	if conf.Web.Active {
		webServer := web.GetServer(watcherInstance, conf.Web.Port)
		go webServer.Run()
		ns = append(ns, notifiers.WebNotifier{
			Server: webServer,
		})
	}
	if conf.PostMark.Active {
		ns = append(ns, notifiers.NewPostMarkNotifier(
			conf.PostMark.Emails,
			conf.PostMark.APIKey,
			conf.PostMark.FromEmail,
			conf.PostMark.Subject,
			conf.PostMark.MessageText,
		))
	}
	watcherInstance.Start(ns)
}
