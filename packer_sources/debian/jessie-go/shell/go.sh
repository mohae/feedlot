# install go
sudo vagrant tar -C /usr/local -xzf go1.4.2.linux-amd64.tar.gz
sudo vagrant export PATH=$PATH:/usr/local/go/bin
sudo vagrant mkdir $HOME/code
sudo vagrant mkdir $HOME/code/go
sudo vagrant export GOPATH=/$HOME/code/go
sudo vagrant export PATH=$PATH:$GOPATH/bin

# liteide: https://github.com/visualfc/liteide
# get the latest liteide
sudo vagrant wget http://sourceforge.net/projects/liteide/files/latest/download liteide-latest.tar.bz2
# extract download to /usr/local
sudo vagrant tar xzf liteide-latest-tar.bz2 /usr/local/

# delve debugger: https://github.com/derekparker/delve
sudo vagrant go get -u github.com/derekparker/delve/cmd/dlv