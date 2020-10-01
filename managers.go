package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gocolly/colly"
)

type Match struct {
	Day    int    `json:"day"` // TODO: need better name
	Team1  string `json:"team1"`
	Team2  string `json:"team2"`
	Result string `json:"result"` //TODO: separate struct? Half time result?
}

type Pdata struct {
	Match    Match   `json:"match"`
	Grade    float32 `json:"grade"`
	Scp      int     `json:"scp"`
	Playtime int     `json:"playtime"`
	SubIn    int     `json:"sub-in"`
	SubOut   int     `json:"sub-out"`
}

type Player struct {
	Name       string `json:"name"`
	Team       string `json:"team"`
	Position   string `json:"position"`
	KickerName string `json:"kickerName"`
	KickerTeam string `json:"kickerTeam"`
	Data       Pdata  `json:"data"`
}

func getPdata(players []Player) (err error) {
	c := colly.NewCollector(
		colly.AllowedDomains("www.kicker.de"),
		colly.Async(true),
	)

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("visiting", r.URL.String())
	})

	c.OnHTML("div.kick__vita__statistic", func(e *colly.HTMLElement) {
		player := e.Request.Ctx.GetAny("player").(Player)
		if e.ChildText("option[selected=selected]") != "2020/21" {
			fmt.Println("No data yet for " + player.Name + " this season")
			return
		}

		bundesliga := false

		e.ForEachWithBreak("tr", func(_ int, row *colly.HTMLElement) bool {
			match := row.ChildText("td:nth-child(1)")
			if match == "Bundesliga" {
				bundesliga = true
				return true
			}

			if !bundesliga || match == "" {
				return true
			}

			if !strings.Contains(match, "Spieltag") {
				return false
			}

			grade := row.ChildText("td:nth-child(2)")
			scp := row.ChildText("td:nth-child(6)")
			subIn := row.ChildText("td:nth-child(7)")
			subOut := row.ChildText("td:nth-child(8)")
			fmt.Println("Match:", match)
			fmt.Println("Note:", grade, "SCP:", scp, "in:", subIn, "out:", subOut)

			return true
		})
	})

	for i := 0; i < len(players); i++ {
		url := "https://www.kicker.de/" + players[i].KickerName + "/spieler/bundesliga/2020-21/" + players[i].KickerTeam
		ctx := colly.NewContext()
		ctx.Put("player", players[i])
		c.Request("GET", url, nil, ctx, nil)
	}

	c.Wait()
	return nil
}

func main() {
	// TODO: pass year/season as flag
	playersFile := flag.String("players", "players.json", "Json file with all players to be scanned")
	flag.Parse()

	jsonFile, err := os.Open(*playersFile)
	if err != nil {
		fmt.Println(err)
	}

	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)

	var players []Player
	json.Unmarshal(byteValue, &players)
	getPdata(players)
}
