#!/usr/bin/env bash
<<<<<<< HEAD
=======
apt-get -y install deborphan
>>>>>>> c5c1889d741c448ab595f6ebdcddfd5acc59b2d8
deborphan | xargs sudo apt-get purge -y
apt-get clean
apt-get autoclean
apt-get autoremove
apt-get purge
