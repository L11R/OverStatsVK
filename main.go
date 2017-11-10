package main

import (
	"github.com/sirupsen/logrus"
	"github.com/Dimonchik0036/vk-api"
	r "gopkg.in/gorethink/gorethink.v3"
	"os"
	"strings"
)

var log = logrus.New()

var (
	client  *vkapi.Client
	session *r.Session
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

		if strings.HasPrefix(command, "ratingtop") {
			commandLogger.Info("command ratingtop triggered")
			if strings.HasSuffix(update.Message.Text, "console") {
				go RatingTopCommand(update, "console")
			} else {
				go RatingTopCommand(update, "pc")
			}
		}
	}
}
