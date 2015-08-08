# install commonly used programs
# chrome
wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb
dpkg -i google-chrome-stable_current_i386.deb

# chromium
apt-get update
apt-get install -y chromium-browser

# firefox
apt-get install -y firefox

# install vim
apt-get install vim

# install spf13-vim: http://vim.spf13.com/
sudo vagrant curl https://j.mp/spf13-vim3 -L > spf13-vim.sh && sh spf13-vim.sh

# install atom: https://atom.io
# 1.0.5 is latest at time of script creation
wget https://github.com/atom/atom/releases/download/v1.0.5/atom-amd64.deb
sudo dpkg --install atom-amd64.deb