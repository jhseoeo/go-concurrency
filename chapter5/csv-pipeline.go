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

func workerPoolpipelineStage[IN any, OUT any](input <-chan IN, output chan<- OUT, process func(IN) OUT, numWorkers int) {
	defer close(output)

	wg := sync.WaitGroup{}
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for data := range input {
				output <- process(data)
			}
		}()
	}

	wg.Wait()
}

func asyncPipeline(input *csv.Reader) {
	parseInputCh := make(chan []string)
	convertInputCh := make(chan Record)
	encodeInputCh := make(chan Record)
	outputCh := make(chan []byte)
	done := make(chan struct{})

	const numWorkers = 2
	go workerPoolpipelineStage(parseInputCh, convertInputCh, parse, numWorkers)
	go workerPoolpipelineStage(convertInputCh, encodeInputCh, convert, numWorkers)
	go workerPoolpipelineStage(encodeInputCh, outputCh, encode, numWorkers)

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

	asyncPipeline(input)
}
