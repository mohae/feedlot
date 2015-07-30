#!/usr/bin/env bash
# Dev stuff installs some common and not so common languages and tools
#
# Installed:
#	Chromium
#	Chrome
#	Git
#	Mercurial
#
# Supported:
#	C
#	Go
#	Dart
#	PHP 5.x
#	Python
#	Ruby
##	
# Notes: 
#	* This neither sets up nor configures iptables. Firewall adjustments are 
#	  left to you, for now.
#	* Assumes running as root

# bazaar
apt-get -y install bzr

# git
apt-get -y install git

# hg (mercurial)
apt-get -y install mercurial

# Misc
apt-get -y install make bison 

# setup directory for code
run user -l vagrant -c 'mkdir $HOME/code'

# C & C++ compiler stuff: i386 only needed for 32bit compiles
apt-get -y install build-essential

# Python
# Add sqlite libs for sqlite support
apt-get -y install libsqlite3-dev sqlite3 bzip2 libbz2-dev

# Download and compile Python
wget http://www.python.org/ftp/python/3.4.1/Python-3.4.1.tar.xz
tar xJf Python-3.4.1.tar.xz
cd Python-3.3.5
./configure --prefix=/opt/python3.4
make && sudo make install

# symlink to make a `py` command
runuser -l vagrant -c 'mkdir ~/bin'
runuser -l vagrant -c 'ln -s /opt/python3.4/python3.4 ~/bin/py'

# Required to install Go from source:
apt-get -y install gcc libc6-dev libc6-dev-i386 

# Install Go from source: clone
runuser -l vagrant -c 'hg clone -u release https://code.google.com/p/go'

# And switch to development branch. Comment out if no changes are going to be
# made to the source.
runuser -l vagrant -c  'hg update default'

# Build it...if 'ALL TESTS PASSED' it was ok.
runuser -l vagrant -c  'cd go/src'
runuser -l vagrant -c './all.bash'

# get additional tools
runuser -l vagrant -c 'go get code.google.com/p/go.tools/cmd/...'

# add your github user as a subdirectory of github.com
run user -l vagrant -c 'mkdir -p $GOPATH/src/github.com/'

# set up path/workspace for go
run user -l vagrant -c 'mkdir $HOME/code/go'
run user -l vagrant -c 'echo # Add Go path info'
run user -l vagrant -c 'echo export PATH=$PATH:/usr/local/go/bin >> ~/.bashrc'
run user -l vagrant -c 'echo export GOPATH=$HOME/code/go  >> ~/.bashrc'
run user -l vagrant -c 'echo export PATH=$PATH:$GOPATH/bin >> ~/.bashrc'

# Java
apt-get -y  install python-software-properties
apt-get--y repository ppa:webupd8team/java
apt-get -y  update
apt-get -y install oracle-java7-installer

# Dart
run user -l vagrant -c 'cd ~'
run user -l vagrant -c 'wget http://storage.googleapis.com/dart-archive/channels/stable/release/latest/editor/darteditor-linux-x64.zip'
run user -l vagrant -c 'tar xzf darteditor-linux-x64.zip ./'

# ruby
apt-get -y install ruby


