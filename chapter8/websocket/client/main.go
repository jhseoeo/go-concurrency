package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/jhseoeo/go-concurrency/chapter8/websocket/coder"
	"github.com/jhseoeo/go-concurrency/chapter8/websocket/message"
	"golang.org/x/net/websocket"
	"os"
)

func main() {
	cli, err := websocket.Dial("ws://localhost:8080/chat", "", "http://localhost:8080")
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	decoder := json.NewDecoder(cli)
	rcvCh, rcvErrCh := coder.DecodeToChan(func(msg *message.Message) error {
		return decoder.Decode(&msg)
	})
	sendCh := make(chan message.Message)
	sendErrCh := coder.EncodeFromChan(sendCh, func(msg message.Message) ([]byte, error) {
		return json.Marshal(msg)
	}, cli.Write)

	done := make(chan struct{})
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			text := scanner.Text()
			select {
			case <-done:
				return
			default:
			}
			sendCh <- message.Message{
				Message: text,
			}
		}
	}()

	for {
		select {
		case msg, ok := <-rcvCh:
			if !ok {
				close(done)
				return
			}
			fmt.Println(msg.From, msg.Timestamp.Format("15:04:05"), msg.Message)
		case <-rcvErrCh:
			return
		case <-sendErrCh:
			return
		}
	}
}
