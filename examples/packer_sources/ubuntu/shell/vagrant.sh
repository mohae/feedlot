#!/usr/bin/env bash
<<<<<<< HEAD

=======
>>>>>>> c5c1889d741c448ab595f6ebdcddfd5acc59b2d8
# Set up vagrant user for sudo
echo %vagrant ALL=NOPASSWD:ALL > /etc/sudoers.d/vagrant
chmod 0440 /etc/sudoers.d/vagrant

# Add the vagrant user to sudoers
usermod -a -G sudo vagrant

<<<<<<< HEAD
# setup vagrant keys
# note: vagrant keys are well known and insecure, which is why we can add them
#       this way. For your private keys, make sure they do not end up in your 
#       Git repo, or any other publicly available resource. You should have a 
#       secure way of handling private keys, and NO ROT13 is not secure! :)
mkdir /home/vagrant/.ssh
wget --no-check-certificate 'https://raw.githubusercontent.com/mitchellh/vagrant/master/keys/vagrant.pub' -O /home/vagrant/.ssh/authorized_keys
=======
# setup vagrant keys using vagrant's insecure keys
# note: vagrant keys are well known and insecure, which is why we can add them
#       this way. For your private keys, make sure they do not end up in your 
#       Git repo, or any other publicly available resource. You should have a 
#       secure way of handling private keys.
mkdir /home/vagrant/.ssh
wget --no-check-certificate 'https://raw.githubusercontent.com/mitchellh/vagrant/master/keys/vagrant.pub' 
mv vagrant.pub /home/vagrant/.ssh/authorized_keys
>>>>>>> c5c1889d741c448ab595f6ebdcddfd5acc59b2d8
chown -R vagrant /home/vagrant/.ssh
chmod 700 /home/vagrant/.ssh
chmod 600 /home/vagrant/.ssh/authorized_keys
