#!/usr/bin/env bash
# Update the system and install basics
apt-get update -q > /dev/null
apt-get upgrade -y -q > /dev/null
apt-get install -y -q curl wget git vim rsync tmux tree sudo
apt-get -y -q install build-essential linux-headers-$(uname -r)
