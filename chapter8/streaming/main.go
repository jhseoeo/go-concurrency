package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/jhseoeo/go-concurrency/chapter8/streaming/coder"
	"github.com/jhseoeo/go-concurrency/chapter8/streaming/filters"
	"github.com/jhseoeo/go-concurrency/chapter8/streaming/store"
	"math/rand"
	"net/http"
	"time"

	_ "modernc.org/sqlite"
)

func initDB(dbName string) {
	db, err := sql.Open("sqlite", dbName)
	if err != nil {
		panic(err)
	}

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	_, err = tx.Exec(`CREATE TABLE IF NOT EXISTS Measurements(at integer, value double)`)
	if err != nil {
		panic(err)
	}

	result, err := tx.Query(`SELECT COUNT(*) FROM Measurements`)
	if err != nil {
		panic(err)
	}

	result.Next()
	var nItems int
	if err = result.Scan(&nItems); err != nil {
		panic(err)
	}

	tx.Commit()
	if nItems < 10000 {
		tx, err := db.Begin()
		if err != nil {
			panic(err)
		}
		fmt.Printf("nRows: %d, inserting data ...\n", nItems)
		tm := time.Now().UnixMilli()
		stmt, err := tx.Prepare(`INSERT INTO Measurements(at, value) VALUES(?, ?)`)
		if err != nil {
			panic(err)
		}

		for i := 0; i < 10000; i++ {
			if _, err := stmt.Exec(tm, rand.Float64()); err != nil {
				panic(err)
			}
			tm -= 100
		}
		tx.Commit()
	}
}

func streamMain() {
	initDB("test.db")
	db, err := sql.Open("sqlite", "test.db")
	if err != nil {
		panic(err)
	}
	st := store.Store{DB: db}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	entries, err := st.Stream(ctx, store.Request{})
	if err != nil {
		panic(err)
	}

	filteredEntries := filters.MinFilter(0.001, entries)
	entryCh, errCh := filters.ErrFilter(filteredEntries)
	resultCh := filters.MovingAvg(0.5, 5, entryCh)
	var streamErr error
	go func() {
		for err := range errCh {
			if streamErr == nil {
				streamErr = err
				cancel()
			}
		}
	}()

	for entry := range resultCh {
		fmt.Printf("entry: %v\n", entry)
	}
	if streamErr != nil {
		fmt.Println(streamErr)
	}
}

func httpMain() {
	initDB("test.db")
	db, err := sql.Open("sqlite", "test.db")
	if err != nil {
		panic(err)
	}
	st := store.Store{DB: db}

	app := fiber.New()
	app.Get("/db", func(c *fiber.Ctx) error {
		data, err := st.Stream(c.Context(), store.Request{})
		if err != nil {
			fmt.Println(err)
			return c.SendStatus(500)
		}

		errCh := coder.EncodeFromChan(data, func(entry store.Entry) ([]byte, error) {
			msg := store.Message{
				At:    entry.At,
				Value: entry.Value,
			}
			if entry.Error != nil {
				msg.Error = entry.Error.Error()
			}
			return json.Marshal(msg)
		}, c.Write)

		err = <-errCh
		if err != nil {
			fmt.Println("Encode error", err)
			return c.SendStatus(500)
		}
		return c.SendStatus(200)
	})

	go func() {
		fmt.Println("Listening on :3000")
		err := app.Listen(":3000")
		if err != nil {
			panic(err)
		}
	}()

	resp, err := http.Get("http://localhost:3000/db")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	entries, rcvErr := coder.DecodeToChan(func(entry *store.Entry) error {
		var msg store.Message
		if err := decoder.Decode(&msg); err != nil {
			return err
		}
		entry.At = msg.At
		entry.Value = msg.Value
		if msg.Error != "" {
			entry.Error = fmt.Errorf(msg.Error)
		}

		return nil
	})

	filteredEntries := filters.MinFilter(0.001, entries)
	entryCh, errCh := filters.ErrFilter(filteredEntries)
	resultCh := filters.MovingAvg(0.5, 5, entryCh)

	go func() {
		for err := range errCh {
			fmt.Println(err)
		}
	}()
	for entry := range resultCh {
		fmt.Printf("entry: %v\n", entry)
	}
	err = <-rcvErr
	if err != nil {
		fmt.Println("Decode error", err)
	}
}

func main() {
	//streamMain()
	httpMain()
}
