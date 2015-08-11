#!/usr/bin/env bash
apt-get update
apt-get upgrade -y
apt-get -y install linux-headers-$(uname -r)
<<<<<<< HEAD
apt-get install -y curl wget rsync sudo
=======
apt-get install -y curl wget rsync
>>>>>>> c5c1889d741c448ab595f6ebdcddfd5acc59b2d8
