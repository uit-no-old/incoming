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
#
# copy project diretcory over (note that files in .dockerignore are omitted)
ADD . /incoming_examples/

# contact points to the outside world: HTTP port
EXPOSE 4001

# finally, clean up
RUN apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*
