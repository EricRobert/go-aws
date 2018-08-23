# Publish to SNS

This lambda is useful to publish messages to an SNS topic.

It is mostly used with:
- AWS Kinesis
- AWS DynamoDB streams
- AWS SQS

## Usage

```
curl -o main.zip https://github.com/EricRobert/go-aws/archive/publish-to-sns.0.1.zip
export NAME=updates
export ROLE=arn:aws:iam::xyz:role/lambda-publish-updates
aws lambda create-function --function-name $NAME --runtime go1.x --role $ARN --handler main --zip-file fileb://main.zip
```
