#!/usr/bin/env bash
# Add/set the sudoers stuff
apt-get install -y sudo
groupadd -r admin

# Back up before making changes
cp /etc/sudoers /etc/sudoers.orig

# Exempt.sudoers
sed -i -e '/Defaults\s\+env_reset/a Defaults\texempt_group=sudo' /etc/sudoers

# No passwords for admins
sed -i -e 's/%admin ALL=(ALL) ALL/%admin ALL=NOPASSWD:ALL/g' /etc/sudoers

echo "UseDNS no" >> /etc/ssh/sshd_config