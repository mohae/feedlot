#!/usr/bin/env bash
## 13.10 or later
## remove scope stuff
apt-get purge unity-scope-calculator unity-scope-chromiumbookmarks unity-scope-clementine unity-scope-colourlovers unity-scope-devhelp unity-scope-firefoxbookmarks unity-scope-gdrive unity-scope-gmusicbrowser unity-scope-guayadeque unity-scope-home unity-scope-manpages unity-scope-musicstores unity-scope-openclipart unity-scope-texdoc unity-scope-tomboy unity-scope-video-remote unity-scope-yelp unity-scope-zotero unity-scopes-master-default unity-lens-applications unity-lens-files unity-lens-friends unity-lens-music unity-lens-video
## add gnome option; aka Do I have to use Unity?...no 
#apt-get install gnome-session-fallback
## or use Cinnamon...there's others ofc.
add-apt-repository ppa:gwendal-lebihan-dev/cinnamon-stable
apt-get update
apt-get install cinnamon
##remove auto reporting
service apport stop ; sudo sed -ibak -e s/^enabled\=1$/enabled\=0/ /etc/default/apport ; sudo mv /etc/default/apportbak 

