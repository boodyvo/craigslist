package scrapper

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gocolly/colly/v2"
)

const (
	SearchURL = "https://sfbay.craigslist.org/search/%s/%s"
	JobURL    = "https://sfbay.craigslist.org/%s/cto/%d.html"

	WorkerNumber = 40
	JobsNumber   = 3 * WorkerNumber

	NotificationBufferSize = 10
	MarkerSize             = 100

	TickerNumber         = 4
	RestartTickerTimeout = 2 * time.Minute
)

var (
	extraSearchURLs = []string{
		"https://sfbay.craigslist.org/search/sss?condition=10&condition=20&condition=30&condition=40&condition=50&condition=60",
		"https://sfbay.craigslist.org/search/sss?condition=10&condition=20&condition=30&condition=40&condition=50&condition=60&auto_fuel_type=1&auto_fuel_type=2&auto_fuel_type=3&auto_fuel_type=4&auto_fuel_type=6",
		"https://sfbay.craigslist.org/search/sby/sss?condition=10&condition=20&condition=30&condition=40&condition=50&condition=60&auto_size=1&auto_size=2&auto_size=3&auto_size=4",
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
		"https://newyork.craigslist.org/search/sss?condition=10&condition=20&condition=30&condition=40&condition=50&condition=60",
	}
	searchCategories = []string{"sss", "cta", "cto"}
	areas            = []string{"eby", "nby", "pen", "sfc", "scz", "sby"}
)

type Scrapper interface {
	Start() error
	Stop() error
	GetLastIndex() (int64, error)
	SubscriptionChan() <-chan string
}

type ticker struct {
	breaker  chan struct{}
	isClosed uint32
	Last     int64
}

func (t *ticker) Stop() {
	log.Printf("closed ticker: %d", t.Last)
	if atomic.CompareAndSwapUint32(&t.isClosed, 0, 1) {
		close(t.breaker)
	}
}

type httpScrapper struct {
	quit          chan struct{}
	notifications chan string
	jobIndexes    chan int64
	workerWG      sync.WaitGroup
	subscriptions []chan string

	lastFoundIndexes map[string]int64
	lastMarker       int64

	lastIndexFound int64

	tickers      []*ticker
	restartTimer *time.Timer

	sync.RWMutex
}

func New() Scrapper {
	return &httpScrapper{
		quit:             make(chan struct{}),
		notifications:    make(chan string, NotificationBufferSize),
		jobIndexes:       make(chan int64, JobsNumber),
		tickers:          make([]*ticker, 0, TickerNumber),
		subscriptions:    make([]chan string, 0),
		lastFoundIndexes: make(map[string]int64, 2*MarkerSize),
	}
}

func (s *httpScrapper) Start() error {
	index, err := s.GetLastIndex()
	if err != nil {
		return err
	}

	s.lastIndexFound = index

	s.workerWG = sync.WaitGroup{}
	for i := 0; i < WorkerNumber; i++ {
		s.workerWG.Add(1)
		go s.startWorker(fmt.Sprintf("%d", i), s.jobIndexes)
	}

	go s.watchNotifications()

	s.addTicker()

	<-s.quit
	close(s.jobIndexes)
	s.workerWG.Wait()

	return nil
}

func (s *httpScrapper) watchNotifications() {
	log.Printf("start watcher")
	for {
		select {
		case url := <-s.notifications:
			s.Lock()
			// check duplicates
			if _, ok := s.lastFoundIndexes[url]; ok {
				s.Unlock()
				continue
			}
			// fill duplicates array and fresh it
			s.lastMarker++
			s.lastFoundIndexes[url] = s.lastMarker
			freshIndexes := make(map[string]int64, len(s.lastFoundIndexes))
			for k, v := range s.lastFoundIndexes {
				if v > s.lastMarker-MarkerSize {
					freshIndexes[k] = v
				}
			}
			s.lastFoundIndexes = freshIndexes

			for _, subscriber := range s.subscriptions {
				log.Printf("will push message to subscriber: %s", url)
				subscriber <- url
				log.Printf("pushed message to subscriber: %s", url)
			}
			s.Unlock()
		case <-s.quit:
			log.Printf("stop watcher")

			return
		}
	}
}

func (s *httpScrapper) Stop() error {
	close(s.quit)

	return nil
}

func (s *httpScrapper) SubscriptionChan() <-chan string {
	subscriber := make(chan string, NotificationBufferSize)
	s.subscriptions = append(s.subscriptions, subscriber)

	return subscriber
}

