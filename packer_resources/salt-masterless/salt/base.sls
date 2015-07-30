include:
  - hosts
  - iptables
  - users
  - vim
  - date
  - sudo
  - ssh.server
  - files
  - python
  - python.python-pip
{% if grains['os'] == 'Debian' %}
  - python.python-apt
{% endif %}

# set default timezone
{% system_settings = pillar.get('system_settings', {}) %}
{% timezone_setting = system_settings['timezone'] %}
{% utc_setting = system_settings['utc'] %}

{{timezone_setting}}:
  timezone.system:
  utc: {{utc_setting}}

#default_locale:
#  
#en_US.UTF-8:
#  locale.system

psmisc:
  pkg:
    - installed
