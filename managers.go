package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gocolly/colly"
)

type Players struct {
	Players []Player `json:"players"`
}

type Player struct {
	KickerName string `json:"kickerName"`
	KickerTeam string `json:"kickerTeam"`
}

func getPdata(players Players) (err error) {
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

	for i := 0; i < len(players.Players); i++ {
		url := "https://www.kicker.de/" + players.Players[i].KickerName + "/spieler/bundesliga/2020-21/" + players.Players[i].KickerTeam
		ctx := colly.NewContext()
		ctx.Put("player", players.Players[i].KickerName)
		c.Request("GET", url, nil, ctx, nil)
	}

	c.Wait()
	return nil
}

func main() {
	// TODO: pass year/season as flag
	playersFile := flag.String("players", "players.json", "Json file with all players to be scanned")
	flag.Parse()
	fmt.Println(*playersFile)

	jsonFile, err := os.Open(*playersFile)
	if err != nil {
		fmt.Println(err)
	}

	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)

	var players Players
	json.Unmarshal(byteValue, &players)

	getPdata(players)
}