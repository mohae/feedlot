<esc><wait>
install <wait>
preseed/url=http://{{ .HTTPIP }}:{{ .HTTPPort }}/preseed.cfg <wait>
debian-installer=en_US <wait>
auto <wait>
locale=en_US.UTF-8 <wait>
netcfg/get_hostname={{ .Name }} <wait>
netcfg/get_domain=vagrantup.com <wait>
fb=false <wait>
debconf/frontend=noninteractive <wait>
console-setup/ask_detect=false <wait>
console-keymaps-at/keymap=us <wait>
keyboard-configuration/xkb-keymap=ys <wait>
<enter><wait>
