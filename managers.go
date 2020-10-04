package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
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
	Top11    bool    `json:"top11"`
}

type Player struct {
	Name       string         `json:"name"`
	Team       string         `json:"team"`
	Position   string         `json:"position"`
	KickerName string         `json:"kickerName"`
	KickerTeam string         `json:"kickerTeam"`
	Matches    map[int]*Pdata `json:"matches"`
	Average    Average        `json:"average"`
}

type Average struct {
	Grade    float64 `json:"grade"`
	Scp      int     `json:"scp"`
	Playtime int     `json:"playtime"`
	Top11    int     `json:"top11"`
}

func (m Match) String() string {
	return fmt.Sprintf("%v %v (%v) %v", m.HomeTeam, m.EndScore, m.HalftimeScore, m.GuestTeam)
}

func (d Pdata) String() string {
	return fmt.Sprintf("Grade: %v, Scp: %v, Playtime: %v, SubIn: %v, SubOut: %v, Top11: %v", d.Grade, d.Scp, d.Playtime, d.SubIn, d.SubOut, d.Top11)
}

func getPdata(players []Player) (err error) {
	c := colly.NewCollector(
		colly.AllowedDomains("www.kicker.de"),
		colly.Async(true),
	)

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.OnHTML("div.kick__vita__statistic", func(e *colly.HTMLElement) {
		player := e.Request.Ctx.GetAny("player").(*Player)
		player.Matches = make(map[int]*Pdata)

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

			noplay := row.ChildText("td.kick__vita__statistic--table-second_noplay") == "ohne Einsatz im Kader"
			if !bundesliga || match == "" || noplay {
				return true
			}

			if !strings.Contains(match, "Spieltag") {
				return false
			}

			matchday, data := parsePdata(row)
			player.Matches[matchday] = data
			return true
		})
	})

	for i := 0; i < len(players); i++ {
		ctx := colly.NewContext()
		url := "https://www.kicker.de/" + players[i].KickerName + "/spieler/bundesliga/2020-21/" + players[i].KickerTeam
		ctx.Put("player", &players[i])
		c.Request("GET", url, nil, ctx, nil)
	}

	c.Wait()

	for i := 0; i < len(players); i++ {
		avg, err := getAverageData(players[i].Matches)
		if err == nil {
			players[i].Average = *avg
		}
	}

	return nil
}

func getAverageData(data map[int]*Pdata) (avg *Average, err error) {
	matches := len(data)
	avg = new(Average)
	if matches == 0 {
		return avg, errors.New("No data")
	}

	gradedMatches := 0.0
	scp := 0
	playtime := 0
	grade := 0.0
	top11 := 0
	for _, match := range data {
		scp += match.Scp
		playtime += match.Playtime
		if match.Grade != 0 {
			grade += match.Grade
			gradedMatches += 1
		}
		if match.Top11 {
			top11 += 1
		}
	}
	if gradedMatches == 0.0 {
		avg.Grade = 0.0
	} else {
		avg.Grade = grade / gradedMatches
	}

	avg.Scp = scp / matches
	avg.Playtime = playtime / matches
	avg.Top11 = top11
	return avg, nil
}

func getTop11Data(players []Player) (err error) {
	top11s := make(map[string][]*int)

	c := colly.NewCollector(
		colly.AllowedDomains("www.kicker.de"),
		colly.Async(true),
	)

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.OnHTML("a.kick__lineup-player-card", func(e *colly.HTMLElement) {
		player := strings.Split(e.Attr("href"), "/")[1]
		matchday := e.Request.Ctx.GetAny("matchday").(int)
		top11s[player] = append(top11s[player], &matchday)
	})

	for i := 1; i < 35; i++ {
		url := "https://www.kicker.de/bundesliga/elf-des-tages/2020-21/" + strconv.Itoa(i)
		ctx := colly.NewContext()
		ctx.Put("matchday", i)
		c.Request("GET", url, nil, ctx, nil)
	}

	c.Wait()

	for _, player := range players {
		if val, ok := top11s[player.KickerName]; ok {
			for _, matchday := range val {
				player.Matches[*matchday].Top11 = true
			}
		}
	}

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
	data.SubIn, _ = strconv.Atoi(strings.ReplaceAll(row.ChildText("td:nth-child(7)"), ".", ""))
	data.SubOut, err = strconv.Atoi(strings.ReplaceAll(row.ChildText("td:nth-child(8)"), ".", ""))
	if err != nil {
		data.SubOut = 90
	}
	data.Playtime = data.SubOut - data.SubIn
	data.Match = *match
	data.Top11 = false // default Top11 is false, may get changed to true later

	return matchDay, data
}

func writeCsv(players []Player) (err error) {
	file, err := os.Create("playerdata.csv")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)

	rows := [][]string{{""}, {""}}
	subheaders := []string{"Note", "SCP", "Spielzeit", "11 des Tages"}
	summary := []string{"Summary"}
	for _, p := range players {
		rows[0] = append(rows[0], p.Name, p.Team, p.Position, "")
		rows[1] = append(rows[1], subheaders...)
		summary = append(summary, fmt.Sprintf("%.2f", p.Average.Grade), strconv.Itoa(p.Average.Scp), strconv.Itoa(p.Average.Playtime), strconv.Itoa(p.Average.Top11))
	}

	for i := 1; i < 35; i++ {
		row := []string{strconv.Itoa(i)}
		for _, p := range players {
			if m, ok := p.Matches[i]; ok {
				row = append(row, fmt.Sprintf("%.1f", m.Grade), strconv.Itoa(m.Scp), strconv.Itoa(m.Playtime), strconv.FormatBool(m.Top11))
			} else {
				row = append(row, "-", "-", "-", "-")
			}
		}
		rows = append(rows, row)
	}
	rows = append(rows, summary)

	writer.WriteAll(rows)
	writer.Flush()
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
	getTop11Data(players)

	jsonString, err := json.MarshalIndent(players, "", " ")
	if err != nil {
		fmt.Println(err)
	}
	ioutil.WriteFile("playerdata.json", jsonString, os.ModePerm)

	writeCsv(players)
}
