package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
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

	file := flag.String("f", "", "file with JSON data to convert to AWS DynamoDB")
	data := flag.String("d", "", "JSON data to convert to AWS DynamoDB")
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

	d := json.NewDecoder(r)
	for {
		err := read(d)

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Panic(err)
		}
	}
}

func read(d *json.Decoder) (err error) {
	var item interface{}

	err = d.Decode(&item)
	if err != nil {
		if err == io.EOF {
			return
		}

		fmt.Errorf("decode: %s", err)
		return
	}

	value := encode(item)

	if value.M == nil {
		write(value)
		return
	}

	i := 0
	for k, v := range value.M {
		if i == 0 {
			ws("{")
			i++
		} else {
			ws(",")
		}

		js(k)

		ws(":")
		write(v)
	}

	ws("}\n")
	return
}

func encode(value interface{}) *dynamodb.AttributeValue {
	switch item := value.(type) {
	case bool:
		return &dynamodb.AttributeValue{BOOL: aws.Bool(true)}

	case float64:
		f := strconv.FormatFloat(item, 'g', -1, 64)
		return &dynamodb.AttributeValue{N: aws.String(f)}

	case string:
		return &dynamodb.AttributeValue{S: aws.String(item)}

	case map[string]string:
		items := make(map[string]*dynamodb.AttributeValue)

		for k, v := range item {
			items[k] = &dynamodb.AttributeValue{S: aws.String(v)}
		}

		return &dynamodb.AttributeValue{M: items}

	case map[string]float64:
		items := make(map[string]*dynamodb.AttributeValue)

		for k, v := range item {
			f := strconv.FormatFloat(v, 'g', -1, 64)
			items[k] = &dynamodb.AttributeValue{N: aws.String(f)}
		}

		return &dynamodb.AttributeValue{M: items}

	default:
		v := reflect.ValueOf(value)
		t := v.Type()

		switch t.Kind() {
		case reflect.Ptr:
			if v.IsNil() {
				return &dynamodb.AttributeValue{NULL: aws.Bool(true)}
			}

			return encode(v.Elem().Interface())

		case reflect.Slice:
			n := v.Len()
			s := make([]*dynamodb.AttributeValue, 0, n)

			for i := 0; i < n; i++ {
				item := encode(v.Index(i).Interface())
				s = append(s, item)
			}

			return &dynamodb.AttributeValue{L: s}

		case reflect.Map:
			m := make(map[string]*dynamodb.AttributeValue)

			for _, i := range v.MapKeys() {
				item := encode(v.MapIndex(i).Interface())
				if item != nil {
					m[i.String()] = item
				}
			}

			return &dynamodb.AttributeValue{M: m}

		default:
			log.Panicf("unhandled kind: %s", t.Kind())
		}
	}

	return nil
}

func js(item interface{}) {
	p, err := json.Marshal(item)
	if err != nil {
		log.Panicf("encode: %s", err)
	}

	_, err = os.Stdout.Write(p)
	if err != nil {
		log.Panicf("stdout: %s", err)
	}
}

func ws(s string) {
	_, err := io.WriteString(os.Stdout, s)
	if err != nil {
		log.Panicf("stdout: %s", err)
	}
}

func write(item *dynamodb.AttributeValue) (n int, err error) {
	ws("{")
	defer ws("}")

	if item.B != nil {
		ws(`"B":`)
		js(item.B)
		return
	}

	if item.BOOL != nil {
		ws(`"BOOL":`)
		js(*item.BOOL)
		return
	}

	if item.BS != nil {
		for i, b := range item.BS {
			if i == 0 {
				ws(`"BS":[`)
			} else {
				ws(",")
			}

			js(b)
		}

		ws("]")
		return
	}

	if item.L != nil {
		for i, j := range item.L {
			if i == 0 {
				ws(`"L":[`)
			} else {
				ws(",")
			}

			write(j)
		}

		ws("]")
		return
	}

	if item.M != nil {
		k := 0
		for i, j := range item.M {
			if k == 0 {
				ws(`"M":{`)
				k++
			} else {
				ws(",")
			}

			js(i)
			ws(":")
			write(j)
		}

		ws("}")
		return
	}

	if item.N != nil {
		ws(`"N":`)
		js(item.N)
		return
	}

	if item.NS != nil {
		for i, n := range item.NS {
			if i == 0 {
				ws(`"NS":[`)
			} else {
				ws(",")
			}

			js(n)
		}

		ws("]")
		return
	}

	if item.NULL != nil {
		ws(`"NULL":`)
		js(*item.NULL)
		return
	}

	if item.S != nil {
		ws(`"S":`)
		js(item.S)
		return
	}

	if item.SS != nil {
		for i, s := range item.SS {
			if i == 0 {
				ws(`"SS":[`)
			} else {
				ws(",")
			}

			js(s)
		}

		ws("]")
		return
	}

	log.Panicf("empty attribute: %+v", item)
	return
}
