package main

import (
	"fmt"
	"github.com/Dimonchik0036/vk-api"
	"strings"
)

func StartCommand(update vkapi.LPUpdate) {
	msg := vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID), "Simple bot for Overwatch by @kraso\n\n"+
		"How to use:\n"+
		"1. Use \"save\" to save your game profile.\n"+
		"2. Use \"me\" to see your stats.\n"+
		"3. ???\n"+
		"4. PROFIT!\n\n"+
		"Features:\n"+
		"— Player profile (me command)\n"+
		"— Small summary for heroes\n"+
		"— Reports after every game session\n")
	client.SendMessage(msg)

	log.Info("start command executed successful")
}

func DonateCommand(update vkapi.LPUpdate) {
	msg := vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID), "If you find this bot helpful, "+
		"you can make small donation to help me pay server bills! https://paypal.me/krasovsky")
	client.SendMessage(msg)

	log.Info("donate command executed successful")
}

type Hero struct {
	Name                string
	TimePlayedInSeconds int
}

type Heroes []Hero

func (hero Heroes) Len() int {
	return len(hero)
}

func (hero Heroes) Less(i, j int) bool {
	return hero[i].TimePlayedInSeconds < hero[j].TimePlayedInSeconds
}

func (hero Heroes) Swap(i, j int) {
	hero[i], hero[j] = hero[j], hero[i]
}

func SaveCommand(update vkapi.LPUpdate) {
	info := strings.Split(update.Message.Text, " ")
	var text string

	if len(info) == 3 {
		if info[1] != "psn" && info[1] != "xbl" {
			info[2] = strings.Replace(info[2], "#", "-", -1)
		}

		profile, err := GetOverwatchProfile(info[1], info[2])
		if err != nil {
			log.Warn(err)
			text = "Player not found!"
		} else {
			_, err := InsertUser(User{
				Id:      update.Message.FromID,
				Profile: profile,
				Region:  info[1],
				Nick:    info[2],
			})
			if err != nil {
				log.Warn(err)
				return
			}

			log.Info("save command executed successful")
			text = "Saved!"
		}
	} else {
		text = "Example: save eu|us|kr|psn|xbl BattleTag#1337|ConsoleLogin (sic, hyphen!)"
	}

	msg := vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID), text)
	client.SendMessage(msg)
}

func MeCommand(update vkapi.LPUpdate) {
	user, err := GetUser(update.Message.FromID)
	if err != nil {
		log.Warn(err)
		return
	}

	place, err := GetRatingPlace(update.Message.FromID)
	if err != nil {
		log.Warn(err)
		return
	}

	log.Info("me command executed successful")
	text := MakeSummary(user, place)

	msg := vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID), text)
	client.SendMessage(msg)
}

func HeroCommand(update vkapi.LPUpdate) {
	user, err := GetUser(update.Message.FromID)
	if err != nil {
		log.Warn(err)
		return
	}

	log.Info("h_ command executed successful")
	hero := strings.Split(update.Message.Text, "_")[1]
	text := MakeHeroSummary(hero, user)

	msg := vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID), text)
	client.SendMessage(msg)
}

func RatingTopCommand(update vkapi.LPUpdate, platform string) {
	top, err := GetRatingTop(platform, 20)
	if err != nil {
		log.Warn(err)
		return
	}

	text := "Rating Top:\n"
	for i := range top {
		nick := top[i].Nick
		if top[i].Region != "psn" && top[i].Region != "xbl" {
			nick = strings.Replace(nick, "-", "#", -1)
		}
		text += fmt.Sprintf("%d. %s (%d)\n", i+1, nick, top[i].Profile.Rating)
	}

	msg := vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID), text)
	client.SendMessage(msg)
}
