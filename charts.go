package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-echarts/go-echarts/charts"
)

func goalsHeat(matchdays []int, players []Player) *charts.HeatMap {
	names := make([]string, len(players))
	data := make([][3]interface{}, 0)
	maxGoals := 0.0
	for i, p := range players {
		names[i] = fmt.Sprintf("%v (%v)", p.Name, p.Team)
		for j, m := range p.Matches {
			if m.ConcededGoals != 99 {
				data = append(data, [3]interface{}{int(j) - 1, i, m.ConcededGoals})
				if m.ConcededGoals > maxGoals {
					maxGoals = m.ConcededGoals
				}
			}
		}
	}

	hm := charts.NewHeatMap()
	hm.SetGlobalOptions(charts.TitleOpts{Title: "Gegentore"}, charts.InitOpts{Height: "800px", Width: "1200px"})

	hm.AddXAxis(matchdays).AddYAxis("Gegentore", data)
	hm.SetGlobalOptions(
		charts.YAxisOpts{Data: names, Type: "category", SplitArea: charts.SplitAreaOpts{Show: true}},
		charts.XAxisOpts{Type: "category", SplitArea: charts.SplitAreaOpts{Show: true}},
		charts.VisualMapOpts{Calculable: true, Max: float32(maxGoals), Min: 0.0, InRange: charts.VMInRange{Color: []string{"#40b860", "#f7eb83", "#d6454e"}}},
	)

	return hm
}

func scpHeat(matchdays []int, players []Player) *charts.HeatMap {
	names := make([]string, len(players))
	data := make([][3]interface{}, 0)
	maxScp := 0
	for i, p := range players {
		names[i] = fmt.Sprintf("%v (%v)", p.Name, p.Team)
		for j, m := range p.Matches {
			data = append(data, [3]interface{}{int(j) - 1, i, m.Scp})
			if m.Scp > maxScp {
				maxScp = m.Scp
			}
		}
	}
	hm := charts.NewHeatMap()
	hm.SetGlobalOptions(charts.TitleOpts{Title: "SCP"}, charts.InitOpts{Height: "800px", Width: "1200px"})

	hm.AddXAxis(matchdays).AddYAxis("SCP", data)
	hm.SetGlobalOptions(
		charts.YAxisOpts{Data: names, Type: "category", SplitArea: charts.SplitAreaOpts{Show: true}},
		charts.XAxisOpts{Type: "category", SplitArea: charts.SplitAreaOpts{Show: true}},
		charts.VisualMapOpts{Calculable: true, Max: float32(maxScp), Min: 0, InRange: charts.VMInRange{Color: []string{"#d6454e", "#f7eb83", "#40b860"}}},
	)

	return hm
}

func gradeHeat(matchdays []int, players []Player) *charts.HeatMap {
	names := make([]string, len(players))
	data := make([][3]interface{}, 0)
	for i, p := range players {
		names[i] = fmt.Sprintf("%v (%v)", p.Name, p.Team)
		for j, m := range p.Matches {
			if m.Grade == 0.0 {
				data = append(data, [3]interface{}{int(j) - 1, i, "-"})
			} else {
				data = append(data, [3]interface{}{int(j) - 1, i, m.Grade})
			}
		}
	}
	hm := charts.NewHeatMap()
	hm.SetGlobalOptions(charts.TitleOpts{Title: "Note"}, charts.InitOpts{Height: "800px", Width: "1200px"})

	hm.AddXAxis(matchdays).AddYAxis("Note", data)
	hm.SetGlobalOptions(
		charts.YAxisOpts{Data: names, Type: "category", SplitArea: charts.SplitAreaOpts{Show: true}},
		charts.XAxisOpts{Type: "category", SplitArea: charts.SplitAreaOpts{Show: true}},
		charts.VisualMapOpts{Calculable: true, Max: 6.0, Min: 1.0, InRange: charts.VMInRange{Color: []string{"#40b860", "#f7eb83", "#d6454e"}}},
	)

	return hm
}

func top11Heat(matchdays []int, players []Player) *charts.HeatMap {
	names := make([]string, len(players))
	data := make([][3]interface{}, 0)
	for i, p := range players {
		names[i] = fmt.Sprintf("%v (%v)", p.Name, p.Team)
		for j, m := range p.Matches {
			if m.Top11 == true {
				data = append(data, [3]interface{}{int(j) - 1, i, 1})
			} else {
				data = append(data, [3]interface{}{int(j) - 1, i, 0})
			}
		}
	}
	hm := charts.NewHeatMap()
	hm.SetGlobalOptions(charts.TitleOpts{Title: "11 des Tages"}, charts.InitOpts{Height: "800px", Width: "1200px"})

	hm.AddXAxis(matchdays).AddYAxis("11 des Tages", data)
	hm.SetGlobalOptions(
		charts.YAxisOpts{Data: names, Type: "category", SplitArea: charts.SplitAreaOpts{Show: true}},
		charts.XAxisOpts{Type: "category", SplitArea: charts.SplitAreaOpts{Show: true}},
		charts.VisualMapOpts{Calculable: true, Max: 1, Min: 0, InRange: charts.VMInRange{Color: []string{"#d6454e", "#f7eb83", "#40b860"}}},
	)

	return hm
}

func playtimeHeat(matchdays []int, players []Player) *charts.HeatMap {
	names := make([]string, len(players))
	data := make([][3]interface{}, 0)
	for i, p := range players {
		names[i] = fmt.Sprintf("%v (%v)", p.Name, p.Team)
		for j, m := range p.Matches {
			data = append(data, [3]interface{}{int(j) - 1, i, m.Playtime})
		}
	}
	hm := charts.NewHeatMap()
	hm.SetGlobalOptions(charts.TitleOpts{Title: "Spielzeit"}, charts.InitOpts{Height: "800px", Width: "1200px"})

	hm.AddXAxis(matchdays).AddYAxis("Spielzeit", data)
	hm.SetGlobalOptions(
		charts.YAxisOpts{Data: names, Type: "category", SplitArea: charts.SplitAreaOpts{Show: true}},
		charts.XAxisOpts{Type: "category", SplitArea: charts.SplitAreaOpts{Show: true}},
		charts.VisualMapOpts{Calculable: true, Max: 90, Min: 0, InRange: charts.VMInRange{Color: []string{"#d6454e", "#f7eb83", "#40b860"}}},
	)

	return hm
}

func handler(w http.ResponseWriter, _ *http.Request) {
	host := "http://127.0.0.1:8000"
	rs := make([]charts.RouterOpts, 0)
	rs = append(rs, charts.RouterOpts{URL: host + "/managers", Text: "Managers"})
	page := charts.NewPage(rs...)

	var players []Player
	jsonFile, err := os.Open("playerdata.json")
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &players)

	matchdays := make([]int, 34)
	for i := range matchdays {
		matchdays[i] = 1 + i
	}

	page.Add(
		scpHeat(matchdays, players),
		playtimeHeat(matchdays, players),
		gradeHeat(matchdays, players),
		goalsHeat(matchdays, players),
		top11Heat(matchdays, players),
	)

	f, err := os.Create("managers.html")
	if err != nil {
		fmt.Println(err)
	}
	page.Render(w, f)
}

func serveCharts() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	server := &http.Server{Addr: ":8000"}
	http.HandleFunc("/", handler)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			fmt.Println(err)
		}
	}()

	fmt.Println("Server started...")

	<-stop
	fmt.Println("Stopping server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	if err := server.Shutdown(ctx); err != nil {
		fmt.Println("Error during shutdown:", err)
	}
	fmt.Println("Done.")
}
