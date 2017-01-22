#!/bin/bash
jq -r '.Records[] | "s3://\(.s3.bucket.name)/\(.s3.object.key)"'
