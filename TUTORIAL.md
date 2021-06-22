Turandot Tutorial: Hello World
==============================

*Note that this tutorial is intentionally simplified in that it does not support
CSAR packages that include container images as artifacts. That feature is more complex to
set up and is covered in [a more advanced tutorial](TUTORIAL-SELF-CONTAINED.md).*

Start by cloning the Turandot git repository so that you'll have access to all the example
files and helper scripts:

    git clone https://github.com/tliron/turandot.git
    cd turandot

If you prefer to read code rather than prose, check out the [`lab/`](lab/) directory,
where we have scripts that do much of what is explained here.


Requirements
------------

### Tools

Download the binary release of [Turandot](https://github.com/tliron/turandot/releases).
Specifically you'll need the `turandot` CLI utility (it's a self-contained executable).

We'll also need the `puccini-csar` CLI utility from [Puccini](https://puccini.cloud/)
in order to package our CSAR. It's generally useful to have Puccini available in
order to validate and otherwise work with your TOSCA and CSAR.

A few other tools used by the scripts: `zip`, `zipinfo`, `tar`.

### Kubernetes Cluster

To get started with Turandot you need `kubectl` access to a working Kubernetes
cluster.

For development work, [Minikube](https://minikube.sigs.k8s.io/docs/) is a quick way
to get started. Just make sure to start Minikube with its registry add-on enabled:

    minikube start --addons=registry ...

The `turandot` utility uses the same local configuration you have for `kubectl`, and like
`kubectl` they can accept a `--namespace` argument for selecting the namespace in which to
work. To make commands more succinct in this guide let's set a default namespace:

    kubectl create namespace workspace
    kubectl config set-context --current --namespace=workspace


Installing the Operator
-----------------------

Here we're giving this cluster the "central" site identifier. This will be used
for multi-cluster policy-based TOSCA service composition:

    turandot operator install --site=central --wait -v

Note the operator's container image is downloaded from Docker Hub. Here is a
[direct link](https://hub.docker.com/r/tliron/turandot-operator).

The `--wait` flag tells the command to block until the operator is running
successfully. The `-v` flag adds more verbosity so you can see what the command is
doing. (You can use `-vv` for even more verbosity.)


Building the "Hello World" CSAR
-------------------------------

Let's use the included
[Hello World](https://github.com/tliron/turandot/tree/main/examples/hello-world/) example,
which is based on [this Kubernetes demo](https://github.com/paulbouwer/hello-kubernetes).

You'll use the [build-csar](examples/hello-world/scripts/build-csar) script to package
the TOSCA topology template, profiles, and artifacts into a CSAR:

    examples/hello-world/scripts/build-csar

The CSAR file should now sit in the "dist" directory.


Deploying "Hello World"
-----------------------

You can now deploy the CSAR to your cluster:

    turandot service deploy hello-world --file=dist/hello-world.csar -v

Follow the logs to see what Turandot is doing:

    turandot operator logs --follow

To list all deployed services:

    turandot service list

Note that the "Hello World" example includes a LoadBalancer Kubernetes service that would
allow you to access its web page from outside the cluster. If your cluster is not configured
with LoadBalancer support then, even when successfully
deployed, the service will never get an IP address, and the TOSCA "url"
output for your "Hello World" service will show `http://<unknown>:80`.

If you're using Minikube, it comes with a
[primitive ingress solution](https://minikube.sigs.k8s.io/docs/commands/tunnel/) based on ssh
tunneling that can be useful for testing. To run it (it's blocking, so you might want to do
it in a separate terminal session):

    minikube tunnel

Once the tunnel is up, the LoadBalancer should get its IP address, and Turandot would soon
update the "url" output with the correct URL. You can then use curl or a web browser to access
it:

    xdg-open $(turandot service output hello-world url)
