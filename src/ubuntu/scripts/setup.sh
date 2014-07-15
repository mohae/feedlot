#!/usr/bin/env bash
# Update the system and install basics
yum update -y -q > /dev/null
yum upgrade -y -q > /dev/null
yum install -y -q curl wget git vim rsync openssh sudo
yum install -y gcc make gcc-c++ kernel-devel-`uname-r`
