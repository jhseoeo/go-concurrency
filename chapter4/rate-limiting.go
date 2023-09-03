package main

import (
	"net/http"
	"time"
)

type RateLimiter interface {
	Wait()
}

type ChannelRate struct {
	bucket chan struct{}
	ticker *time.Ticker
	done   chan struct{}
}

func NewChannelRate(rate float64, limit int) *ChannelRate {
	r := &ChannelRate{
		bucket: make(chan struct{}, limit),
		ticker: time.NewTicker(time.Second / time.Duration(rate)),
		done:   make(chan struct{}),
	}

	for i := 0; i < limit; i++ {
		r.bucket <- struct{}{}
	}

	go func() {
		for {
			select {
			case <-r.done:
				return
			case <-r.ticker.C:
				select {
				case r.bucket <- struct{}{}:
				default:
				}
			}
		}
	}()

	return r
}

func (r *ChannelRate) Wait() {
	<-r.bucket
}

func (r *ChannelRate) Close() {
	close(r.done)
	r.ticker.Stop()
}

func handle(w http.ResponseWriter, r *http.Request) {
	//limiter.Wait()
	// do something
}
