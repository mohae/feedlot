#
# salt/users.sls
# based on: https://gist.github.com/UtahDave/3785738
# 
{% for user, args in pillar['users'].iteritems() %}
{{user}}:
  group.present:
    - gid: {{args['gid']}}
  user.present:
    - home: {{args['home']}}
    - shell: {{args['shell']}}
    - uid: {{args['uid']}}
    - gid: {{args['gid']}}
{% if 'password' in args %}
    - password: {{args['password']}}
{% if 'enforce_password' in args %}
    - enforce_password: {{args['enforce_password']}}
{% endif %}
{% endif %}
    - fullname: {{args['fullname']}}
{% if 'groups' in args %}
    - groups: {{args['groups']}}
{% endif %}
    - require:
      - group: {{user}}

{% if 'key.pub' in args and args['key.pub'] == True %}
{{user}}_sshdir:
  file.directory:
    - name: /home/{{user}}/.ssh
    - user: {{user}}
    - group: {{user}}
    - mode: 0700
    - makedirs: True
    - require:
      - user: {{user}}

{{user}}_id_rsa.pub:
  file.managed:
    - user: {{user}}
    - group: {{user}}
    - mode: 0600
    - contents_pillar: users:{{user}}:user.ssh_key
    - name: /home/{{user}}/.ssh/authorized_keys
    - require:
      - file: {{user}}_sshdir
    - watch:
      - user: {{user}}
{% endif %}
{% endfor %}
