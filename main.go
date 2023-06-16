package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	opt_config_file = flag.String("conf", DefautlConfigFile, "configuration file")
	opt_log_level   = flag.String("log", "", "log level")
	opt_log_file    = flag.String("log-output", "", "log output file, val: stdout, <file nam>")
)

func load_config() {
	flag.CommandLine.Init(AppName, flag.ContinueOnError)

	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			fmt.Println()
			os.Exit(0)
		}

		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if err := LoadConfig(*opt_config_file); err != nil {
		fmt.Fprintf(os.Stderr, "load configuration file %s error: %v\n", *opt_config_file, err)
		os.Exit(1)
	}

	if opt_log_level != nil && *opt_log_level != "" {
		GetConfig().LogLevel = *opt_log_level
	}

	if opt_log_file != nil && *opt_log_file != "" {
		GetConfig().LogFile = *opt_log_file
	}

	if GetConfig().LogFile == "stdout" {
		log.SetOutput(os.Stdout)
	} else {
		writer, err := os.OpenFile(GetConfig().LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			fmt.Printf("open log file fail: %s", err.Error())
			os.Exit(1)
		}
		log.SetOutput(writer)
	}

	logLevel, err := log.ParseLevel(GetConfig().LogLevel)
	if err != nil {
		fmt.Printf("invalid log level: %s", err.Error())
		os.Exit(1)
	}

	log.SetLevel(logLevel)
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true, DisableColors: false, TimestampFormat: time.DateTime})
}

func runSignalLoop() {
	c := make(chan os.Signal, 3)

	signal.Notify(c, syscall.SIGINT)
	defer close(c)

	exit := false
	for !exit {
		s := <-c
		if s == syscall.SIGINT {
			exit = true
		}
	}
}

func main() {
	load_config()

	if len(GetConfig().Mirrors) == 0 {
		log.Errorf("no mirror configuration")
		os.Exit(2)
	}

	log.Infof("start ...")
	log.Infof("config: \n%s", GetConfigJson())

	for _, mirrorConfig := range GetConfig().Mirrors {
		serv, err := NewServer(mirrorConfig)
		if err != nil {
			log.Errorf("start mirror server local %s target %s fail: %s", mirrorConfig.Local, mirrorConfig.Target, err.Error())
			os.Exit(2)
		}

		log.Infof("start mirror server local %s target %s success", mirrorConfig.Local, mirrorConfig.Target)
		serv.Start()
	}

	runSignalLoop()
}
