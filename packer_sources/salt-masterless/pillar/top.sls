## top.sls
##  define the Salt States 

base:              
  '*':
    - users
    - groups
    - timezone
    - vim