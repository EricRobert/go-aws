# goaws

> Collection of simple tools and small shell scripts for AWS

## Credentials

Credentials can be specified with the `~/.aws/credentials` file.
The `default` profile is used unless specified by the `AWS_PROFILE` environment variable.

They can be overriden by the `AWS_ACCESS_KEY_ID` and the `AWS_SECRET_ACCESS_KEY` environment variables.

## Region

The region can be specified along with the profile in the `~/.aws/config` file.
It can also be overriden by the `AWS_REGION` environment variable.

# s3cat

Downloads an S3 object to STDOUT.

```bash
# download a file
s3cat s3://bucket/object > local-file

# print logs
s3cat s3://bucket/file.log.gz | zcat
```

# sqscat

Receives messages from the specified queue and invokes a command with the message body as STDIN.
The message is deleted from SQS only if the command succeed.

```bash
URL="https://sqs.us-east-1.amazonaws.com/123/s3-log-received"

# wait and print 1 message
sqscat -url "$URL" -n 1 -cmd cat

# print all messages
sqscat -url "$URL" -cmd cat

# pretty print 1 messages
sqscat -url "$URL" -n 1 -cmd bash -- -c "jq ."

# print 1 batch of events from S3
sqscat -url "$URL" -n 1 -cmd bash -- -c "sh/sqs-s3-events.sh"
```

# js2ddb & ddb2js

Performs the conversion between JSON and the DynamoDB format.

```bash
# create a new item in a table from a JSON object
aws dynamodb put-item --table-name samples --item "`cat sample.json | js2ddb`"

# scan table
aws dynamodb scan --table-name samples | jq .Items[] | ddb2js

```

# ec2-autoscaling-ssh

Run a command via SSH on all running instances of an auto scaling group.

```bash
# uptime on all instances?
ec2-autoscaling-ssh.sh --group my-group --sh uptime -p
```

# ec2-import-key

```bash
# import from private key
ec2-import-key.sh --name my-key --key ~/.ssh/id_rsa
```
