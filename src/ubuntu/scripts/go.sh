#!/usr/bin/env bash
# Dev stuff installs some common and not so common languages and tools
#
# Installed:
#	Chromium
#	Chrome
#	Git
#
# Supported:
#	C
#	Go
##	
# Notes: 
#	* This neither sets up nor configures iptables. Firewall adjustments are 
#	  left to you, for now.
#	* Assumes running as root
VERSION="1.4.2"
DFILE = "go$VERSION.linux-amd64.tar.gz"
# git
apt-get -y install git

# setup directory for source code
run user -l vagrant -c 'mkdir $HOME/code'
run user -l vagrant -c 'mkdir $HOME/code/go'

# Get the latest installer
run user -l vagrant -c wget https://storage.googleapis.com/golang/$DFILE -O /tmp/go.tar.gz
if [ $? -ne 0 ]; then
  echo "Download of $DFILE failed! Go was not installed. Exiting."
  exit 1
fi

run user -l vagrant -c tar -C /usr/local -xzf /tmp/go.tar.gz
rm /tmp/go.tar.gz

# set up path/workspace for go
run user -l vagrant -c 'mkdir $HOME/code/go'
run user -l vagrant -c 'echo # Add Go path info'
run user -l vagrant -c 'echo export PATH=$PATH:/usr/local/go/bin >> ~/.bashrc'
run user -l vagrant -c 'echo export GOPATH=$HOME/code/go  >> ~/.bashrc'
run user -l vagrant -c 'echo export PATH=$PATH:$GOPATH/bin >> ~/.bashrc'

# go get tools
run user -l vagrant -c 'go get golang.org/x/tools/cmd/...'
# vim-go
run user -l vagrant -c 'git clone https://github.com/gmarik/Vundle.vim.git ~/.vim/bundle/Vundle.vim'

