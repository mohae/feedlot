# install some stuff we'll need
apt-get update
apt-get install -y build-essential libssl-dev libcurl4-gnutls-dev libexpat1-dev gettext

# install git
apt-get install -y git

# install c and c++ compilers and manpages
apt-get install -y gcc g++ manpages-dev

# install dart
# Enable HTTPS for apt.
apt-get install apt-transport-https
# Get the Google Linux package signing key.
sh -c 'curl https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add -'
# Set up the location of the stable repository.
sh -c 'curl https://storage.googleapis.com/download.dartlang.org/linux/debian/dart_stable.list > /etc/apt/sources.list.d/dart_stable.list'
apt-get update
apt-get install dart

# java
apt-get install default-jre
apt-get install default-jdk

#
apt-get install lua5.2

# switch to vagrant user
su vagrant

# nodejs: from nvm. 0.12.7 is current latest. Install what works for you
curl https://raw.githubusercontent.com/creationix/nvm/v0.16.1/install.sh | sh
source ~/.profile
nvm install 0.12.7
nvm use 0.12.7

# rust
curl -sSf https://static.rust-lang.org/rustup.sh | sh

# switch back to root
sudo su
