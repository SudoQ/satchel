package external

import (
	"log"
	"net/http"
	"time"
	"io/ioutil"
)

type Ext struct {
	url       string
	timestamp time.Time
	period    time.Duration
	init bool
	dataCh    chan []byte
	quitCh		chan bool
}

func New(url string, period time.Duration) *Ext {
	return &Ext{
		url:       url,
		timestamp: time.Time{},
		init: true,
		period:    period,
		dataCh:    make(chan []byte, 0), // Buffer?
		quitCh: make(chan bool, 0),
	}
}

func (this *Ext) DataChannel() <-chan []byte {
	return this.dataCh
}

func (this *Ext) Close() {
	this.quitCh <- true
	close(this.quitCh)
	close(this.dataCh)
}

func (this *Ext) Collect() {
	for {
		now := time.Now()
		timeDelta := now.Sub(this.timestamp)
		timeCondition := (this.timestamp.Before(now) && timeDelta > this.period)
		if timeCondition || this.init {
			this.init = false
			log.Printf("GET %s\n", this.url)

			resp, err := http.Get(this.url)
			if err != nil {
				log.Println(err)
				return
			}

			payload, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Println(err)
				return
			}
			this.dataCh <- payload
		}
		select {
		case <-time.After(this.period):
			continue
		case <-this.quitCh:
			return
		}
	}
}
