package store

import (
	"context"
	"database/sql"
	"time"
)

type Store struct {
	DB *sql.DB
}

type Entry struct {
	At    time.Time
	Value float64
	Error error
}

type AboveThresholdEntry struct {
	Entry
	Avg float64
}

type Request struct{}

type Message struct {
	At    time.Time `json:"at"`
	Value float64   `json:"value"`
	Error string    `json:"error,omitempty"`
}

func (svc Store) Stream(ctx context.Context, req Request) (<-chan Entry, error) {
	rows, err := svc.DB.Query(`SELECT at, value FROM measurements`)
	if err != nil {
		return nil, err
	}

	ret := make(chan Entry)
	go func() {
		defer close(ret)
		defer rows.Close()
		for {
			var at int64
			var entry Entry
			select {
			case <-ctx.Done():
				return
			default:
			}

			if !rows.Next() {
				break
			}
			err := rows.Scan(&at, &entry.Value)
			if err != nil {
				ret <- Entry{Error: err}
				continue
			}
			entry.At = time.UnixMilli(at)
			ret <- entry
		}
		err := rows.Err()
		if err != nil {
			ret <- Entry{Error: err}
		}
	}()

	return ret, nil
}
