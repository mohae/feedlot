#!/usr/bin/env bash
yum install -y yum-utils
yum erase -y gtk2 libX11 hicolor-icon-theme avahi freetype bitstream-vera-fonts
yum erase -y $(package-cleanup --leaves)
yum erase -y yum-utils
yum clean -y all
rm -rf VBoxGuestAdditions_*.iso
rm -rf /tmp/rubygems-*
