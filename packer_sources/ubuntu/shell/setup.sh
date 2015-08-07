#!/usr/bin/env bash
apt-get update
apt-get upgrade -y
apt-get -y install linux-headers-$(uname -r)
apt-get install -y curl wget rsync
