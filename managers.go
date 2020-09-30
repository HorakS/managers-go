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

func getPdata() (players Players, err error) {
	skip := false
	c := colly.NewCollector(
		colly.AllowedDomains("www.kicker.de"),
		colly.Async(true),
	)

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("visiting", r.URL.String())
	})

	c.OnHTML("option[selected=selected]", func(e *colly.HTMLElement) {
		if e.Text != "2020/21" {
			skip = true
		}
	})

	c.OnHTML("table[data-target=playerstatisticstable]", func(e *colly.HTMLElement) {
		fmt.Println("found table")
		e.ForEach("tr", func(_ int, row *colly.HTMLElement) {
			fmt.Println(row.ChildText("td:nth-child(3)"))
		})
	})

	c.Visit("https://www.kicker.de/timo-werner/spieler/bundesliga/2019-20/rb-leipzig")
	c.Wait()
	return players, nil
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

	for i := 0; i < len(players.Players); i++ {
		fmt.Println("Player name:", players.Players[i].KickerName)
		fmt.Println("Player team:", players.Players[i].KickerTeam)
	}
}
