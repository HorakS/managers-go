package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gocolly/colly"
)

type Player struct {
	Name       string `json:"name"`
	Team       string `json:"team"`
	Position   string `json:"position"`
	KickerName string `json:"kickerName"`
	KickerTeam string `json:"kickerTeam"`
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
		if e.ChildText("option[selected=selected]") != "2020/21" {
			fmt.Println("No data yet for " + e.Request.Ctx.Get("player") + " this season")
			return
		}
		e.ForEach("tr", func(_ int, row *colly.HTMLElement) {
			fmt.Println(row.ChildText("td:nth-child(3)"))
		})
	})

	for i := 0; i < len(players); i++ {
		url := "https://www.kicker.de/" + players[i].KickerName + "/spieler/bundesliga/2020-21/" + players[i].KickerTeam
		ctx := colly.NewContext()
		ctx.Put("player", players[i].KickerName)
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
