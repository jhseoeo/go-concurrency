package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type TCPServer struct {
	Listener    net.Listener
	HandlerFunc func(context.Context, net.Conn)
	wg          sync.WaitGroup
}

func (srv *TCPServer) Listen() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	fmt.Println("Listening on", srv.Listener.Addr().String())

	for {
		conn, err := srv.Listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			fmt.Println(err)
		}

		srv.wg.Add(1)
		go func() {
			defer srv.wg.Done()
			srv.HandlerFunc(ctx, conn)
		}()
	}
}

func (srv *TCPServer) StopListener() error {
	return srv.Listener.Close()
}

func (srv *TCPServer) WaitForConnections(timeout time.Duration) bool {
	toCh := time.After(timeout)
	doneCh := make(chan struct{})
	go func() {
		srv.wg.Wait()
		close(doneCh)
	}()

	select {
	case <-toCh:
		return false
	case <-doneCh:
		return true
	}
}

func NewTCPServer(handler func(ctx context.Context, conn net.Conn)) *TCPServer {
	var err error
	srv := &TCPServer{}
	srv.Listener, err = net.Listen("tcp", "")
	if err != nil {
		panic(err)
	}
	srv.HandlerFunc = handler

	return srv
}

func main() {
	srv := NewTCPServer(func(ctx context.Context, conn net.Conn) {
		defer conn.Close()
		io.Copy(conn, conn)
	})
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sigCh
		fmt.Println("Shutting down...")
		err := srv.StopListener()
		if err != nil {
			fmt.Println(err)
		}
		srv.WaitForConnections(5 * time.Second)
	}()

	srv.Listen()
}
