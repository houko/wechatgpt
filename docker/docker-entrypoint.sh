#!/bin/sh
set -e

# do real entrypoint here
echo "receiving comand:" "$@"
# https://stackoverflow.com/a/17529221
( "$@" )
