#!/usr/bin/env bash
apt-get -y install deborphan
deborphan | xargs sudo apt-get purge -y
apt-get clean
apt-get autoclean
apt-get autoremove
apt-get purge
