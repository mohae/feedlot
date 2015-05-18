#
# salt/timezone/init.sls
#
# set timezone

## set default timezone
{{salt['pillar.get']('timezone', 'America/Chicago')}}:
  timezone.system:
    - utc: True