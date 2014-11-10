Installation
============

At present, we don't provide binaries or installers. Compiling and installing Incoming!! from source is easy, though. We also provide [Ansible](http://www.ansible.com/home) playbooks to automate the build and install of Incoming!! along with two example apps and a reverse proxy into [Docker](https://www.docker.com/) containers.

**The example apps as well as the Ansible and Docker content we provide are purely optional, you don't need any of that to get Incoming!! up and running.** However, if you are familiar with Ansible and perhaps Docker, you will quickly be able to set up a complete example system, and our Ansible content and Dockerfiles may serve you well as starting points for your own setup.


Compile and install the Incoming!! server manually
--------------------------------------------------

If you want to use Ansible right away, you may skip this section. We have Ansible goodies for you to compile and install Incoming!!.

In order to compile the Incoming!! server, you need [Go](http://www.golang.org). If you don't have it installed, [install](http://golang.org/doc/install) it. Go's installation manual is a bit long, so here's how you install Go: on Ubuntu, "apt-get golang" will install the compiler and other stuff. Install mercurial and git as well so that Go can install dependencies automatically. Once you installed the packages, set up your "Go path" like this: `mkdir $HOME/go` and then `export GOPATH=$HOME/go`. You might want to put the latter in your `.profile`.

Now you can get and build Incoming!! like this:

    $ go install github.com/USERNAME/incoming

When this is done, the repository has been cloned to `$GOPATH/src/github.com/USERNAME/incoming`. cd there. Then build incoming again, just to have the compiled binary in that directory too (it is already in $GOPATH/bin):

    $ go build

Now you have an executable file called 'incoming' in the directory. That's the Incoming!! server, with Go runtime and all dependencies linked in statically. You can execute the file right there, or on any machine that runs the same OS and architecture. No need to install dependencies.

In order to deploy the Incoming!! server, just copy the executable and the config file [incoming\_cfg.yaml](../incoming_cfg.yaml) to a directory of your choice. Now edit your config file. The first options are the ones you are most likely to edit. Chunk sizes and timeouts and such are performance related options which you are likely only to touch if you optimize the system to your setup. Anyways, all options are explained in the config file, play with them to your heart's content.

When it comes to setting up the Incoming!! server for use, we recommend to run it behind a firewall and to use a reverse proxy that shields all the 'backend' URLs from accesses from the outside. Since Incoming!! uses WebSockets, you might have to specify some special options in your reverse proxy config in order to support WebSocket proxying. Further, it is a good idea to set up the reverse proxy as an SSL endpoint in order to support encrypted connections. For reference, here is the nginx config snippet that we ship with our example web apps (derived from the template in [ansible/roles/incoming\_and\_examples\_on\_one\_host/templates/sites-enabled/example\_apps](../ansible/roles/incoming_and_examples_on_one_host/templates/sites-enabled/example_apps)):

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

First you need to look at the inventory file in [ansible/inventory/default](../ansible/inventory/default). Either edit it, or make your own. What you need is a group called 'build' with one host in it. That host needs the variable `host_build_user`: a user account on that machine in whose home directory we can store stuff. The host must also be root-accessible for Ansible. Further, Docker must be installed on that host, as our Ansible role doesn't take care of that yet (sorry). If the build host is localhost, you might want to make sure that Ansible uses the SSH connection type and not the local connection type, as per the time of this writing Ansible's 'synchronize' module doesn't work properly on 'local' connections.

Another thing you might want to have a look at is the Ansible config file we're using. It's in [ansible/ansible.cfg](../ansible/ansible.cfg). Note the "transport" option, which ties Ansible down to using SSH connections, not local ones, even for localhost.

Unless you want to build Incoming!! *and* the example web apps and reverse proxy (which is described further below), you need to hack the Ansible playbook that makes it all (or copy it and modify the copy). This playbook is in [ansible/build\_and\_run\_incoming\_and\_examples.yml](../ansible/build_and_run_incoming_and_examples.yml). Remove the role invocations for the examples and the example webserver from the 'build' play, and remove the 'test' play entirely. When you run the playbook now, you get a Docker image file stored into the [ansible/docker\_images](../ansible/docker_images) directory. That image can be loaded into a Docker daemon with `docker load`. Tadaa, Incoming!! in a container.

If you want SSH access to Incoming!! Docker containers, copy your public key(s) to the [ansible/authorized\_keys](../ansible/authorized_keys) directory before you run the playbook.

In order to run the Incoming!! server, you need to load the Docker image into the Docker daemon on the machine you want to run the server on, and then just run it. Map port 4000 (Incoming!!'s default port), and optionally port 22 (SSH). Map in a host directory to the container's /var/incoming directory, where uploaded files end up. If you want to see how we do that with Ansible, check the first few tasks in [ansible/roles/incoming\_and\_examples\_on\_one\_host/tasks/main.yml](../ansible/roles/incoming_and_examples_on_one_host/tasks/main.yml). Note that the setup there doesn't map in a specific host directory for /var/incoming but rather accesses that volume from another Docker container running the example web app on the same host.

The Incoming!! server log is in /var/log/incoming.log inside the container. You either have to SSH into the container to check the log, or you have to map in a host directory to /var/log when starting the container if you want to read the log without having to get into the container.

The Incoming!! server is stateless, so we recommend to just discard Docker containers after using them.


Optional: run the example web apps (manually)
---------------------------------------------

If you want to deploy and run everything including the examples automatically with Ansible, skip this section.

The example web apps can run on a machine that shares a filesystem with the Incoming!! server. The directory in which the Incoming!! server stores uploads must have the same name on both machines (for example /var/incoming). The example app and the Incoming!! server must also be able to talk to each other directly. (All of this is given if you run both Incoming!! and examples on the same machine.)

On the machine you run the example app(s) on, you need Python 2 and pip. We recommend virtualenv (get it with "sudo pip install virtualenv"), but it's not necessary. In the following, we describe the install with virtualenv.

    $ cd examples
    examples$ mkdir py-env
    examples$ virtualenv py-env
    examples$ source py-env/bin/activate
    (py-env) examples$ pip install -r pip-req.txt

Now you can run the example backends in whichever shell you have sourced py-env/bin/activate in.

Both backends are started the same way and accept the same command line parameters. There are parameters to specify the Incoming!! server hostname and port, and the app hostname and port. Do the following to get an overview. Then you will know what to do.

    (py-env) examples$ cd 1-simple
    (py-env) 1-simple$ python backend.py --help

When both the Incoming!! server and one of the web apps are running, point your browser to the host:port you have the example web app running on and start playing.


Alternative: build and run the whole example setup automatically with Ansible
-----------------------------------------------------------------------------

The Ansible playbook [ansible/build\_and\_run\_incoming\_and\_examples.yml](../ansible/build_and_run_incoming_and_examples.yml) automates the building and running of an Incoming!! server, one of the two example web apps, and a reverse proxy. At present, it is written having the author's test setup in mind. This is of course rude and will be changed to a Vagrant based setup that works anywhere. But for now you will have to edit the supplied Ansible inventory file in order to run the playbook. Also, you need to have Docker installed already on the machines you use for building and running the Docker containers.

The first steps to build everything are the same as for building only the Incoming!! server with Ansible (see above). You need to edit the inventory file and you might want to have a look at the Ansible config. There is only one difference to the above case when it comes to the inventory: in addition to the 'build' group, you also need a 'test' group, in which you place machines that should (each) run the whole combo of Incoming!! server, example web app, and reverse proxy. The machine(s) you put in that group can of course be the same machine as in the 'build' group.


### Configure the example setup

You can configure the setup you are going to get by modifying the playbook and by adding files into certain directories before running the playbook. In the following, we highight the most likely things you might want to do.

There are two example web apps, but only one web app will be served in our example setup. You can configure which web app that will be in the playbook ([ansible/build\_and\_run\_incoming\_and\_examples.yml](../ansible/build_and_run_incoming_and_examples.yml)). in the play that runs on the 'test' hosts, edit the 'example\_port' variable that is passed to the 'incoming\_and\_examples\_on\_one\_host' role. Set that variable to 4001, and example web app 1 will be served. 4002 will serve example 2.

In the same playbook, you can also configure whether you want to be able to access the running containers with SSH from the outside or not. The 'incoming\_\[...\]\_port\_maps' variables set up these port mappings for the incoming and the example web apps containers, respectively.

The containers only permit SSH logins with key-based authentication, so you need to have your public keys installed in the containers. Just copy the public keys you want to have set up in the containers to the [ansible/authorized\_keys](../ansible/authorized\_keys) and/or the [examples/ansible/authorized\_keys](../examples/ansible/authorized\_keys) directories.

If you don't want to be able to SSH directly into the containers from the outside (you probably don't want to be able to in a production setup later), you can still SSH into them, by going through the host on which the Docker containers run. You SSH to the host, and from there into the containers. To disable 'external' SSH logins, modify (empty) the port forwarding variables in the playbook. To be able to SSH 'internally' into the containers, you probably need a private key on the host, and corresponding public keys in the containers. Put private key file(s) into the [ansible/private\_keys](../ansible/private\_keys) and/or the [examples/ansible/private\_keys](../examples/ansible/private\_keys) directories.

If you want an SSL enabled setup, just add SSL certificates to the [ansible/ssl-certs](../ansible/ssl-certs) directory, and your installation on the corresponding host will support HTTPS. In order for this to work, name your files like this: `<fqdn of host>-nginx.pem` for the certificate file (including possible intermediates), and `<fqdn of host>.key` for the key file. 'fqdn' means 'fully qualified domain name', for example 'my-example-installation.my-company.com'. Note: you might want to run `ansible <host> -i <your inventory file> -m setup` and check the `ansible_fqdn` variable to make sure that Ansible figures out the fqdn correctly.


### Notes on running the example setup

To inspect running containers, log in to them using SSH. Incoming!! and example web apps write logs to /var/log/incoming\_example\_\[12\].log and /var/log/incoming in their respective containers.

Note that if you stop and start either the Incoming!! or the web app example container manually, or if the Docker daemon is restarted for some reason, at least at the time of this writing the internal IP addresses of the containers change and the installation will break because the reverse proxy setup is suddenly wrong (it will answer with "502 - Bad Gateway"). This odd behavior might be fixed in Docker at some point. Until then, the easiest way to fix this problem is to just run the playbook again. (This is not a good solution for a production setup, but we're only doing examples here after all.)


### Notes on hacking all of this

This is roughly what happens when an Incoming!! container is built automatically: first, the Incoming!! source files and some other stuff are copied over to the build host. Then, a Docker image is built there, using a [Dockerfile](../Dockerfile) we provide. During the build, Ansible is installed and executed *inside* the container (check [ansible/inside-docker.yml](../ansible/inside-docker.yml)). The inside-docker.yml playbook installs Go, builds the Incoming!! server, and installs your SSH keys. Then, the Docker image is exported into a tarball, which is downloaded to the Ansible control host, into the [ansible/docker\_images](../ansible/docker_images) directory.

The example web apps container is made using the same process as the Incoming!! server container, but using a different in-container playbook and of course a different project directory. The reverse proxy container is defined by nothing but a Dockerfile, and configuration is later done with config files that are mapped into the container (check [ansible/roles/incoming\_and\_examples\_on\_one\_host/tasks](../ansible/roles/incoming_and_examples_on_one_host/tasks)).

Deploying and running the containers then roughly works like this: copy the Docker image tarballs to the target host, load them into the Docker container there, then configure and start the containers.

All the '.rsync-filter-\*' files that are scattered throughout the source repository are filter definitions that are passed in to rsync. '.rsync-filter' are filters that are always applied, '-incoming' are filters that are only applied when the Incoming!! container is built, and '-examples' are filters that are only applied when the example web apps container is built.



[Main page](../README.md) | [system overview](overview.md) | continue to [examples](examples.md)
