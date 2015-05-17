#!/usr/bin/env bash
# Install VirtualBox Guest Additions - Server style
VBOX_VERSION=$(cat /home/vagrant/.vbox_version)
cd /tmp
mount -o loop /home/vagrant/VBoxGuestAdditions_$VBOX_VERSION.iso /mnt
sh /mnt/VBoxLinuxAdditions.run
unmount /mnt
rm -rf /home/vagrant/VBoxGuestAdditions_*.iso
