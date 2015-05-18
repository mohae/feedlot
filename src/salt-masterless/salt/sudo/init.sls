#
# salt/sudo/init.sls
#

sudo:
  pkg:
    - installed
  service:
    - running
    - require:
      - pkg: sudo

sudoers:
  file:
    - managed
    - name: /etc/sudoers
    - user: root
    - group: root
    - mode: 440
    - source: salt://sudo/sudoers
    - require:
      - pkg: sudo
