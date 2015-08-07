#!/usr/bin/env bash
# sbin
#sed -i "s/usr\/games/usr\/games:\/usr\/local\/sbin:\/usr\/sbin:\/sbin/" /etc/profile
# Update the system and install basics
apt-get update -y -q > /dev/null
apt-get upgrade -y -q > /dev/null
apt-get install -y -q curl wget git vim rsync
