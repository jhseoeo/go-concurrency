package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"runtime/pprof"
	"time"
)

type Request struct{}
type Response struct{}

var ErrBusy = errors.New("server is busy")
var ErrTimeout = errors.New("server is timeout")

type Monitor[Req, Res any] struct {
	CallTimeout  time.Duration
	AlertTimeout time.Duration
	Alert        chan struct{}
	SlowFunc     func(Req) (Res, error)
	Done         chan struct{}
	active       chan struct{}
	full         chan struct{}
	heartBeat    chan struct{}
}

func (mon Monitor[Req, Res]) Call(ctx context.Context, req Req) (Res, error) {
	var res Res
	var err error

	select {
	case mon.active <- struct{}{}:
	default:
		select {
		case mon.active <- struct{}{}:
		case mon.full <- struct{}{}:
			return res, ErrBusy
		default:
			return res, ErrBusy
		}
	}

	complete := make(chan struct{})
	go func() {
		defer func() {
			<-mon.active
			select {
			case mon.heartBeat <- struct{}{}:
			default:
			}
			close(complete)
		}()
		res, err = mon.SlowFunc(req)
	}()

	select {
	case <-time.After(mon.CallTimeout):
		return res, ErrTimeout
	case <-complete:
		return res, err
	}
}

func NewMonitor[Req, Res any](callTimeout time.Duration, alertTimeout time.Duration, maxActive int, slowFunc func(Req) (Res, error)) Monitor[Req, Res] {
	mon := Monitor[Req, Res]{
		CallTimeout:  callTimeout,
		AlertTimeout: alertTimeout,
		SlowFunc:     slowFunc,
		Alert:        make(chan struct{}, 1),
		Done:         make(chan struct{}),
		active:       make(chan struct{}, maxActive),
		full:         make(chan struct{}),
		heartBeat:    make(chan struct{}),
	}

	go func() {
		var timer *time.Timer
		for {
			if timer == nil {
				select {
				case <-mon.full:
					timer = time.NewTimer(mon.AlertTimeout)
				case <-mon.Done:
					return
				}
			} else {
				select {
				case <-timer.C:
					mon.Alert <- struct{}{}
				case <-mon.heartBeat:
					if !timer.Stop() {
						<-timer.C
					}
				case <-mon.Done:
					return
				}
				timer = nil
			}
		}
	}()

	return mon
}

func SlowFunc(req *Request) (*Response, error) {
	k := rand.Intn(100)
	if k == 98 {
		fmt.Println("This call will be timeout")
		select {}
	}
	if k > 85 {
		time.Sleep(10 * time.Second)
	}
	time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
	return &Response{}, nil
}

func main() {
	mon := NewMonitor(50*time.Millisecond, 100*time.Second, 10, SlowFunc)
	go func() {
		select {
		case <-mon.Alert:
			pprof.Lookup("goroutine").WriteTo(os.Stderr, 1)
		case <-mon.Done:
			return
		}
	}()

	for i := 0; i < 5; i++ {
		go func() {
			for {
				_, err := mon.Call(context.Background(), &Request{})
				if err != nil {
					fmt.Println(len(mon.active), err)
				}
			}
		}()
	}
	select {}
}
