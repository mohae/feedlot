# Go should be installed under the user not root
su vagrant
# install go
wget https://storage.googleapis.com/golang/go1.4.2.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.4.2.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
mkdir $HOME/code
mkdir $HOME/code/go
export GOPATH=/$HOME/code/go
export PATH=$PATH:$GOPATH/bin

# liteide: https://github.com/visualfc/liteide
# get the latest liteide
wget http://sourceforge.net/projects/liteide/files/latest/download liteide-latest.tar.bz2
# extract download to /usr/local
tar xzf liteide-latest-tar.bz2 /usr/local/

# delve debugger: https://github.com/derekparker/delve
go get -u github.com/derekparker/delve/cmd/dlv

# switch back to root
sudo su