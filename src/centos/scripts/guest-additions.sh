#!/usr/bin/env bash
# Install VirtualBox Guest Additions - Server style
# Install prelims
yum --enablerepo rpmforge install dkms
yum group install "Development Tools"
yum install kernel-devel
mkdir /media/cdrom
mount /dev/scd0  /media/cdrom
sh /media/VBOXADDITIONS_4.3.6_r91406/VBoxLinuxAdditions.run

#NOTE# above not tested and unsure about the mounting of guest additions.
#    I'm pretty sure about the instal stuff
