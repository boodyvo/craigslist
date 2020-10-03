package main

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
)

var (
	SearchURL = "https://sfbay.craigslist.org/search/%s/%s"

	extraSearchURLs = []string{
		//"https://sfbay.craigslist.org/search/sss?condition=10&condition=20&condition=30&condition=40&condition=50&condition=60",
		//"https://sfbay.craigslist.org/search/sss?condition=10&condition=20&condition=30&condition=40&condition=50&condition=60&auto_fuel_type=1&auto_fuel_type=2&auto_fuel_type=3&auto_fuel_type=4&auto_fuel_type=6",
		//"https://sfbay.craigslist.org/search/sby/sss?condition=10&condition=20&condition=30&condition=40&condition=50&condition=60&auto_size=1&auto_size=2&auto_size=3&auto_size=4",
		"https://sfbay.craigslist.org/search/sby/sss?max_price=10000000",
		"https://sfbay.craigslist.org/search/sby/sss?max_price=1000",
		"https://sfbay.craigslist.org/search/sby/sss?max_price=100",
		"https://sfbay.craigslist.org/search/sby/sss?min_price=1000",
		"https://sfbay.craigslist.org/search/sby/sss?min_price=1",
		"https://sfbay.craigslist.org/search/sby/sss?min_price=10",
		"https://newyork.craigslist.org/search/sss?max_price=1000",
		"https://newyork.craigslist.org/search/sss?max_price=10",
		"https://newyork.craigslist.org/search/sss?max_price=100000",
		"https://newyork.craigslist.org/search/sss?min_price=1",
		"https://newyork.craigslist.org/search/sss?min_price=10",
		"https://newyork.craigslist.org/search/sss?min_price=100",
		//"https://newyork.craigslist.org/search/sss?condition=10&condition=20&condition=30&condition=40&condition=50&condition=60",
	}

	proxyURL = "http://165.225.84.143:8800"

	searchCategories = []string{"sss", "cta", "cto"}
	areas            = []string{"eby", "nby", "pen", "sfc", "scz", "sby"}
)

func main() {
	GetLastIndex()
	return
	// Instantiate default collector
	c := colly.NewCollector(colly.AllowURLRevisit())
	c.CheckHead = true

	// Rotate two socks5 proxies
	//rp, err := proxy.RoundRobinProxySwitcher(
	//	//"https://1.10.229.25:8080",
	//	"http://37.48.118.90:13042",
	//)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//c.SetProxy("http://37.48.118.90:13042")
	//c.SetProxy("socks5://95.174.67.50:18080")
	c.SetProxy(proxyURL)
	extensions.RandomUserAgent(c)

	// Print the response
	c.OnResponse(func(r *colly.Response) {
		fmt.Println("ok")
		fmt.Printf("%s\n", bytes.Replace(r.Body, []byte("\n"), nil, -1))
	})

	c.OnHTML("body", func(e *colly.HTMLElement) {
		fmt.Println("got some html")
	})

	// Fetch httpbin.org/ip five times
	for i := 0; i < 5; i++ {
		fmt.Println("hi")
		err := c.Visit("https://httpbin.org/ip")
		fmt.Println(err)
	}

	time.Sleep(10 * time.Second)
}

func GetLastIndex() (int64, error) {
	var index int64
	c := colly.NewCollector()
	if err := c.SetProxy(proxyURL); err != nil {
		return 0, err
	}
	extensions.RandomUserAgent(c)

	c.OnHTML("body", func(e *colly.HTMLElement) {
		//fmt.Println("got some html")
	})

	c.OnHTML(".result-row", func(e *colly.HTMLElement) {
		//fmt.Println("found something on the page")
		curIndexStr := e.Attr("data-pid")
		curIndex, err := strconv.ParseInt(curIndexStr, 10, 64)
		fmt.Println("curIndex", curIndex, index, err)
		if err != nil {
			return
		}

		if curIndex > index {
			index = curIndex
		}
	})

	wg := &sync.WaitGroup{}
	for _, area := range areas {
		for _, category := range searchCategories {
			wg.Add(1)
			url := fmt.Sprintf(SearchURL, area, category)

			go func(wg *sync.WaitGroup, url string) {
				defer wg.Done()

				err := c.Visit(url)
				if err != nil {
					return
				}
			}(wg, url)
		}
	}

	for _, url := range extraSearchURLs {
		wg.Add(1)

		go func(wg *sync.WaitGroup, url string) {
			defer wg.Done()

			if err := c.Visit(url); err != nil {
				return
			}
		}(wg, url)
	}
	wg.Wait()

	if index == 0 {
		return 0, errors.New("not found")
	}

	return index, nil
}
