#!/usr/bin/env bash
# Install VMWare tools: based off of
# https://github.com/shiguredo/packer-templates/blob/develop/centos-6.5/scripts/vmware.sh

# setup
yum install -y fuse-libs
mkdir -p /mnt/vmware
mount -o loop /home/vagrant/linux.iso /mnt/vmware

# install
cd /tmp
tar xzf /mnt/vmware/VMwareTools-*.tar.gz
/tmp/vmware-tools-distrib/vmware-install.pl -d

# cleanup
rm -rf /tmp/vmware-tools-distrib
umount /mnt/vmware
rm -rf /home/vagrant/linux.iso

# networking
rm -rf /etc/udev/rules.d/70-persistent-net.rules
sed -i "s/HWADDR=.*//" /etc/sysconfig/network-scripts/ifcfg-eth0
