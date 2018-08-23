package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

// main installs a lambda that publishes every message as-is on SNS.
func main() {
	s := sns.New(session.New())

	lambda.Start(func(ctx context.Context, js json.RawMessage) (string, error) {
		c, ok := lambdacontext.FromContext(ctx)
		if !ok {
			log.Panicf("no context")
		}

		// use the lambda ARN to figure out the associated SNS ARN
		// so something like this:
		// arn:aws:lambda:us-east-1:xyz:function:abc-updates
		f := strings.Split(c.InvokedFunctionArn, ":")
		if len(f) != 7 {
			log.Panicf("cannot use arn: %s", c.InvokedFunctionArn)
		}

		// has to become this:
		// arn:aws:sns:us-east-1:xyz:abc-updates
		topic := fmt.Sprintf("arn:aws:sns:%s:%s:%s", f[3], f[4], f[6])

		m := string(js)

		r, err := s.Publish(&sns.PublishInput{Message: &m, TopicArn: &topic})
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("success: %s", *r.MessageId), nil
	})
}
