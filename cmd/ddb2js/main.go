package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			os.Exit(1)
		}
	}()

	log.SetFlags(0)

	file := flag.String("f", "", "file with AWS DynamoDB data to convert to JSON")
	data := flag.String("d", "", "AWS DynamoDB data to convert to JSON")
	flag.Parse()

	var r io.Reader

	if *file != "" {
		f, err := os.Open(*file)
		if err != nil {
			log.Panicf("open: %s", err)
		}

		r = f
		defer f.Close()
	}

	if *data != "" {
		r = strings.NewReader(*data)
	}

	if r == nil {
		r = os.Stdin
	}

	item := make(map[string]*dynamodb.AttributeValue)

	err := json.NewDecoder(r).Decode(&item)
	if err != nil {
		log.Panicf("decode: %s", err)
	}

	m := make(map[string]interface{})

	for k, v := range item {
		m[k] = decode(v)
	}

	js, err := json.Marshal(m)
	if err != nil {
		log.Panicf("encode: %s", err)
	}

	_, err = os.Stdout.Write(js)
	if err != nil {
		log.Panicf("stdout: %s", err)
	}
}

func decode(value *dynamodb.AttributeValue) interface{} {
	if value.BOOL != nil {
		return aws.BoolValue(value.BOOL)
	}

	if len(value.B) != 0 {
		return value.B
	}

	if len(value.BS) != 0 {
		s := make([][]byte, len(value.BS))

		for i, j := range value.BS {
			s[i] = j
		}

		return s
	}

	if len(value.L) != 0 {
		s := make([]interface{}, len(value.L))

		for i, j := range value.L {
			s[i] = decode(j)
		}

		return s
	}

	if len(value.M) != 0 {
		m := make(map[string]interface{})

		for k, v := range value.M {
			m[k] = decode(v)
		}

		return m
	}

	if value.N != nil {
		f, err := strconv.ParseFloat(aws.StringValue(value.N), 64)
		if err != nil {
			log.Panicf("parse: %s", err)
		}

		return f
	}

	if len(value.NS) != 0 {
		s := make([]float64, len(value.NS))

		for i, j := range value.NS {
			f, err := strconv.ParseFloat(aws.StringValue(j), 64)
			if err != nil {
				log.Panicf("parse: %s", err)
			}

			s[i] = f
		}

		return s
	}

	if value.NULL != nil {
		return nil
	}

	if value.S != nil {
		return aws.StringValue(value.S)
	}

	if len(value.SS) != 0 {
		return aws.StringValueSlice(value.SS)
	}

	log.Panicf("empty attribute: %+v", value)
	return nil
}
