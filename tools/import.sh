#!/bin/sh

# check if argument is given
if [ -z "$1" ]; then
    echo "No argument supplied"
    exit 1
fi

# check if root
if [ "$(id -u)" != "0" ]; then
    echo "This script must be run as root"
    exit 1
fi

cd _tools

python ./import_$1.py
if [ $? -ne 0 ]; then
    echo "Error while importing $1"
    exit 2
fi
