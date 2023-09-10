package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
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

func syncPipeline(input *csv.Reader) {
	input.Read() // skip header

	for {
		rec, err := input.Read()
		if err == io.EOF {
			return
		}
		if err != nil {
			panic(err)
		}

		out := encode(convert(parse(rec)))
		fmt.Println(string(out))
	}
}

func main() {
	// open csv file
	file, err := os.Open("sample.csv")
	if err != nil {
		panic(err)
	}

	// create csv reader
	input := csv.NewReader(file)

	syncPipeline(input)
}
