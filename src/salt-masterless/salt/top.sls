## saltbase/top.sls
##   saltbase is used for a basic server installation. No webserver, db, etc.
##   is installed. This is used as the basic repo to use for other 
##   salt configurations.
## 
##   Supported environments:
##
##   set base tree state: calls base.sls for most of the work

base:
  '*':
    - curl
    - date
    - git
    - hosts
    - iptables
    - locale
    - logrotate
    - ntp
    - openssh  
    - sudo
    - timezone
    - users
    - vim
