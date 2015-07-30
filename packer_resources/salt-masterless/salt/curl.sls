#
# salt/httpd/curl.sls
#
# install curl
#
curl:
  pkg: 
    - installed
    - name: curl
