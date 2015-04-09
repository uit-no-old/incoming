## About this role

This role runs an Incoming!! Docker container on a host. An Incoming!! container image should already be in place on the host.

The role needs a couple of variables, and sets some facts when it runs.

## Role variables

* `incoming_image_name`: name of the Incoming!! docker container image (can include tag)
* `incoming_container_name`: name the running Incoming!! docker container should have. Defaults to 'incoming'.
* `incoming_remove_if_exists`: whether to kill and remove the Incoming!! container if it already runs. Defaults to false.
* `incoming_port_forward`: IF:port on target host to map Incoming!!'s port to. Example: `0.0.0.0:4000` makes Incoming!! available at port 4000 on the target host. Defaults to false, which does not expose Incoming!!'s port. In that case, Incoming!! can be reached on port 4000 in Docker's internal network on the target host.
* `incoming_uploads_volume_path`: directory on target host which to map Incoming!!'s upload directory to. Defaults to false, which does not map the directory. In that case, the volume /var/incoming in the Incoming!! container can still be accessed by linking it in another container.

## Facts this role sets

* `incoming_container_facts`: a hash (dictionary) containing the following keys:
    * `internal_ip`: IP of the container in Docker's host-internal network
    * `container_name`: name of the running container (it's the same as `incoming_container_name`)
    * `docker_inspect`: parsed result of docker inspect.
