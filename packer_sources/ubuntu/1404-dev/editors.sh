# editors
# install vim
apt-get install vim

# install to vagrant user
su vagrant
# install spf13-vim: http://vim.spf13.com/
curl https://j.mp/spf13-vim3 -L > spf13-vim.sh && sh spf13-vim.sh

# switch back to root
sudo su

# install atom: https://atom.io
# 1.0.5 is latest at time of script creation
wget https://github.com/atom/atom/releases/download/v1.0.5/atom-amd64.deb
dpkg --install atom-amd64.deb
apt-get install -y -f