package server

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/jhseoeo/go-concurrency/chapter8/websocket/coder"
	"github.com/jhseoeo/go-concurrency/chapter8/websocket/message"
)

func main() {
	dispatch := make(chan message.Message)
	connectCh := make(chan chan message.Message)
	disconnectCh := make(chan chan message.Message)
	go func() {
		clients := make(map[chan message.Message]struct{})
		for {
			select {
			case c := <-connectCh:
				clients[c] = struct{}{}
			case c := <-disconnectCh:
				delete(clients, c)
			case msg := <-dispatch:
				for c := range clients {
					select {
					case c <- msg:
					default:
						close(c)
					}
				}
			}
		}
	}()

	app := fiber.New()
	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		client := c.RemoteAddr().String()
		inputCh := make(chan message.Message, 10)
		connectCh <- inputCh
		defer func() {
			disconnectCh <- inputCh
		}()

		decoder := json.NewDecoder(c.NetConn())
		data, decodeErrCh := coder.DecodeToChan(func(msg *message.Message) error {
			err := decoder.Decode(msg)
			msg.From = client
			return err
		})
		encodeErrCh := coder.EncodeFromChan(inputCh, func(msg message.Message) ([]byte, error) {
			return json.Marshal(msg)
		}, c.NetConn().Write)

		for {
			select {
			case msg, ok := <-data:
				if !ok {
					return
				} else {
					dispatch <- msg
				}
			case <-decodeErrCh:
				return
			case <-encodeErrCh:
				return
			}
		}
	}))

	fmt.Println("Server is running at :8080")
	err := app.Listen(":8080")
	if err != nil {
		panic(err)
	}
}
