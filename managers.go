package main

import (
	"fmt"

	"github.com/gocolly/colly"
)

type Pdata struct {
	Name string
}

func getPdata() (pData Pdata, err error) {
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
	return pData, nil
}

func main() {
	getPdata()
}
