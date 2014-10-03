#!/bin/bash
# {{ ansible_managed }}

cd "/home/{{ host_example_user }}/incoming_examples/{{ item.directory }}"
source ../py-env/bin/activate
exec {{ item.invocation }} >> /var/log/{{ item.name }}.log 2>&1

# from baseimage documentation, for reference:
# `/sbin/setuser memcache` runs the given command as the user `memcache`.
# If you omit that part, the command will be run as root.
#exec /sbin/setuser memcache /usr/bin/memcached >>/var/log/memcached.log 2>&1
