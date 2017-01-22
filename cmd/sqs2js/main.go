package main

import (
	"bytes"
	"flag"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func main() {
	log.SetFlags(0)

	url := flag.String("url", "", "url of SQS queue")
	cmd := flag.String("cmd", "", "command to invoke for each message")
	n := flag.Int("n", 0, "number of messages to process")
	flag.Parse()

	if url == nil || *url == "" || cmd == nil || *cmd == "" {
		usage()
	}

	args := flag.Args()

	log.Println(*url)
	log.Println(*cmd, args)

	s, err := session.NewSession()
	if err != nil {
		log.Fatalln("session:", err)
	}

	q := &queue{
		q:    sqs.New(s),
		n:    10,
		url:  *url,
		cmd:  *cmd,
		args: flag.Args(),
	}

	for {
		if n != nil {
			k := int64(*n)

			if q.i == k {
				break
			}

			if left := k - q.i; q.n > left {
				q.n = left
			}
		}

		if err = q.read(); err != nil {
			break
		}
	}

	if err != nil {
		os.Exit(-1)
	}
}

func usage() {
	log.Println("sqs2js -url=sqs-url [-n=N] -cmd=prog [args...]")
	log.Println()
	flag.PrintDefaults()
	os.Exit(-1)
}

type queue struct {
	q    *sqs.SQS
	i    int64
	n    int64
	url  string
	cmd  string
	args []string
}

func (q *queue) read() (err error) {
	args := &sqs.ReceiveMessageInput{
		QueueUrl:            &q.url,
		MaxNumberOfMessages: &q.n,
		WaitTimeSeconds:     aws.Int64(20),
	}

	r, err := q.q.ReceiveMessage(args)
	if err != nil {
		log.Println("receive message:", err)
		return
	}

	process := func(m *sqs.Message) (err error) {
		if err = q.invoke([]byte(*m.Body)); err != nil {
			return
		}

		args := &sqs.DeleteMessageInput{
			QueueUrl:      &q.url,
			ReceiptHandle: m.ReceiptHandle,
		}

		_, err = q.q.DeleteMessage(args)
		if err != nil {
			log.Println("delete message:", err)
		}

		return
	}

	for _, m := range r.Messages {
		if err = process(m); err != nil {
			break
		}

		q.i++
	}

	return
}

func (q *queue) invoke(p []byte) (err error) {
	cmd := exec.Command(q.cmd, q.args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	r := io.MultiReader(bytes.NewReader(p), bytes.NewReader([]byte("\n")))
	cmd.Stdin = r
	err = cmd.Run()
	return
}
