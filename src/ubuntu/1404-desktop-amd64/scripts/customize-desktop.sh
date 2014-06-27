#!/usr/bin/env bash
## 13.10 or later
## remove scope stuff
apt-get purge -y unity-scope-calculator unity-scope-chromiumbookmarks unity-scope-clementine unity-scope-colourlovers unity-scope-devhelp unity-scope-firefoxbookmarks unity-scope-gdrive unity-scope-gmusicbrowser unity-scope-guayadeque unity-scope-home unity-scope-manpages unity-scope-musicstores unity-scope-openclipart unity-scope-texdoc unity-scope-tomboy unity-scope-video-remote unity-scope-yelp unity-scope-zotero unity-scopes-master-default unity-lens-applications unity-lens-files unity-lens-friends unity-lens-music unity-lens-video
## add gnome option; aka Do I have to use Unity?...no 
#apt-get install gnome-session-fallback
## or use Cinnamon...there's others ofc.
add-apt-repository -y ppa:tsvetko.tsvetkov/cinnamon
apt-get update -y
apt-get install -y cinnamon
##remove auto reporting
service apport stop ; sed -ibak -e s/^enabled\=1$/enabled\=0/ /etc/default/apport ; mv /etc/default/apportbak ~/ 
apportbak 
gksu getdit /etc/default/appport
## Remove dots from login...because why are they there?
gsettings set com.canonical.unity-greeter draw-grid false
## Remove guest login
echo allow-guest=false | tee -a /etc/lightdm/lightdm.conf.d/50-unity-greeter.conf
## Remove remote login
echo greeter-show-remote-login=false | tee -a /etc/lightdm/lighdm.conf.d/50-unity-greeter.conf

