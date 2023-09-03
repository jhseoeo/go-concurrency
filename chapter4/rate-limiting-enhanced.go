package main

import (
	"sync"
	"time"
)

type LimiterEnhanced struct {
	mu         sync.Mutex
	rate       int
	bucketSize int
	nTokens    int
	lastToken  time.Time
}

func NewLimiterEnhanced(rate int, bucketSize int) *LimiterEnhanced {
	return &LimiterEnhanced{
		rate:       rate,
		bucketSize: bucketSize,
		nTokens:    bucketSize,
		lastToken:  time.Now(),
	}
}

func (l *LimiterEnhanced) Wait() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.nTokens > 0 {
		l.nTokens--
		return
	}

	timeElapsed := time.Since(l.lastToken)
	period := time.Duration(int(time.Second) / l.rate)
	nTokens := int(timeElapsed / period)
	l.nTokens = nTokens
	l.lastToken = l.lastToken.Add(time.Duration(nTokens) * period)
	if l.nTokens > l.bucketSize {
		l.nTokens = l.bucketSize
	}
	if l.nTokens > 0 {
		l.nTokens--
		return
	}

	next := l.lastToken.Add(period)
	wait := next.Sub(time.Now())
	if wait > 0 {
		time.Sleep(wait)
	}
	l.lastToken = next
}
