#!/usr/bin/env bash

# Set up packer user for sudo
echo %packer ALL=NOPASSWD:ALL > /etc/sudoers.d/packer
chmod 0440 /etc/sudoers.d/packer

# Add the packer user to sudoers
usermod -a -G sudo packer

# setup packer keys using vagrant's insecure keys
# note: vagrant keys are well known and insecure, which is why we can add them
#       this way. For your private keys, make sure they do not end up in your 
#       Git repo, or any other publicly available resource. You should have a 
#       secure way of handling private keys.
mkdir /home/packer/.ssh
wget --no-check-certificate 'https://raw.githubusercontent.com/mitchellh/vagrant/master/keys/vagrant.pub' 
mv packer.pub /home/packer/.ssh/authorized_keys
chown -R packer /home/packer/.ssh
chmod 700 /home/packer/.ssh
chmod 600 /home/packer/.ssh/authorized_keys
