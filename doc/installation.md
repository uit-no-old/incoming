Installation
============

At present, we don't provide binaries or installers. Compiling and installing Incoming!! from source is easy, though. We also provide [Ansible](http://www.ansible.com/home) playbooks to automate the build and install of Incoming!! along with two example apps and a reverse proxy into [Docker](https://www.docker.com/) containers. The example apps as well as the Ansible and Docker content we provide is purely optional, you don't need any of it to get Incoming!! up and running. However, if you are familiar with Ansible and perhaps Docker, you will quickly be able to set up a complete example system, and our Ansible content and Dockerfiles may serve you well as starting points for your own setup.


Compile and install the Incoming!! server manually
--------------------------------------------------

If you want to use Ansible right away, skip this section. We have Ansible goodies for you to get you started.

You need [Go](http://www.golang.org). Then go fetch and go build, and you have the executable. Edit the config file, deploy the executable and config file in one directpry wherever you want. We recommend to use a reverse proxy that shields all the backend-internal things. Example nginx setup snippet.TODO elaborate all of this


Compile and install the Incoming!! server automatically with Ansible
--------------------------------------------------------------------

The Ansible roles and playbooks that are included in the source repository include the automatic build of a Docker container in which the Incoming!! server is being built and deployed. Check TODO for details.


Run the example web apps manually
---------------------------------

If you want to deploy and run everything including the examples automatically with Ansible, skip this section.

You need Python 2 and pip. We recommend virtualenv (get it with "sudo pip install virtualenv"), but it's not necessary. In the following, we describe the install with virtualenv.


Run the example web apps automatically with Ansible
---------------------------------------------------

The Ansible playbook BLABLA automates the building and running of an Incoming!! server, one of the two example web apps, and a reverse proxy. At present, it is written having my work setup in mind. This is rude and will be changed to a Vagrant based setup. But for now you will have to edit the supplied Ansible inventory file in order to run the playbook. Also, on the "build" machine, you need to have Docker installed.

EDIT INVENTORY

EDIT PLAYBOOK TO SELECT WHICH EXAMPLE TO RUN

IF YOU HAVE SSL CERTS, PUT THEM IN THE RIGHT PLACE AND MAGIC WILL HAPPEN

INSPECT CONTAINERS
