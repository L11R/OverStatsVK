package main

import (
	"fmt"
	"github.com/Dimonchik0036/vk-api"
	"strconv"
	"strings"
)

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
			if gamesTied, ok := competitiveStats.Game["gamesTied"]; ok {
				oldStats.Ties = int(gamesTied.(float64))
			}
			if gamesLost, ok := competitiveStats.Game["gamesLost"]; ok {
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
			if gamesTied, ok := competitiveStats.Game["gamesTied"]; ok {
				newStats.Ties = int(gamesTied.(float64))
			}
			if gamesLost, ok := competitiveStats.Game["gamesLost"]; ok {
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

		AddDiffString := func(name string, oldInfo int, newInfo int, diffInfo int) string {
			text := fmt.Sprintf("%s:\n%d | %d |", name, oldInfo, newInfo)
			if diffInfo > 0 {
				text += fmt.Sprintf(" +%d 📈\n", diffInfo)
			} else if diffInfo == 0 {
				text += fmt.Sprintf(" %d —\n", diffInfo)
			} else {
				text += fmt.Sprintf(" %d 📉\n", diffInfo)
			}

			return text
		}

		if diffStats.Games > 0 || diffStats.Level != 0 {
			log.Infof("sending report to %s", change.NewVal.Id)
			text := "Session Report\n\n"

			text += AddDiffString("Rating", oldStats.Rating, newStats.Rating, diffStats.Rating)
			text += AddDiffString("Wins", oldStats.Wins, newStats.Wins, diffStats.Wins)
			text += AddDiffString("Losses", oldStats.Losses, newStats.Losses, diffStats.Losses)
			text += AddDiffString("Ties", oldStats.Ties, newStats.Ties, diffStats.Ties)
			text += AddDiffString("Level", oldStats.Level, newStats.Level, diffStats.Level)

			id, _ := strconv.ParseInt(strings.Split(change.NewVal.Id, ":")[1], 10, 64)
			msg := vkapi.NewMessage(vkapi.NewDstFromUserID(id), text)
			client.SendMessage(msg)
		}
	}
}
