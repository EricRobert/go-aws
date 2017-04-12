#!/bin/bash
jq -r '.Message' | jq -r '.Records[] | "s3://\(.s3.bucket.name)/\(.s3.object.key)"'
