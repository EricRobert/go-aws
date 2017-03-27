package main

import (
	"bytes"
	"flag"
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			os.Exit(1)
		}
	}()

	log.SetFlags(0)

	url := flag.String("url", "", "url of the SQS queue")
	cmd := flag.String("cmd", "", "command to invoke for each SQS message")
	n := flag.Int("n", 0, "number of SQS messages to retrieve (0 means unlimited)")
	t := flag.Duration("t", 0, "time to wait between requests to SQS")
	batch := flag.Int("batch", 10, "batch size (must be between 1 and 10)")
	flag.Parse()

	if url == nil || *url == "" || cmd == nil || *cmd == "" {
		usage()
	}

	if k := *batch; k < 1 || k > 10 {
		log.Panicf("batch size %d is not between 1 and 10", k)
	}

	s, err := session.NewSession()
	if err != nil {
		log.Panicf("session: %s", err)
	}

	q := &queue{
		q:     sqs.New(s),
		batch: *batch,
		n:     int64(*n),
		t:     *t,
		url:   *url,
		cmd:   *cmd,
		args:  flag.Args(),
	}

	for {
		if q.n != 0 && q.i == q.n {
			break
		}

		if err = q.read(); err != nil {
			break
		}

		if q.i == q.n {
			break
		}

		if q.t != 0 {
			time.Sleep(q.t)
		}
	}

	if err != nil {
		log.Panicf("read: %s", err)
	}
}

func usage() {
	log.Println(os.Args[0], "-url=sqs-url [-n=N] -cmd=prog [args...]")
	log.Println()
	flag.PrintDefaults()
	os.Exit(-1)
}

type queue struct {
	q     *sqs.SQS
	batch int
	i     int64
	n     int64
	t     time.Duration
	url   string
	cmd   string
	args  []string
}

func (q *queue) read() (err error) {
	j := int64(q.batch)

	if q.n != 0 {
		j = q.n - q.i
	}

	if j > 10 {
		j = 10
	}

	args := &sqs.ReceiveMessageInput{
		QueueUrl:            &q.url,
		MaxNumberOfMessages: &j,
		WaitTimeSeconds:     aws.Int64(20), // using SQS long polling
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

	if err = cmd.Run(); err == nil {
		return
	}

	if _, ok := err.(*exec.ExitError); !ok {
		log.Panicf("run: %s", err)
	}

	return
}
