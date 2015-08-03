{# localhost_present_host:}
#   host.present:
#     - ip: 127.0.0.1
#     - name: {{ grains['id'] }} #}
local_address_host:
  host.absent:
    - ip: 127.0.1.1

{% if grains['ec2_local-ipv4'] is defined -%}

{% for name, ip in salt['informer.all']().iteritems() %}

# {{ name }}
{#% if name != grains['id'] -#}
cleanup_{{ name }}_internal_host:
  cmd.run:
    - name: "grep -v {{ name }} /etc/hosts > /etc/hosts.new; mv /etc/hosts.new /etc/hosts"
    - unless: "CURR_IP=$(grep -m 1 {{ name }} /etc/hosts | awk '{print $1}') && [[ \"$CURR_IP\" == \"{{ ip }}\" ]]"
    - user: root

{{ name }}_internal_host:
  cmd.wait:
    - name: "echo '{{ ip }}    {{ name }}' >> /etc/hosts"
    - unless: grep {{ name }} /etc/hosts
    - user: root
    - watch:
      - cmd: cleanup_{{ name }}_internal_host
{#% endif -#}

{% endfor -%}
{% else -%}

{% for name, ip in salt['informer.all']().iteritems() %}

{#% if name != grains['id'] -#}
cleanup_{{ name }}_internal_host:
  cmd.run:
    - name: "grep -v {{ name }} /etc/hosts > /etc/hosts.new; mv /etc/hosts.new /etc/hosts"
    - unless: "CURR_IP=$(grep -m 1 {{ name }} /etc/hosts | awk '{print $1}') && [[ \"$CURR_IP\" != \"{{ ip }}\" ]]"
    - user: root

{{ name }}_internal_host:
  cmd.wait:
    - name: "echo '{{ ip }}    {{ name }}' >> /etc/hosts"
    - unless: grep {{ name }} /etc/hosts
    - user: root
    - watch:
      - cmd: cleanup_{{ name }}_internal_host
{#% endif -#}

{% endfor -%}
{% endif %}
