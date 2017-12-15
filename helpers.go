package main

import (
	"errors"
	"fmt"
	"github.com/sdwolfe32/ovrstat/ovrstat"
	r "gopkg.in/gorethink/gorethink.v3"
	"sort"
	"strings"
)

// Make small text summary based on profile
func MakeSummary(user User, top Top, mode string) string {
	text := fmt.Sprintf("%s (%d sr / %d lvl)\n", user.Patreon+user.Profile.Name, user.Profile.Rating, user.Profile.Prestige*100+user.Profile.Level)

	var stats ovrstat.StatsCollection
	if mode == "CompetitiveStats" {
		stats = user.Profile.CompetitiveStats
	}
	if mode == "QuickPlayStats" {
		stats = user.Profile.QuickPlayStats
	}

	if careerStats, ok := stats.CareerStats["allHeroes"]; ok {
		var basicStats Report
		if gamesPlayed, ok := careerStats.Game["gamesPlayed"]; ok {
			basicStats.Games = int(gamesPlayed.(float64))
		}
		if gamesWon, ok := careerStats.Game["gamesWon"]; ok {
			basicStats.Wins = int(gamesWon.(float64))
		}
		if gamesTied, ok := careerStats.Game["gamesTied"]; ok {
			basicStats.Ties = int(gamesTied.(float64))
		}
		if gamesLost, ok := careerStats.Game["gamesLost"]; ok {
			basicStats.Losses = int(gamesLost.(float64))
		}

		if mode == "CompetitiveStats" {
			text += fmt.Sprintf("%d-%d-%d / %0.2f%% winrate\n", basicStats.Wins, basicStats.Losses, basicStats.Ties, float64(basicStats.Wins)/float64(basicStats.Games)*100)
		} else if mode == "QuickPlayStats" {
			text += fmt.Sprintf("%d wins\n", basicStats.Wins)
		}

		// Temp struct for k/d counting
		type KD struct {
			Eliminations float64
			Deaths       float64
			Ratio        float64
		}

		var kd KD

		if eliminations, ok := careerStats.Combat["eliminations"]; ok {
			kd.Eliminations = eliminations.(float64)
		}
		if deaths, ok := careerStats.Deaths["deaths"]; ok {
			kd.Deaths = deaths.(float64)
		}

		if kd.Deaths > 0 {
			kd.Ratio = kd.Eliminations / kd.Deaths
			text += fmt.Sprintf("%0.2f k/d\n\n", kd.Ratio)
		}

		if mode == "CompetitiveStats" {
			text += fmt.Sprintf("Rating Top:\n#%d (%0.2f%%)\n\n", top.Place, top.Rank)
		}

		text += "7 top played heroes:\n"
		var topPlayedHeroes Heroes
		for name, elem := range stats.TopHeroes {
			topPlayedHeroes = append(topPlayedHeroes, Hero{
				Name:                name,
				TimePlayedInSeconds: elem.TimePlayedInSeconds,
			})
		}

		// Sort top played heroes in descending
		sort.Sort(sort.Reverse(topPlayedHeroes))

		for i := 0; i < 7; i++ {
			var format string
			if mode == "CompetitiveStats" {
				format = fmt.Sprint("%s (%s) /h_%s\n")
			} else if mode == "QuickPlayStats" {
				format = fmt.Sprint("%s (%s) /h_%s_quick\n")
			}

			text += fmt.Sprintf(
				format,
				strings.Title(strings.ToLower(topPlayedHeroes[i].Name)),
				stats.TopHeroes[topPlayedHeroes[i].Name].TimePlayed,
				topPlayedHeroes[i].Name,
			)
		}
	}

	text += fmt.Sprint("\nLast Updated:\n", user.Date.Format("15:04:05 / 02.01.2006 MST"))

	return text
}

