#!/usr/bin/env bash
# Add/set the sudoers stuff
groupadd -r admin
groupadd -r sudo

# Back up before the screw up
cp /etc/sudoers /etc/sudoers.orig

# Make the sudo group exempt. 
sed -i -e '/Defaults\s\+env_reset/a Defaults\texempt_group=sudo' /etc/sudoers

# Let Admins use no password.
sed -i -e 's/%admin ALL=(ALL) ALL/%admin ALL=NOPASSWD:ALL/g' /etc/sudoers
