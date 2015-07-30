#!/usr/bin/env bash

# Set up sudo
echo %packer ALL=NOPASSWD:ALL > /etc/sudoers.d/packer
chmod 0440 /etc/sudoers.d/packer

# Setup sudo to allow no-password sudo for "sudo"
usermod -a -G sudo packer

# Install Vagrant keys: packer user will use vagrant's insecure keys
mkdir /home/packer/.ssh
wget --no-check-certificate 'https://raw.githubusercontent.com/mitchellh/vagrant/master/keys/vagrant.pub' 
mv vagrant.pub /home//packer.ssh/authorized_keys
chown -R packer /home/packer/.ssh
chmod 700 /home/packer/.ssh
chmod 600 /home/packer/.ssh/authorized_keys
