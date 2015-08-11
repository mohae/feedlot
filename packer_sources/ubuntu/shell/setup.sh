#!/usr/bin/env bash
# Update the system and install basics
apt-get update  -q > /dev/null
apt-get upgrade -y  -q > /dev/null
apt-get install -y -q curl wget vim rsync tree tmux sudo
apt-get -y -q install build-essential linux-headers-$(uname -r)
