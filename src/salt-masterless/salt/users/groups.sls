#
# salt/groups.sls
# Based on: https://gist.github.com/UtahDave/3785738
#
{% for group, args in pillar['groups'].iteritems() %}
{{group}}:
  group.present:
    - name: {{group}}
{% if 'gid' in args %}
    - gid: {{ args['gid'] }}
{% endif %}
{% endfor %}
