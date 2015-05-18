"""
This enables us to call the minions and search for a specific role
Roles are set using grains (described in http://www.saltstat.es/posts/role-infrastructure.html)
and propagated using salt-mine
"""

import logging

# Import salt libs
import salt.utils
import salt.payload

log = logging.getLogger(__name__)

def get_roles(role, *args, **kwargs):
    """
    Send the informer.is_role command to all minions
    """
    ret = []
    nodes = __salt__['mine.get']('*', 'grains.item')
    print "-------------------------------> NODES {0}".format(nodes)
    for name, node_details in nodes.iteritems():
      name = _realname(name)
      roles = node_details.get('roles', [])
      if role in roles:
        ret.append(name)

    return ret

def get_node_grain_item(name, item):
  """Get the details of a node by the name nodename"""
  name = _realname(name)
  node = __salt__['mine.get'](name, 'grains.item')
  print "NODE DETAILS ------> {0}: {1}".format(name, node[name])
  return node[name][item]

def all():
  """Get all the hosts and their ip addresses"""
  ret = {}
  nodes = __salt__['mine.get']('*', 'grains.item')
  for name, node_details in nodes.iteritems():
    if 'ec2_local-ipv4' in node_details:
      ret[_realname(name)] = node_details['ec2_local-ipv4']
    else:
      ip = __salt__['mine.get'](name, 'network.ip_addrs')[name][0]
      print "-----------------------------> {0}".format(ip)
      ret[_realname(name)] = ip
  return ret
  
def _realname(name):
  """Basically a filter to get the 'real' name of a node"""
  if name == 'master':
    return 'saltmaster'
  else:
    return name