# goaws

> Collection of small scripts & tools for AWS

# Credentials

Credentials can be specified with the `~/.aws/credentials` file.
The `default` profile is used unless specified by the `AWS_PROFILE` environment variable.
Credentials from the profile can be overriden by the `AWS_ACCESS_KEY_ID` and the `AWS_SECRET_ACCESS_KEY` environment variables.

# Region

The region can be specified along with the profile in the `~/.aws/config` file.
It can also be overriden by the `AWS_REGION` environment variable.

# sqs2js

Receives messages from the specified queue and invokes a command with the message body as STDIN.
The message is deleted from SQS only if the command succeed.

```bash
URL="https://sqs.us-east-1.amazonaws.com/123/s3-log-received"

# wait and print 1 message
sqs2js -url "$URL" -n 1 -cmd echo

# print all messages
sqs2js -url "$URL" -cmd bash echo

# pretty print 1 messages
sqs2js -url "$URL" -n 1 -cmd bash -- -c "jq ."

# print 1 batch of events from S3
sqs2js -url "$URL" -n 1 -cmd bash -- -c "cmd/s3files.sh"
```
