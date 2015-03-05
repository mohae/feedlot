#!/usr/bin/env bash
# Install lxde 
apt-get -y install lxde-core

#  Install spf13-vim: https://github.com/spf13/spf13-vim for more info
#runuser -l vagrant -c 'curl https://j.mp/spf13-vim3  -L > spf13-vim.sh && sh spf13-vim.sh'
runuser -l vagrant -c 'sh  <(curl https://j.mp/spf13-vim3 -L)'

apt-get install aptitude openssh-server p7zip-full ssh symlinks  

# chrome
# get the key and add it
wget -q -O https://dl-ssl.google.com/linux/linux_signing_key.pub | sudo apt-key add -
# set up the repository
sudo sh -c 'echo "deb http://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google.list'
apt-get -y update
apt-get -y install google-chrome-stable

# Set up aliases
runuser -l vagrant -c "echo alias c=\'clear\' >> ~/.bash_aliases"
