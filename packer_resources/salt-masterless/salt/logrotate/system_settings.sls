#
# salt/timezone.sls
#
# set timezone

## set default timezone
{{salt['pillar.get']('timezone_settings:timezone', 'America/Chicago')}}:
  timezone:
    - system
  utc: 
    - {{salt['pillar.get']('timezone_settings:utc', True)}}