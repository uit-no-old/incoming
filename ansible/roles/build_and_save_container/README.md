# "Build and save container" ansible role

This role rsyncs a directory with a Dockerfile in it to a remote host, builds a
docker image there, and then saves and downloads that docker image to the
ansible control host.

The role uses the following variables:

* dockerfile\_dir: directory with Dockerfile describing the container you want
  to build, relative from the playbook's directory. Trailing slash is optional.

TODO describe!
