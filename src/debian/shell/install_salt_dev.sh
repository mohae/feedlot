#!/usr/bin/env bash
# install dev version from github using salt-bootstrap
curl -L https://bootstrap.saltstack.com -o install_salt.sh
sh install_salt.sh git develop
