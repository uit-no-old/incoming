Installation
============

At present, we don't provide binaries or installers. Compiling and installing Incoming!! from source is easy, though. We also provide [Ansible](http://www.ansible.com/home) playbooks to automate the build and install of Incoming!! along with two example apps and a reverse proxy into [Docker](https://www.docker.com/) containers.

The example apps as well as the Ansible and Docker content we provide is purely optional, you don't need any of that to get Incoming!! up and running. However, if you are familiar with Ansible and perhaps Docker, you will quickly be able to set up a complete example system, and our Ansible content and Dockerfiles may serve you well as starting points for your own setup.


Compile and install the Incoming!! server manually
--------------------------------------------------

If you want to use Ansible right away, you may skip this section. We have Ansible goodies for you to compile and install Incoming!!.

You need [Go](http://www.golang.org). If you don't have it installed, [install](http://golang.org/doc/install) it. The installation manual is a bit long, so here's how you install Go: on Ubuntu, "apt-get golang" will install the compiler and other stuff. Install mercurial and git as well so that go can install dependencies automatically. Once you installed the packages, set up your "Go path" like this: `mkdir $HOME/go` and then `export GOPATH=$HOME/go`. You might want to put the latter in your `.profile`.

Now you can get and build Incoming!! like this:

    $ go install github.com/USERNAME/incoming

The repository has been cloned to `$GOPATH/src/github.com/USERNAME/incoming`. cd there. Then build incoming again, just to have the binary in that directory:

    $ go build

Now you have an executable file called 'incoming' in the directory. That's the Incoming!! server, with Go runtime and all dependencies linked in statically. You can execute the file right there, or on any machine that runs the same OS and architecture.

In order to deploy the Incoming!! server, just copy the executable and the config file (`incoming_cfg.yaml`) to a directory of your choice. Now edit your config file. The first options are the ones you are most likely to edit. Chunk sizes and timeouts and such are performance related options which you are likely only to touch if you optimize the system to your setup. Anyways, all options are explained in the config file, play with them to your heart's content.

When it comes to setting up the Incoming!! server for use, we recommend to run it behind a firewall and to use a reverse proxy that shields all the 'backend' URLs from accesses from the outside. Since Incoming!! uses WebSockets, you might have to specify some special options in your reverse proxy config in order to support WebSocket proxying. Further, it is a good idea to set up the reverse proxy as an SSL endpoint in order to support encrypted connections. For reference, here is the nginx config snippet that we ship with our example web apps (derived from the template in [`ansible/roles/incoming_and_examples_on_one_host/templates/sites-enabled/example_apps`](../ansible/roles/incoming_and_examples_on_one_host/templates/sites-enabled/example_apps)):

```
# WebSocket magic - when a client requests a "connection upgrade", we
# have to forward that to the WebSocket server (but only then!)

map $http_upgrade $connection_upgrade {
    default upgrade;
    '' close;
}

server {
    server_name <YOUR SERVER NAME>;
    listen 80;

    # SSL config (optional of course, but recommended)

    listen 443 ssl;
    ssl_certificate     <PATH TO CERT FILE>;
    ssl_certificate_key <PATH TO KEY FILE>;
    <MORE SSL CONFIG (sessions, ciphers and such)>

    # Incoming!! URLs all start with /incoming/, which makes it easy to integrate
    # it into a web app - just funnel all requests to /incoming/... to the
    # Incoming!! server.

    location /incoming/ {
        proxy_pass http://<MACHINE INCOMING!! SERVER RUNS ON>:4000;

        # the following options are necessary for WebSocket support
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $connection_upgrade;

        # this one is necessary to let the Incoming!! server know whether the
        # connection between reverse proxy and client is encrypted or not
        proxy_set_header X-Forwarded-Proto $scheme;

        # these ones because we are good citizens
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }

    # The Incoming!! server exposes a few URLs that only your web app backend should
    # be able to access. If only requests "from the outside" go through this reverse
    # proxy, then we can just flat out deny all accesses to the "backend only" URLs.

    location /incoming/backend/ {
        deny all;
    }
}
```

Incoming!! logs accesses and error messages to stdout/stderr. Redirect that to the log file of your choice.


Alternative: compile and install the Incoming!! server automatically with Ansible
---------------------------------------------------------------------------------

If you are familiar with Ansible and Docker, then the provided Ansible roles and plays can help you get an automated install of the Incoming!! server up and running quickly. More importantly, they can serve as reference for your own setup.

First you need to look at the inventory file in [`ansible/inventory/default`](../ansible/inventory/default). Either edit it, or make your own. What you need is a group called 'build' with one host in it. That host needs the variable `host_build_user`: a user account on that machine in whose home directory we can store stuff. The host must also be root-accessible for Ansible. Further, Docker must be installed on that host, as our Ansible role doesn't take care of that yet (sorry). If the build host is localhost, you might want to make sure that Ansible uses the SSH connection type and not the local connection type, as per the time of this writing the 'synchronize' module doesn't work properly on 'local' connections.

Another thing you might want to have a look at is the Ansible config file we're using. It's in [`ansible/ansible.cfg`](../ansible/ansible.cfg). Note the "transport" option, which ties Ansible down to using SSH connections, not local ones, even for localhost.

If you don't want to build Incoming!! *and* the example web apps and reverse proxy, you need to hack the Ansible playbook that makes it all (or copy it and modify the copy). This playbook is in [`ansible/build_and_run_incoming_and_examples.yml`](../ansible/build_and_run_incoming_and_examples.yml). Remove the role invocations for the examples and the example webserver from the 'build' play, and remove the 'test' play entirely. When you run the playbook now, you get a Docker image file stored into the [`ansible/docker_images`](../ansible/docker_images) directory. That image can be loaded into a Docker daemon with `docker load`.

If you want SSH access to Incoming!! Docker containers, copy your public key(s) to the [`ansible/authorized_keys`](../ansible/authorized_keys) directory before you run the playbook.

Feel free to explore and hack the Ansible plays and roles. This is roughly what happens: first, the Incoming!! source files and some other stuff are copied over to the build host. Then, a Docker image is built using a [Dockerfile](../Dockerfile) we provide. During the build, Ansible is installed and executed *within* the container (check [`ansible/inside-docker.yml`](../ansible/inside-docker.yml)). The inside-docker.yml playbook installs Go, builds the Incoming!! server, and installs your SSH keys. Then, the Docker image is exported into a tarball, which is downloaded to the Ansible control host, into the [`ansible/docker_images`](../ansible/docker_images) directory.

In order to run the Incoming!! server, you need to load the Docker image into the Docker daemon on the machine you want to run the server on, and then just run it. Map port 4000, and optionally port 22. Map in a host directory to the container's /var/incoming directory, where uploaded files end up. If you want to see how we do that with Ansible, check the first few tasks in [`ansible/roles/incoming_and_examples_on_one_host/tasks/main.yml`](../ansible/roles/incoming_and_examples_on_one_host/tasks/main.yml). Note that the setup there doesn't map in a host directory for /var/incoming but rather uses that volume from another Docker container on the same host that runs the exaple web app.

The Incoming!! server log is in /var/log/incoming.log inside the container. You either have to SSH into the container to check the log, or you have to map in a host directory to /var/log when starting the container if you want to read the log without having to get into the container.

The Incoming!! server is stateless, so we recommend to just discard Docker containers after using them.


Optional: run the example web apps (manually)
---------------------------------------------

If you want to deploy and run everything including the examples automatically with Ansible, skip this section.

You need Python 2 and pip. We recommend virtualenv (get it with "sudo pip install virtualenv"), but it's not necessary. In the following, we describe the install with virtualenv.


Alternative: build and run everything automatically with Ansible
----------------------------------------------------------------

The Ansible playbook BLABLA automates the building and running of an Incoming!! server, one of the two example web apps, and a reverse proxy. At present, it is written having my work setup in mind. This is rude and will be changed to a Vagrant based setup. But for now you will have to edit the supplied Ansible inventory file in order to run the playbook. Also, on the "build" machine, you need to have Docker installed.

EDIT INVENTORY

EDIT PLAYBOOK TO SELECT WHICH EXAMPLE TO RUN

IF YOU HAVE SSL CERTS, PUT THEM IN THE RIGHT PLACE AND MAGIC WILL HAPPEN

INSPECT CONTAINERS
