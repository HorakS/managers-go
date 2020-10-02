package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
)

type Match struct {
	HomeTeam      string `json:"homeTeam"`
	GuestTeam     string `json:"guestTeam"`
	EndScore      string `json:"endScore"`
	HalftimeScore string `json:"halftimeScore"`
}

type Pdata struct {
	Match    Match   `json:"match"`
	Grade    float64 `json:"grade"`
	Scp      int     `json:"scp"`
	Playtime int     `json:"playtime"`
	SubIn    int     `json:"subIn"`
	SubOut   int     `json:"subOut"`
}

type Player struct {
	Name       string        `json:"name"`
	Team       string        `json:"team"`
	Position   string        `json:"position"`
	KickerName string        `json:"kickerName"`
	KickerTeam string        `json:"kickerTeam"`
	Data       map[int]Pdata `json:"data"`
}

func (m Match) String() string {
	return fmt.Sprintf("%v %v (%v) %v", m.HomeTeam, m.EndScore, m.HalftimeScore, m.GuestTeam)
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

			matchday, data := parsePdata(row)
			fmt.Println(matchday, data.Match)
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

func parsePdata(row *colly.HTMLElement) (matchDay int, pData *Pdata) {
	data := new(Pdata)
	match := new(Match)
	matchInfo := row.ChildText("div.kick__vita__statistic--table-second_dateinfo")
	matchDay, _ = strconv.Atoi(strings.Split(matchInfo, ".")[0])
	teams := row.ChildTexts("div.kick__v100-gameCell__team__name")
	// TODO: Instead extract kicker team names from hrefs?
	match.HomeTeam = teams[0]
	match.GuestTeam = teams[1]

	scores := row.ChildTexts("div.kick__v100-scoreBoard__scoreHolder__score")
	match.EndScore = scores[0] + ":" + scores[1]
	match.HalftimeScore = scores[2] + ":" + scores[3]

	var err error
	data.Grade, _ = strconv.ParseFloat(strings.ReplaceAll(row.ChildText("td:nth-child(2)"), ",", "."), 64)
	data.Scp, _ = strconv.Atoi(row.ChildText("td:nth-child(6)"))
	data.SubIn, _ = strconv.Atoi(row.ChildText("td:nth-child(7"))
	data.SubOut, err = strconv.Atoi(row.ChildText("td:nth-child(8)"))
	if err != nil {
		data.SubOut = 90
	}
	data.Playtime = data.SubOut - data.SubIn
	data.Match = *match

	return matchDay, data
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
