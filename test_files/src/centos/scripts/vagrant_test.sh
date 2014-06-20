#!/usr/bin/env bash
# setup vagrant stuff

mkdir /home/vagrant/.ssh
wget --no-check-certificate -O 'https://github.com/mitchellh/vagrant/raw/master/keys/vagrant.pub' /home/vagrant/.ssh/authorized_keys
mv authorized_keys /home/vagrant/.ssh
chown -R vagrant /home/vagrant/.ssh
chmod 700 /home/vagrant/.ssh
chmod 600 /home/vagrant/.ssh/authorized_keys
