#!/usr/bin/env bash
# Add/set the sudoers stuff
<<<<<<< HEAD
groupadd -r admin
groupadd -r sudo
=======
apt-get install -y sudo
groupadd -r admin
>>>>>>> c5c1889d741c448ab595f6ebdcddfd5acc59b2d8

# Back up before making changes
cp /etc/sudoers /etc/sudoers.orig

# Exempt.sudoers
sed -i -e '/Defaults\s\+env_reset/a Defaults\texempt_group=sudo' /etc/sudoers

# No passwords for admins
sed -i -e 's/%admin ALL=(ALL) ALL/%admin ALL=NOPASSWD:ALL/g' /etc/sudoers

echo "UseDNS no" >> /etc/ssh/sshd_config