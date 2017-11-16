package main

import (
	"github.com/Dimonchik0036/vk-api"
	"github.com/sirupsen/logrus"
	r "gopkg.in/gorethink/gorethink.v3"
	"os"
	"strings"
)

var log = logrus.New()

var (
	client     *vkapi.Client
	session    *r.Session
	dbPKPrefix = "vk:"
)

func main() {
	log.Formatter = new(logrus.TextFormatter)
	log.Info("OverStatsVK 1.0 started!")

	var err error

	token := os.Getenv("TOKEN")
	if token == "" {
		log.Fatal("TOKEN env variable not specified!")
	}

	client, err = vkapi.NewClientFromToken(token)
	if err != nil {
		log.Fatal(err)
	}

	// Database pool init
	go InitConnectionPool()

	//log.Infof("authorized on account @%s", bot.Self.UserName)
	if err := client.InitLongPoll(0, 2); err != nil {
		log.Fatal(err)
	}

	updates, _, err := client.GetLPUpdatesChan(100, vkapi.LPConfig{25, vkapi.LPModeAttachments})
	if err != nil {
		log.Fatal(err)
	}

	for update := range updates {
		if update.Message == nil || !update.IsNewMessage() || update.Message.Outbox() {
			continue
		}

		command := strings.ToLower(update.Message.Text)

		// userId for logger
		commandLogger := log.WithFields(logrus.Fields{"user_id": update.Message.FromID})

		if strings.HasPrefix(command, "start") {
			commandLogger.Info("command start triggered")
			go StartCommand(update)
		}

		if strings.HasPrefix(command, "donate") {
			commandLogger.Info("command donate triggered")
			go DonateCommand(update)
		}

		if strings.HasPrefix(command, "save") {
			commandLogger.Info("command save triggered")
			go SaveCommand(update)
		}

		if strings.HasPrefix(command, "me") {
			commandLogger.Info("command me triggered")
			go MeCommand(update)
		}

		if strings.HasPrefix(command, "h_") {
			commandLogger.Info("command h_ triggered")
			go HeroCommand(update)
		}

		if strings.HasPrefix(update.Message.Text, "consoletop") {
			commandLogger.Info("command consoletop triggered")
			go RatingTopCommand(update, "console")
		}

		if strings.HasPrefix(update.Message.Text, "pctop") {
			commandLogger.Info("command pctop triggered")
			go RatingTopCommand(update, "pc")
		}
	}
}
