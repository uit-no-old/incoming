Installation
============

At present, we don't provide binaries or installers. Compiling and installing Incoming!! from source is easy, though. We also provide a [Vagrant](http://www.vagrantup.com)file along with [Ansible](http://www.ansible.com/home) playbooks to automate the build and install of Incoming!! along with two example apps and a reverse proxy into [Docker](https://www.docker.com/) containers.

**The example apps as well as the Vagrant, Ansible, and Docker content we provide are purely optional, you don't need any of that to get Incoming!! up and running.** However, if you are familiar with either Vagrant, Ansible, or Docker, you will quickly be able to set up a complete example system, and especially our Ansible content and Dockerfiles may serve you well as starting points for your own setup.


Optional: build an example system automatically with Vagrant and Ansible
------------------------------------------------------------------------

We provide a Vagrantfile so that you can quickly build a complete example setup including a web app and reverse proxy into a Vagrant VM. If you don't want to try the examples on a preconfigured setup, you can skip this section.

As prerequisites, you need to have Vagrant and Ansible installed. Then, make an empty directory, cd to it, and clone the Incoming!! repository:

    $ git clone TODO paste repo URL here .

Then, let Vagrant start and Ansible provision a VM with everything for you:

    $ vagrant up

This will take a looong time (~30 minutes on my laptop). If something fails along the way due to some server we fetch dependencies from not responding (this happened to me once), try again like this:

    $ vagrant provision

Finally, if everything ran through nicely, you can point your browser to <http://10.20.1.4> and start playing with Incoming!!.

The bits and pieces that make this setup (Ansible roles and playbooks, Dockerfiles) are worth exploring when you go about setting up your own Incoming!! installation. For details on this, check [doc/installation\_ansible.md](doc/installation_ansible.md).


Compile and install the Incoming!! server manually
--------------------------------------------------

You can ignore all the Vagrant, Ansible, and Docker magic, and just build the Incoming!! server, set it up yourself, and get going. Here is how.

In order to compile the Incoming!! server, you need [Go](http://www.golang.org). If you don't have it installed, [install](http://golang.org/doc/install) it. Go's installation manual is a bit long, so here's how you install Go: on Ubuntu, "apt-get golang" will install the compiler and other stuff. Install mercurial and git as well so that Go can install dependencies automatically. Once you installed the packages, set up your "Go path" like this: `mkdir $HOME/go` and then `export GOPATH=$HOME/go`. You might want to put the latter in your `.profile`.

Now you can get and build Incoming!! like this:

    $ go install github.com/USERNAME/incoming

When this is done, the repository has been cloned to `$GOPATH/src/github.com/USERNAME/incoming`. cd there. Then build incoming again, just to have the compiled binary in that directory too (it is already in $GOPATH/bin):

    $ go build

Now you have an executable file called 'incoming' in the directory. That's the Incoming!! server, with Go runtime and all dependencies linked in statically. You can execute the file right there, or on any machine that runs the same OS and architecture. No need to install dependencies.

In order to deploy the Incoming!! server, just copy the executable and the config file [incoming\_cfg.yaml](../incoming_cfg.yaml) to a directory of your choice. Now edit your config file. The first options are the ones you are most likely to edit. Chunk sizes and timeouts and such are performance related options which you are likely only to touch if you optimize the system to your setup. Anyways, all options are explained in the config file, play with them to your heart's content.

When it comes to setting up the Incoming!! server for use, we recommend to run it behind a firewall and to use a reverse proxy that shields all the 'backend' URLs from accesses from the outside. Since Incoming!! uses WebSockets, you might have to specify some special options in your reverse proxy config in order to support WebSocket proxying. Further, it is a good idea to set up the reverse proxy as an SSL endpoint in order to support encrypted connections. Encryption should be available for all 'world-facing' connections, and also for the connections between web app backends and the Incoming!! server if the network between them can't be trusted.

For reference, here is the nginx config snippet that we ship with our example web apps (derived from the template in [ansible/roles/incoming\_and\_examples\_on\_one\_host/templates/sites-enabled/example\_apps](../ansible/roles/incoming_and_examples_on_one_host/templates/sites-enabled/example_apps)):

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

    location /incoming/0.1/ {
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

    location /incoming/0.1/backend/ {
        deny all;
    }
}
```

Incoming!! logs accesses and error messages to stdout/stderr. Redirect that to the log file of your choice.


Optional: run the example web apps manually
-------------------------------------------

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


[Main page](../README.md) | [system overview](overview.md) | continue to [examples](examples.md)
