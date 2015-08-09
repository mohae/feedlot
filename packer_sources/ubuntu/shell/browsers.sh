# install commonly used programs
# chrome
wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb
dpkg -i google-chrome-stable_current_i386.deb
apt-get install -y -f

# chromium
apt-get update
apt-get install -y chromium-browser

# firefox
apt-get install -y firefox