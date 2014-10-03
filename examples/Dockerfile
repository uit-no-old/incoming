FROM phusion/baseimage:0.9.13
MAINTAINER Lars Tiede <lars.tiede@uit.no>

# update base ubuntu
# NOTE this is discouraged by docker: https://docs.docker.com/reference/builder/#run
#   But we do it anyway...
RUN apt-get update; apt-get -y upgrade; apt-get -y dist-upgrade


## some baseimage related things

# make new ssh host keys
ENV HOME /root
RUN /etc/my_init.d/00_regen_ssh_host_keys.sh

# use baseimage's init system
CMD ["/sbin/my_init"]


## my stuff

# install ansible into container (we'll run a 'local playbook' instead of running
# a playbook from another host inside a running and SSH enabled docker container)
RUN apt-get install -y ansible aptitude python-apt python-pip
# TODO: remove python-pip, it's only here to speed up repeated imsge builds
# during development

# copy project diretcory over (note that files in .dockerignore are omitted)
ADD . /tmp/incoming_examples

# run ansible playbook to install dependencies, build and install the example
# apps
RUN cd /tmp/incoming_examples/ansible && ansible-playbook -i localhost, install_inside_docker.yml -c local

# contact points to the outside world: HTTP port
EXPOSE 4001

# finally, clean up
RUN apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*
