<esc><wait>
<esc><wait>
<enter><wait>
/install/vmlinuz<wait>
 auto<wait>
 console-setup/ask_detect=false<wait>
 console-setup/layoutcode=us<wait>
 console-setup/modelcode=pc105<wait>
 debconf/frontend=noninteractive<wait>
 debian-installer=en_US<wait>
 fb=false<wait>
 initrd=/install/initrd.gz<wait>
 kbd-chooser/method=us<wait>
 keyboard-configuration/layout=USA<wait>
 keyboard-configuration/variant=USA<wait>
 locale=en_US<wait>
 netcfg/get_hostname=vagrant<wait>
 netcfg/get_domain=vagrantup.com<wait>
 noapic<wait>
 preseed/url=http://{{ .HTTPIP }}:{{ .HTTPPort }}/preseed.cfg<wait>
 -- <wait>
<enter><wait>
