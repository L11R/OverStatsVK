package main

import (
	"fmt"
	"strings"
	"github.com/Dimonchik0036/vk-api"
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
		profile, err := GetOverwatchProfile(info[1], info[2])
		if err != nil {
			log.Warn(err)
			text = "ERROR:\n" + fmt.Sprint(err)
		} else {
			_, err := InsertUser(User{
				Id:      update.Message.FromID,
				Profile: profile,
				Region:  info[1],
				Nick:    info[2],
			})
			if err != nil {
				log.Warn(err)
				text = "ERROR:\n" + fmt.Sprint(err)
			} else {
				log.Info("save command executed successful")
				text = "Saved!"
			}
		}
	} else {
		text = "Example: save eu|us|kr|psn|xbl BattleTag-1337|ConsoleLogin (sic, hyphen!)"
	}

	msg := vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID), text)
	client.SendMessage(msg)
}

func MeCommand(update vkapi.LPUpdate) {
	user, err := GetUser(update.Message.FromID)
	var text string

	if err != nil {
		log.Warn(err)
		text = fmt.Sprint("ERROR:\n", err)
	} else {
		log.Info("me command executed successful")
		text = MakeSummary(user)
	}

	msg := vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID), text)
	client.SendMessage(msg)
}

func HeroCommand(update vkapi.LPUpdate) {
	user, err := GetUser(update.Message.FromID)
	var text string

	if err != nil {
		log.Warn(err)
		text = "ERROR:\n" + fmt.Sprint(err)
	} else {
		log.Info("h_ command executed successful")
		hero := strings.Split(update.Message.Text, "_")[1]

		text = MakeHeroSummary(hero, user)
	}

	msg := vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID), text)
	client.SendMessage(msg)
}

func RatingTopCommand(update vkapi.LPUpdate, platform string) {
	top, err := GetRatingTop(platform, 20)
	var text string

	if err != nil {
		log.Warn(err)
		text = "ERROR:\n" + fmt.Sprint(err)
	} else {
		text = "Rating Top:\n"
		for i := range top {
			text += fmt.Sprintf("%d. %s (%d)\n", i+1, top[i].Nick, top[i].Profile.Rating)
		}
	}

	msg := vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID), text)
	client.SendMessage(msg)
}

func SessionReport(change Change) {
	// Check OldVal and NewOld existing
	if change.OldVal.Profile != nil && change.NewVal.Profile != nil {
		oldStats := Report{
			Rating: change.OldVal.Profile.Rating,
			Level:  change.OldVal.Profile.Prestige*100 + change.OldVal.Profile.Level,
		}

		if competitiveStats, ok := change.OldVal.Profile.CompetitiveStats.CareerStats["allHeroes"]; ok {
			if gamesPlayed, ok := competitiveStats.Game["gamesPlayed"]; ok {
				oldStats.Games = int(gamesPlayed.(float64))
			}
			if gamesWon, ok := competitiveStats.Game["gamesWon"]; ok {
				oldStats.Wins = int(gamesWon.(float64))
			}
			if gamesTied, ok := competitiveStats.Miscellaneous["gamesTied"]; ok {
				oldStats.Ties = int(gamesTied.(float64))
			}
			if gamesLost, ok := competitiveStats.Miscellaneous["gamesLost"]; ok {
				oldStats.Losses = int(gamesLost.(float64))
			}
		}

		newStats := Report{
			Rating: change.NewVal.Profile.Rating,
			Level:  change.NewVal.Profile.Prestige*100 + change.NewVal.Profile.Level,
		}

		if competitiveStats, ok := change.NewVal.Profile.CompetitiveStats.CareerStats["allHeroes"]; ok {
			if gamesPlayed, ok := competitiveStats.Game["gamesPlayed"]; ok {
				newStats.Games = int(gamesPlayed.(float64))
			}
			if gamesWon, ok := competitiveStats.Game["gamesWon"]; ok {
				newStats.Wins = int(gamesWon.(float64))
			}
			if gamesTied, ok := competitiveStats.Miscellaneous["gamesTied"]; ok {
				newStats.Ties = int(gamesTied.(float64))
			}
			if gamesLost, ok := competitiveStats.Miscellaneous["gamesLost"]; ok {
				newStats.Losses = int(gamesLost.(float64))
			}
		}

		diffStats := Report{
			newStats.Rating - oldStats.Rating,
			newStats.Level - oldStats.Level,
			newStats.Games - oldStats.Games,
			newStats.Wins - oldStats.Wins,
			newStats.Ties - oldStats.Ties,
			newStats.Losses - oldStats.Losses,
		}

		if diffStats.Games != 0 {
			log.Infof("sending report to %d", change.NewVal.Id)
			text := "Session Report\n\n"

			text += AddInfo("Rating", oldStats.Rating, newStats.Rating, diffStats.Rating)
			text += AddInfo("Wins", oldStats.Wins, newStats.Wins, diffStats.Wins)
			text += AddInfo("Losses", oldStats.Losses, newStats.Losses, diffStats.Losses)
			text += AddInfo("Ties", oldStats.Ties, newStats.Ties, diffStats.Ties)
			text += AddInfo("Level", oldStats.Level, newStats.Level, diffStats.Level)

			msg := vkapi.NewMessage(vkapi.NewDstFromUserID(change.NewVal.Id), text)
			client.SendMessage(msg)
		}
	}
}
