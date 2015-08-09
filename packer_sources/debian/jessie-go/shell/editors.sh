# editors
# install vim
apt-get install vim

# install spf13-vim: http://vim.spf13.com/
sudo vagrant curl https://j.mp/spf13-vim3 -L > spf13-vim.sh && sh spf13-vim.sh

# install atom: https://atom.io
# 1.0.5 is latest at time of script creation
wget https://github.com/atom/atom/releases/download/v1.0.5/atom-amd64.deb
sudo dpkg --install atom-amd64.deb