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

type sequenced interface {
	Sequence() int
}

func (r Record) Sequence() int {
	return r.Row
}

type fanInRecord[T sequenced] struct {
	index int
	data  T
	pause chan struct{}
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

func orderedFanIn[T sequenced](done <-chan struct{}, channels ...<-chan T) <-chan T {
	fanInCh := make(chan fanInRecord[T])
	wg := sync.WaitGroup{}

	for i := range channels {
		wg.Add(1)

		pauseCh := make(chan struct{})
		go func(pause chan struct{}, index int) {
			defer wg.Done()
			for {
				var ok bool
				var data T

				select {
				case data, ok = <-channels[index]:
					if !ok {
						return
					}
					fanInCh <- fanInRecord[T]{index: index, data: data, pause: pause}
				case <-done:
					return
				}

				select {
				case <-pause:
				case <-done:
					return
				}
			}
		}(pauseCh, i)
	}

	go func() {
		wg.Wait()
		close(fanInCh)
	}()

	outputCh := make(chan T)
	go func() {
		defer close(outputCh)
		expected := 1
		queuedData := make([]*fanInRecord[T], len(channels))
		for in := range fanInCh {
			// 순서가 맞는 데이터는 바로 전달
			if in.data.Sequence() == expected {
				select {
				case outputCh <- in.data:
					in.pause <- struct{}{}
					expected++
					allDone := false
					// 큐에 저장된 다음 데이터가 있는지 확인
					for !allDone {
						allDone = true
						for i, d := range queuedData {
							if d != nil && d.data.Sequence() == expected {
								select {
								case outputCh <- d.data:
									queuedData[i] = nil
									d.pause <- struct{}{}
									expected++
									allDone = false
								case <-done:
									return
								}
							}
						}
					}
				case <-done:
					return
				}
			} else {
				// 순서가 맞지 않는 데이터는 큐에 일시 저장
				in := in
				queuedData[in.index] = &in
			}
		}
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
	convertOutputCh := orderedFanIn(done, fanInCh...)
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
