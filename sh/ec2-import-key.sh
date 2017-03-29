#!/bin/bash

while [[ $# -gt 1 ]]
do
  key="$1"
  case $key in
    -p|--profile)
    PROFILE="--profile $2"
    shift
    ;;
    -r|--region)
    REGION="--region $2"
    shift
    ;;
    -n|--name)
    NAME=$2
    shift
    ;;
    -k|--key)
    SSHKEY=$2
    shift
    ;;
    *)
	echo unknown option: $key
    ;;
  esac
  shift
done

KEY=`openssl rsa -in $SSHKEY -pubout | head -n -1 | tail -n +2 | paste -sd ''`

aws $PROFILE $REGION ec2 import-key-pair \
	--key-name $NAME \
	--public-key-material "$KEY"
