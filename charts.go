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

func scpHeat(players []Player) *charts.HeatMap {
	matchdays := make([]int, 34)
	for i := range matchdays {
		matchdays[i] = 1 + i
	}

	names := make([]string, len(players))
	data := make([][3]interface{}, 0)
	maxScp := 0
	for i, p := range players {
		names[i] = p.Name
		for j, m := range p.Matches {
			data = append(data, [3]interface{}{int(j) - 1, i, m.Scp})
			if m.Scp > maxScp {
				maxScp = m.Scp
			}
		}
	}
	hm := charts.NewHeatMap()
	hm.SetGlobalOptions(charts.TitleOpts{Title: "SCP"}, charts.InitOpts{Height: "800px"})

	hm.AddXAxis(matchdays).AddYAxis("SCP", data)
	hm.SetGlobalOptions(
		charts.YAxisOpts{Data: names, Type: "category", SplitArea: charts.SplitAreaOpts{Show: true}},
		charts.XAxisOpts{Type: "category", SplitArea: charts.SplitAreaOpts{Show: true}},
		charts.VisualMapOpts{Calculable: true, Max: float32(maxScp), Min: 0, InRange: charts.VMInRange{Color: []string{"#d6454e", "#f7eb83", "#34bf5b"}}},
	)

	return hm
}

func gradeHeat(players []Player) *charts.HeatMap {
	matchdays := make([]int, 34)
	for i := range matchdays {
		matchdays[i] = 1 + i
	}

	names := make([]string, len(players))
	data := make([][3]interface{}, 0)
	for i, p := range players {
		names[i] = p.Name
		for j, m := range p.Matches {
			if m.Grade == 0.0 {
				data = append(data, [3]interface{}{int(j) - 1, i, "-"})
			} else {
				data = append(data, [3]interface{}{int(j) - 1, i, m.Grade})
			}
		}
	}
	hm := charts.NewHeatMap()
	hm.SetGlobalOptions(charts.TitleOpts{Title: "Grade"}, charts.InitOpts{Height: "800px"})

	hm.AddXAxis(matchdays).AddYAxis("Grade", data)
	hm.SetGlobalOptions(
		charts.YAxisOpts{Data: names, Type: "category", SplitArea: charts.SplitAreaOpts{Show: true}},
		charts.XAxisOpts{Type: "category", SplitArea: charts.SplitAreaOpts{Show: true}},
		charts.VisualMapOpts{Calculable: true, Max: 6.0, Min: 1.0, InRange: charts.VMInRange{Color: []string{"#34bf5b", "#f7eb83", "#d6454e"}}},
	)

	return hm
}

func top11Heat(players []Player) *charts.HeatMap {
	matchdays := make([]int, 34)
	for i := range matchdays {
		matchdays[i] = 1 + i
	}

	names := make([]string, len(players))
	data := make([][3]interface{}, 0)
	for i, p := range players {
		names[i] = p.Name
		for j, m := range p.Matches {
			if m.Top11 == true {
				data = append(data, [3]interface{}{int(j) - 1, i, 1})
			} else {
				data = append(data, [3]interface{}{int(j) - 1, i, 0})
			}
		}
	}
	hm := charts.NewHeatMap()
	hm.SetGlobalOptions(charts.TitleOpts{Title: "Top11"}, charts.InitOpts{Height: "800px"})

	hm.AddXAxis(matchdays).AddYAxis("Top11", data)
	hm.SetGlobalOptions(
		charts.YAxisOpts{Data: names, Type: "category", SplitArea: charts.SplitAreaOpts{Show: true}},
		charts.XAxisOpts{Type: "category", SplitArea: charts.SplitAreaOpts{Show: true}},
		charts.VisualMapOpts{Calculable: true, Max: 1, Min: 0, InRange: charts.VMInRange{Color: []string{"#d6454e", "#f7eb83", "#34bf5b"}}},
	)

	return hm
}

func playtimeHeat(players []Player) *charts.HeatMap {
	matchdays := make([]int, 34)
	for i := range matchdays {
		matchdays[i] = 1 + i
	}

	names := make([]string, len(players))
	data := make([][3]interface{}, 0)
	for i, p := range players {
		names[i] = p.Name
		for j, m := range p.Matches {
			data = append(data, [3]interface{}{int(j) - 1, i, m.Playtime})
		}
	}
	hm := charts.NewHeatMap()
	hm.SetGlobalOptions(charts.TitleOpts{Title: "Playtime"}, charts.InitOpts{Height: "800px"})

	hm.AddXAxis(matchdays).AddYAxis("Playtime", data)
	hm.SetGlobalOptions(
		charts.YAxisOpts{Data: names, Type: "category", SplitArea: charts.SplitAreaOpts{Show: true}},
		charts.XAxisOpts{Type: "category", SplitArea: charts.SplitAreaOpts{Show: true}},
		charts.VisualMapOpts{Calculable: true, Max: 90, Min: 0, InRange: charts.VMInRange{Color: []string{"#d6454e", "#f7eb83", "#34bf5b"}}},
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

	page.Add(
		scpHeat(players),
		playtimeHeat(players),
		gradeHeat(players),
		top11Heat(players),
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