func MakeHeroSummary(hero string, mode string, user User) string {
	text := fmt.Sprintf("%s", strings.Title(strings.ToLower(hero)))

	var stats ovrstat.StatsCollection
	if mode == "CompetitiveStats" {
		stats = user.Profile.CompetitiveStats
	}
	if mode == "QuickPlayStats" {
		stats = user.Profile.QuickPlayStats
	}

	if heroStats, ok := stats.CareerStats[hero]; ok {
		if heroAdditionalStats, ok := stats.TopHeroes[hero]; ok {
			text += fmt.Sprintf(" (%s)\n", heroAdditionalStats.TimePlayed)
			if cards, ok := heroStats.MatchAwards["cards"]; ok {
				text += fmt.Sprintf("🃏%0.0f ", cards)
			}
			if medalsGold, ok := heroStats.MatchAwards["medalsGold"]; ok {
				text += fmt.Sprintf("🥇%0.0f ", medalsGold)
			}
			if medalsSilver, ok := heroStats.MatchAwards["medalsSilver"]; ok {
				text += fmt.Sprintf("🥈%0.0f ", medalsSilver)
			}
			if medalsBronze, ok := heroStats.MatchAwards["medalsBronze"]; ok {
				text += fmt.Sprintf("🥉%0.0f ", medalsBronze)
			}

			text += "\n"

			if mode == "CompetitiveStats" {
				text += fmt.Sprintf("%d%% hero winrate", heroAdditionalStats.WinPercentage)

				res, err := GetRank(
					user.Id,
					r.Row.Field("profile").Field(mode).Field("TopHeroes").Field(hero).Field("WinPercentage"),
				)
				if err != nil {
					text += fmt.Sprint(" (error)\n")
				} else {
					text += fmt.Sprintf(" (#%d, %0.0f%%)\n", res.Place, res.Rank)
				}
			}

			if eliminationsPerLife, ok := heroStats.Combat["eliminationsPerLife"]; ok {
				text += fmt.Sprintf("%0.2f k/d ratio", eliminationsPerLife)

				res, err := GetRank(
					user.Id,
					r.Row.Field("profile").Field(mode).Field("CareerStats").Field(hero).Field("Combat").Field("eliminationsPerLife"),
				)
				if err != nil {
					text += fmt.Sprint(" (error)\n")
				} else {
					text += fmt.Sprintf(" (#%d, %0.0f%%)\n", res.Place, res.Rank)
				}
			}

			if accuracy, ok := heroStats.Combat["weaponAccuracy"]; ok {
				text += fmt.Sprintf("%s accuracy", accuracy)

				res, err := GetRank(
					user.Id,
					r.Row.Field("profile").Field(mode).Field("CareerStats").Field(hero).Field("Combat").Field("weaponAccuracy"),
				)
				if err != nil {
					text += fmt.Sprint(" (error)\n")
				} else {
					text += fmt.Sprintf(" (#%d, %0.0f%%)\n", res.Place, res.Rank)
				}
			}

			timePlayedInMinutes := float64(heroAdditionalStats.TimePlayedInSeconds) / 60

			if eliminations, ok := heroStats.Combat["eliminations"]; ok {
				eliminationsPerMin := eliminations.(float64) / timePlayedInMinutes
				text += fmt.Sprintf("%0.2f eliminations per min\n", eliminationsPerMin)
			}

			if damageDone, ok := heroStats.Combat["damageDone"]; ok {
				damagePerMin := damageDone.(float64) / timePlayedInMinutes
				text += fmt.Sprintf("%0.0f damage per min\n", damagePerMin)
			}

			if blocked, ok := heroStats.Miscellaneous["damageBlocked"]; ok {
				blockedPerMin := blocked.(float64) / timePlayedInMinutes
				text += fmt.Sprintf("%0.0f blocked per min\n", blockedPerMin)
			}

			if healing, ok := heroStats.Miscellaneous["healingDone"]; ok {
				healingPerMin := healing.(float64) / timePlayedInMinutes
				text += fmt.Sprintf("%0.0f healing per min\n", healingPerMin)
			}

			if objKills, ok := heroStats.Combat["objectiveKills"]; ok {
				objKillsPerMin := objKills.(float64) / timePlayedInMinutes
				text += fmt.Sprintf("%0.2f obj. kills per min\n", objKillsPerMin)
			}

			if crits, ok := heroStats.Combat["criticalHits"]; ok {
				critsPerMin := crits.(float64) / timePlayedInMinutes
				text += fmt.Sprintf("%0.2f crits per min\n", critsPerMin)
			}

			// HERO SPECIFIC
			text += "\nHero Specific:\n"
			switch hero {
			case "ana":
				if scopedAccuracy, ok := heroStats.HeroSpecific["scopedAccuracy"]; ok {
					text += fmt.Sprintf("%s scoped accuracy\n", scopedAccuracy)
				}
				if enemiesSlept, ok := heroStats.Miscellaneous["enemiesSlept"]; ok {
					enemiesSleptPerMin := enemiesSlept.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f enemies slept per min\n", enemiesSleptPerMin)
				}
			case "bastion":
				if reconKills, ok := heroStats.HeroSpecific["reconKills"]; ok {
					reconKillsPerMin := reconKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f recon kills per min\n", reconKillsPerMin)
				}
				if sentryKills, ok := heroStats.HeroSpecific["sentryKills"]; ok {
					sentryKillsPerMin := sentryKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f sentry kills per min\n", sentryKillsPerMin)
				}
				if tankKills, ok := heroStats.HeroSpecific["tankKills"]; ok {
					tankKillsPerMin := tankKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f tank kills per min\n", tankKillsPerMin)
				}
			case "dVa":
				if blocked, ok := heroStats.HeroSpecific["damageBlocked"]; ok {
					blockedPerMin := blocked.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.0f blocked per min\n", blockedPerMin)
				}
				if mechsCalled, ok := heroStats.HeroSpecific["mechsCalled"]; ok {
					mechsCalledPerMin := mechsCalled.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f mechs called per min\n", mechsCalledPerMin)
				}
				if mechDeaths, ok := heroStats.HeroSpecific["mechDeaths"]; ok {
					mechDeathsPerMin := mechDeaths.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f mech deaths per min\n", mechDeathsPerMin)
				}
				if selfDestructKills, ok := heroStats.Miscellaneous["selfDestructKills"]; ok {
					selfDestructKillsPerMin := selfDestructKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f self destruct kills per min\n", selfDestructKillsPerMin)
				}
			case "doomfist":
				if abilityDamageDone, ok := heroStats.HeroSpecific["abilityDamageDone"]; ok {
					abilityDamageDonePerMin := abilityDamageDone.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.0f ability damage done per min\n", abilityDamageDonePerMin)
				}
				if meteorStrikeKills, ok := heroStats.HeroSpecific["meteorStrikeKills"]; ok {
					meteorStrikeKillsPerMin := meteorStrikeKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f meteor strike kills per min\n", meteorStrikeKillsPerMin)
				}
				if shieldsCreated, ok := heroStats.HeroSpecific["shieldsCreated"]; ok {
					shieldsCreatedPerMin := shieldsCreated.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.0f shields created per min\n", shieldsCreatedPerMin)
				}
			case "genji":
				if damageReflected, ok := heroStats.HeroSpecific["damageReflected"]; ok {
					damageReflectedPerMin := damageReflected.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.0f damage reflected per min\n", damageReflectedPerMin)
				}
				if dragonbladesKills, ok := heroStats.HeroSpecific["dragonbladesKills"]; ok {
					dragonbladesKillsPerMin := dragonbladesKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f dragonblades kills per min\n", dragonbladesKillsPerMin)
				}
			case "hanzo":
				if dragonstrikeKills, ok := heroStats.HeroSpecific["dragonstrikeKills"]; ok {
					dragonstrikeKillsPerMin := dragonstrikeKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f dragonstrike kills per min\n", dragonstrikeKillsPerMin)
				}
				if scatterArrowKills, ok := heroStats.HeroSpecific["scatterArrowKills"]; ok {
					scatterArrowKillsPerMin := scatterArrowKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f scatter arrow kills per min\n", scatterArrowKillsPerMin)
				}
			case "junkrat":
				if enemiesTrapped, ok := heroStats.HeroSpecific["enemiesTrapped"]; ok {
					enemiesTrappedPerMin := enemiesTrapped.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f enemies trapped per min\n", enemiesTrappedPerMin)
				}
				if ripTireKills, ok := heroStats.HeroSpecific["ripTireKills"]; ok {
					ripTireKillsPerMin := ripTireKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f rip tire kills per min\n", ripTireKillsPerMin)
				}
			case "lucio":
				if soundBarriersProvided, ok := heroStats.HeroSpecific["soundBarriersProvided"]; ok {
					soundBarriersProvidedPerMin := soundBarriersProvided.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f sound barriers provided per min\n", soundBarriersProvidedPerMin)
				}
			case "mccree":
				if deadeyeKills, ok := heroStats.HeroSpecific["deadeyeKills"]; ok {
					deadeyeKillsPerMin := deadeyeKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f deadeye kills per min\n", deadeyeKillsPerMin)
				}
				if fanTheHammerKills, ok := heroStats.HeroSpecific["fanTheHammerKills"]; ok {
					fanTheHammerKillsPerMin := fanTheHammerKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f fan the hammer kills per min\n", fanTheHammerKillsPerMin)
				}
			case "mei":
				if damageBlocked, ok := heroStats.HeroSpecific["damageBlocked"]; ok {
					damageBlockedPerMin := damageBlocked.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.0f damage blocked per min\n", damageBlockedPerMin)
				}
				if blizzardKills, ok := heroStats.HeroSpecific["blizzardKills"]; ok {
					blizzardKillsPerMin := blizzardKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f blizzard kills per min\n", blizzardKillsPerMin)
				}
				if enemiesFrozen, ok := heroStats.HeroSpecific["enemiesFrozen"]; ok {
					enemiesFrozenPerMin := enemiesFrozen.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f enemies frozen per min\n", enemiesFrozenPerMin)
				}
			case "mercy":
				if damageAmplified, ok := heroStats.Miscellaneous["damageAmplified"]; ok {
					damageAmplifiedPerMin := damageAmplified.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.0f damage amplified per min\n", damageAmplifiedPerMin)
				}
				if blasterKills, ok := heroStats.Miscellaneous["blasterKills"]; ok {
					blasterKillsPerMin := blasterKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f blaster kills per min\n", blasterKillsPerMin)
				}
				if playersResurrected, ok := heroStats.HeroSpecific["playersResurrected"]; ok {
					playersResurrectedPerMin := playersResurrected.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f players resurrected per min\n", playersResurrectedPerMin)
				}
			case "moira":
				if coalescenceKills, ok := heroStats.Miscellaneous["coalescenceKills"]; ok {
					coalescenceKillsPerMin := coalescenceKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f coalescence kills per min\n", coalescenceKillsPerMin)
				}
				if coalescenceHealing, ok := heroStats.Miscellaneous["coalescenceHealing"]; ok {
					coalescenceHealingPerMin := coalescenceHealing.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.0f coalescence healing done per min\n", coalescenceHealingPerMin)
				}
			case "orisa":
				if damageAmplified, ok := heroStats.Miscellaneous["damageAmplified"]; ok {
					damageAmplifiedPerMin := damageAmplified.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.0f damage amplified per min\n", damageAmplifiedPerMin)
				}
			case "pharah":
				if barrageKills, ok := heroStats.HeroSpecific["barrageKills"]; ok {
					barrageKillsPerMin := barrageKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f barrage kills per min\n", barrageKillsPerMin)
				}
				if rocketDirectHits, ok := heroStats.HeroSpecific["rocketDirectHits"]; ok {
					rocketDirectHits := rocketDirectHits.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f rocket direct hits per min\n", rocketDirectHits)
				}
			case "reaper":
				if deathsBlossomKills, ok := heroStats.HeroSpecific["deathsBlossomKills"]; ok {
					deathsBlossomKillsPerMin := deathsBlossomKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f blossom kills per min\n", deathsBlossomKillsPerMin)
				}
				if selfHealing, ok := heroStats.Assists["selfHealing"]; ok {
					selfHealingPerMin := selfHealing.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.0f self healing per min\n", selfHealingPerMin)
				}
			case "reinhardt":
				if damageBlocked, ok := heroStats.HeroSpecific["damageBlocked"]; ok {
					damageBlockedPerMin := damageBlocked.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.0f damage blocked per min\n", damageBlockedPerMin)
				}
				if chargeKills, ok := heroStats.HeroSpecific["chargeKills"]; ok {
					chargeKillsPerMin := chargeKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f charge kills per min\n", chargeKillsPerMin)
				}
				if fireStrikeKills, ok := heroStats.HeroSpecific["fireStrikeKills"]; ok {
					fireStrikeKillsPerMin := fireStrikeKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f fire strike kills per min\n", fireStrikeKillsPerMin)
				}
				if earthshatterKills, ok := heroStats.HeroSpecific["earthshatterKills"]; ok {
					earthshatterKillsPerMin := earthshatterKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f earthshatter kills per min\n", earthshatterKillsPerMin)
				}
			case "roadhog":
				if enemiesHooked, ok := heroStats.HeroSpecific["enemiesHooked"]; ok {
					enemiesHookedPerMin := enemiesHooked.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f enemies hooked per min\n", enemiesHookedPerMin)
				}
				if wholeHogKills, ok := heroStats.HeroSpecific["wholeHogKills"]; ok {
					wholeHogKillsPerMin := wholeHogKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f whole hog kills per min\n", wholeHogKillsPerMin)
				}
				if hookAccuracy, ok := heroStats.HeroSpecific["hookAccuracy"]; ok {
					text += fmt.Sprintf("%s hook accuracy\n", hookAccuracy)
				}
			case "soldier76":
				if helixRocketsKills, ok := heroStats.HeroSpecific["helixRocketsKills"]; ok {
					helixRocketsKillsPerMin := helixRocketsKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f helix rockets kills per min\n", helixRocketsKillsPerMin)
				}
				if tacticalVisorKills, ok := heroStats.HeroSpecific["tacticalVisorKills"]; ok {
					tacticalVisorKillsPerMin := tacticalVisorKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f tactical visor kills per min\n", tacticalVisorKillsPerMin)
				}
				if bioticFieldHealingDone, ok := heroStats.HeroSpecific["bioticFieldHealingDone"]; ok {
					bioticFieldHealingDonePerMin := bioticFieldHealingDone.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.0f healing done per min\n", bioticFieldHealingDonePerMin)
				}
			case "sombra":
				if enemiesHacked, ok := heroStats.Miscellaneous["enemiesHacked"]; ok {
					enemiesHackedPerMin := enemiesHacked.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f enemies hacked per min\n", enemiesHackedPerMin)
				}
				if enemiesEmpd, ok := heroStats.Miscellaneous["enemiesEmpd"]; ok {
					enemiesEmpdPerMin := enemiesEmpd.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f enemies emp'd per min\n", enemiesEmpdPerMin)
				}
			case "symmetra":
				if playersTeleported, ok := heroStats.HeroSpecific["playersTeleported"]; ok {
					playersTeleportedPerMin := playersTeleported.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f players teleported per min\n", playersTeleportedPerMin)
				}
				if sentryTurretsKills, ok := heroStats.HeroSpecific["sentryTurretsKills"]; ok {
					sentryTurretsKillsPerMin := sentryTurretsKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f sentry turrets kills per min\n", sentryTurretsKillsPerMin)
				}
			case "torbjorn":
				if torbjornKills, ok := heroStats.HeroSpecific["torbjornKills"]; ok {
					torbjornKillsPerMin := torbjornKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f torbjorn kills per min\n", torbjornKillsPerMin)
				}
				if moltenCoreKills, ok := heroStats.HeroSpecific["moltenCoreKills"]; ok {
					moltenCoreKillsPerMin := moltenCoreKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f molten core kills per min\n", moltenCoreKillsPerMin)
				}
				if turretsKills, ok := heroStats.HeroSpecific["turretsKills"]; ok {
					turretsKillsPerMin := turretsKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f turrets kills per min\n", turretsKillsPerMin)
				}
				if armorPacksCreated, ok := heroStats.HeroSpecific["armorPacksCreated"]; ok {
					armorPacksCreatedPerMin := armorPacksCreated.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f armor packs created per min\n", armorPacksCreatedPerMin)
				}
			case "tracer":
				if pulseBombsKills, ok := heroStats.HeroSpecific["pulseBombsKills"]; ok {
					pulseBombsKillsPerMin := pulseBombsKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f pulse bombs kills per min\n", pulseBombsKillsPerMin)
				}
				if pulseBombsAttached, ok := heroStats.HeroSpecific["pulseBombsAttached"]; ok {
					pulseBombsAttachedPerMin := pulseBombsAttached.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f pulse bombs attached per min\n", pulseBombsAttachedPerMin)
				}
			case "widowmaker":
				if scopedCriticalHits, ok := heroStats.HeroSpecific["scopedCriticalHits"]; ok {
					scopedCriticalHitsPerMin := scopedCriticalHits.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f scoped critical hits per min\n", scopedCriticalHitsPerMin)
				}
				if scopedAccuracy, ok := heroStats.HeroSpecific["scopedAccuracy"]; ok {
					text += fmt.Sprintf("%s scoped accuracy\n", scopedAccuracy)
				}
			case "winston":
				if damageBlocked, ok := heroStats.HeroSpecific["damageBlocked"]; ok {
					damageBlockedPerMin := damageBlocked.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.0f damage blocked per min\n", damageBlockedPerMin)
				}
				if jumpPackKills, ok := heroStats.HeroSpecific["jumpPackKills"]; ok {
					jumpPackKillsPerMin := jumpPackKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f jump pack kills per min\n", jumpPackKillsPerMin)
				}
				if primalRageKills, ok := heroStats.Miscellaneous["primalRageKills"]; ok {
					primalRageKillsPerMin := primalRageKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f primal rage kills per min\n", primalRageKillsPerMin)
				}
				if playersKnockedBack, ok := heroStats.HeroSpecific["playersKnockedBack"]; ok {
					playersKnockedBackPerMin := playersKnockedBack.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f players knocked back per min\n", playersKnockedBackPerMin)
				}
			case "zarya":
				if damageBlocked, ok := heroStats.HeroSpecific["damageBlocked"]; ok {
					damageBlockedPerMin := damageBlocked.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.0f damage blocked per min\n", damageBlockedPerMin)
				}
				if highEnergyKills, ok := heroStats.HeroSpecific["highEnergyKills"]; ok {
					highEnergyKillsPerMin := highEnergyKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f high energy kills per min\n", highEnergyKillsPerMin)
				}
				if gravitonSurgeKills, ok := heroStats.HeroSpecific["gravitonSurgeKills"]; ok {
					gravitonSurgeKillsPerMin := gravitonSurgeKills.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f graviton surge kills per min\n", gravitonSurgeKillsPerMin)
				}
				if projectedBarriersApplied, ok := heroStats.HeroSpecific["projectedBarriersApplied"]; ok {
					projectedBarriersAppliedPerMin := projectedBarriersApplied.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f projected barriers applied per min\n", projectedBarriersAppliedPerMin)
				}
				if averageEnergy, ok := heroStats.HeroSpecific["averageEnergy"]; ok {
					text += fmt.Sprintf("%0.0f%% average energy\n", averageEnergy.(float64)*100)
				}
			case "zenyatta":
				if transcendenceHealing, ok := heroStats.Miscellaneous["transcendenceHealing"]; ok {
					transcendenceHealingPerMin := transcendenceHealing.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.0f transcendence healing per min\n", transcendenceHealingPerMin)
				}
				if offensiveAssists, ok := heroStats.Assists["offensiveAssists"]; ok {
					offensiveAssistsPerMin := offensiveAssists.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f offensive assists per min\n", offensiveAssistsPerMin)
				}
				if defensiveAssists, ok := heroStats.Miscellaneous["defensiveAssists"]; ok {
					defensiveAssistsPerMin := defensiveAssists.(float64) / timePlayedInMinutes
					text += fmt.Sprintf("%0.2f defensive assists per min\n", defensiveAssistsPerMin)
				}
			}
		} else {
			text += "\nNOT AVAILABLE"
		}
	} else {
		text += "\nNOT AVAILABLE"
	}

	text += fmt.Sprint("\nLast Updated:\n", user.Date.Format("15:04:05 / 02.01.2006 MST"))

	return text
}

// Fetch Overwatch profile based on region and BattleTag / PSN ID / Xbox Live Account
func GetOverwatchProfile(region string, nick string) (*ovrstat.PlayerStats, error) {
	if region == "eu" || region == "us" || region == "kr" {
		return ovrstat.PCStats(region, nick)
	} else if region == "psn" || region == "xbl" {
		return ovrstat.ConsoleStats(region, nick)
	}

	return nil, errors.New("region is wrong")
}
