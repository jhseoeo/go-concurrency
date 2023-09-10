package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"
)

type Record struct {
	Row    int
	Height float64
	Weight float64
}

func newRecord(in []string) (Record, error) {
	row, err := strconv.Atoi(in[0])
	if err != nil {
		return Record{}, err
	}

	height, err := strconv.ParseFloat(in[1], 64)
	if err != nil {
		return Record{}, err
	}

	weight, err := strconv.ParseFloat(in[2], 64)
	if err != nil {
		return Record{}, err
	}

	return Record{
		Row:    row,
		Height: height,
		Weight: weight,
	}, nil
}

func parse(input []string) Record {
	record, err := newRecord(input)
	if err != nil {
		panic(err)
	}
	return record
}

func convert(input Record) Record {
	input.Height = input.Height * 2.54
	input.Weight = input.Weight * 0.45359237
	return input
}

func encode(input Record) []byte {
	data, err := json.Marshal(input)
	if err != nil {
		panic(err)
	}
	return data
}

func cancelablePipelineStage[IN any, OUT any](input <-chan IN, done <-chan struct{}, process func(IN) OUT) <-chan OUT {
	outputCh := make(chan OUT)
	go func() {
		for {
			select {
			case data, ok := <-input:
				if !ok {
					close(outputCh)
					return
				}
				outputCh <- process(data)
			case <-done:
				return
			}
		}
	}()
	return outputCh
}

func fanIn[T any](done <-chan struct{}, channels ...<-chan T) <-chan T {
	outputCh := make(chan T)
	wg := sync.WaitGroup{}

	for _, ch := range channels {
		wg.Add(1)
		go func(input <-chan T) {
			defer wg.Done()
			for {
				select {
				case data, ok := <-input:
					if !ok {
						return
					}
					outputCh <- data
				case <-done:
					return
				}
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(outputCh)
	}()
	return outputCh
}

func fanInPipeline(input *csv.Reader) {
	parseInputCh := make(chan []string)
	done := make(chan struct{})
	convertInputCh := cancelablePipelineStage(parseInputCh, done, parse)

	const numWorkers = 2
	fanInCh := make([]<-chan Record, 0)
	for i := 0; i < numWorkers; i++ {
		convertOutputCh := cancelablePipelineStage(convertInputCh, done, convert)
		fanInCh = append(fanInCh, convertOutputCh)
	}
	convertOutputCh := fanIn(done, fanInCh...)
	outputCh := cancelablePipelineStage(convertOutputCh, done, encode)

	go func() {
		for data := range outputCh {
			fmt.Println(string(data))
		}
		close(done)
	}()

	input.Read() // skip header
	for {
		rec, err := input.Read()
		if err == io.EOF {
			close(parseInputCh)
			break
		}
		if err != nil {
			panic(err)
		}
		parseInputCh <- rec
	}
	<-done
}

func main() {
	// open csv file
	file, err := os.Open("sample.csv")
	if err != nil {
		panic(err)
	}

	// create csv reader
	input := csv.NewReader(file)

	fanInPipeline(input)
}
