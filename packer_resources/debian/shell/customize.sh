#!/usr/bin/env bash
# Install lxde 
apt-get -y install lxde-core

#  Install spf13-vim: https://github.com/spf13/spf13-vim for more info
#runuser -l vagrant -c 'curl https://j.mp/spf13-vim3  -L > spf13-vim.sh && sh spf13-vim.sh'
runuser -l vagrant -c 'sh  <(curl https://j.mp/spf13-vim3 -L)'

apt-get install aptitude openssh-server p7zip-full ssh symlinks  

# Set up aliases
runuser -l vagrant -c "echo alias c=\'clear\' >> ~/.bash_aliases"