func (s *httpScrapper) GetLastIndex() (int64, error) {
	var index int64
	c := colly.NewCollector()

	c.OnHTML("body", func(e *colly.HTMLElement) {
		fmt.Println("got some html")
	})

	c.OnHTML(".result-row", func(e *colly.HTMLElement) {
		fmt.Println("found something on the page")
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

			err := c.Visit(url)
			if err != nil {
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

func (s *httpScrapper) resetLastIndex() {
	log.Printf("reseting last index from: %d", s.lastIndexFound)
	lastIndex, err := s.GetLastIndex()
	if err != nil {
		log.Printf("error while restarting ticker: %s", err)

		return
	}
	s.Lock()
	if lastIndex > s.lastIndexFound {
		s.lastIndexFound = lastIndex
	}
	s.Unlock()
	log.Printf("new last: %d", s.lastIndexFound)
}

func (s *httpScrapper) startTicker() *ticker {
	log.Println("starting ticker", len(s.tickers))
	s.RLock()
	t := &ticker{
		breaker:  make(chan struct{}),
		Last:     s.lastIndexFound,
		isClosed: 0,
	}
	s.RUnlock()

	go func() {
		for {
			select {
			case <-t.breaker:
				return
			default:
				t.Last++
				s.jobIndexes <- t.Last
			}
		}
	}()

	return t
}

func (s *httpScrapper) stopAllTickers() {
	s.Lock()
	defer s.Unlock()

	for _, t := range s.tickers {
		t.Stop()
	}
	s.tickers = s.tickers[:0]
}

func (s *httpScrapper) addTicker() {
	ticker := s.startTicker()
	s.Lock()
	defer s.Unlock()
	s.tickers = append(s.tickers, ticker)

	if s.restartTimer != nil {
		s.restartTimer.Stop()
	}
	s.restartTimer = time.AfterFunc(RestartTickerTimeout, s.restartTicker)
	log.Printf("timer was restarted, should trigger at %v", time.Now().Add(RestartTickerTimeout))
}

func (s *httpScrapper) restartTicker() {
	log.Printf("restarting tickers")
	s.resetLastIndex()

	s.Lock()
	if len(s.tickers) == TickerNumber {
		log.Printf("checkging tickers")
		last := int64(0)
		lastIndex := -1

		for i, t := range s.tickers {
			// remove all tickers that are old
			if t.Last < s.lastIndexFound {
				lastIndex = i
				break
			}

			if t.Last > last {
				lastIndex = i
				last = t.Last
			}
		}

		if lastIndex != -1 && s.tickers[lastIndex] != nil {
			s.tickers[lastIndex].Stop()
			s.tickers[len(s.tickers)-1], s.tickers[lastIndex] = s.tickers[lastIndex], s.tickers[len(s.tickers)-1]
			s.tickers = s.tickers[:len(s.tickers)-1]
		}
		log.Printf("checked successfully")
	}
	s.Unlock()

	log.Printf("will add tickers")
	s.addTicker()
	log.Printf("finished restarting successfully")
}

func (s *httpScrapper) resetTickers() {
	log.Printf("resetting tickers")
	s.resetLastIndex()
	s.stopAllTickers()
	s.addTicker()
}

func (s *httpScrapper) startWorker(id string, jobIndexes <-chan int64) {
	defer func() {
		log.Printf("finish worker %s", id)
		s.workerWG.Done()
	}()

	//log.Printf("start worker %s", id)
	for index := range jobIndexes {
		//log.Printf("worker %s: got job: %d", id, index)
		for _, area := range areas {
			url := fmt.Sprintf(JobURL, area, index)
			resp, err := http.Head(url)
			if err != nil {
				log.Printf("unexpected error for head request for index %d: %s", index, err.Error())
				time.Sleep(time.Second)

				continue
			}
			//log.Println(resp.StatusCode)
			if resp.StatusCode != 404 && !isSuccessStatus(resp.StatusCode) {
				log.Printf("found unknown response status for index %d: %s", index, resp.Status)
			}

			if isSuccessStatus(resp.StatusCode) {
				log.Printf("worker %s: found url: %s", id, url)
				s.notifications <- url
				atomic.StoreInt64(&s.lastIndexFound, index)
				s.resetTickers()
				log.Printf("worker %s: ok", id)

				break
			}
		}
		//log.Printf("worker %s: finish job: %d", id, index)
	}
}

func isSuccessStatus(status int) bool {
	return status >= 200 && status < 300
}
