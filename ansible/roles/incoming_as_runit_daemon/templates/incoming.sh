#!/bin/sh
# {{ ansible_managed }}

cd "{{ incoming_source_dir }}" && exec ./incoming >> /var/log/incoming.log 2>&1

# from baseimage documentation, for reference:
# `/sbin/setuser memcache` runs the given command as the user `memcache`.
# If you omit that part, the command will be run as root.
#exec /sbin/setuser memcache /usr/bin/memcached >>/var/log/memcached.log 2>&1
