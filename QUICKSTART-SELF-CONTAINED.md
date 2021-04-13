Turandot Quickstart: Self-Contained
===================================

*Note that this quickstart guide is complex in that it supports CSAR packages that include
container images as artifacts. You may prefer to start with the
[simpler quickstart guide](QUICKSTART.md).*

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

You will also need the [Reposure](https://reposure.puccini.cloud/) CLI utility in
order to configure the registry that Turandot will use to store TOSCA service templates
and artifacts, so download that, too.

Finally, we'll need the `puccini-csar` CLI utility from [Puccini](https://puccini.cloud/)
in order to package our CSAR. It's generally useful to have Puccini available in
order to validate and otherwise work with your TOSCA and CSAR.

A few other tools used by the scripts: `podman` (or `docker`), `pigz` (or `gzip`),
`zip`, `zipinfo`, `tar`, `jq`.

### Kubernetes Cluster

To get started with Turandot you need `kubectl` access to a working Kubernetes
cluster.

For development work, [Minikube](https://minikube.sigs.k8s.io/docs/) is a quick way
to get started. Just make sure to start Minikube with its registry add-on enabled:

    minikube start --addons=registry ...

The `turandot` and `reposure` utilities use the same local configuration you have for
`kubectl`, and like `kubectl` they can accept a `--namespace` argument for selecting
the namespace in which to work. To make commands more succinct in this guide let's set a
default namespace:

    kubectl create namespace workspace
    kubectl config set-context --current --namespace=workspace


Installing the Operators
------------------------

Before you install the operators, note that you may need to specify a cluster-wide role
in order to permit them access to the authentication and authorization secrets necessary
to connect to the registry (see below), because those secrets may be stored outside of
your namespace.

If you indeed need access to secrets for testing you can use "view" role, which already
exists in most Kubernetes deployments. It allows read-access to any resource in any
namespace.

(For production systems you would want more tightened security, in which case it would
be a good idea to create a custom cluster role that would allow access to just the
registry secrets and nothing else. Doing so is beyond the scope of this guide.)

To install the operators with the "view" cluster role you can use these commands:

    turandot operator install --role=view --site=central --wait -v
    reposure operator install --role=view --wait -v

Here we're also giving this cluster the "central" site identifier. This will be used
for multi-cluster policy-based TOSCA service composition.

Note the operators' container images are downloaded from Docker Hub. Here are
direct links for [Turandot](https://hub.docker.com/r/tliron/turandot-operator),
[Reposure operator](https://hub.docker.com/r/tliron/reposure-operator), and
[Reposure surrogate](https://hub.docker.com/r/tliron/reposure-surrogate).

The `--wait` flag tells the commands to block until the operators are running
successfully. The `-v` flag adds more verbosity so you can see what the command is
doing. (You can use `-vv` for even more verbosity.)


Configuring the Registry
------------------------

You will now use Reposure to configure the "default" registry for Turandot. 

This can be simple or complex depending on your Kubernetes cluster. The reason it can be
challenging is that the Turandot operator does more than just deploy TOSCA, it can also deploy
artifacts referred to by your TOSCA, including artifacts of a special type: container images.
Container images are downloaded from a registry by the container runtime (CRI-O, Docker, etc.)
that runs on each of the cluster's hosts, and by the container runtime is likely to be configured
by delay to require TLS authentication (HTTPS) and may even require authorization.

Reposure comes with built-in support for the built-in registries of a few Kubernetes
distributions, making it easy to make use of them. For Minikube:

    reposure registry create default --provider=minikube --wait -v

For OpenShift:

    reposure registry create default --provider=openshift --wait -v

(For both of these cases we would need to add `--role=view` during the operator
installation step above.)

If you're using neither Minikube nor OpenShift then you must set up your own registry.
For production environments you'd likely want to use a robust product, like
[Harbor](https://goharbor.io/) or [Quay](https://www.projectquay.io/).
Alternatively, Reposure comes with a "simple" registry that is suitable for testing and
small deployments.

### Reposure's "Simple" Registry

Installing the "simple" registry is simple, but configuring your Kubernetes container
runtime to accept it is beyond the scope of this guide. Specifically you would need to
allow it to accept your TLS certificate or your custom certificate authority.

However, if you can configure your container runtime to at least accept self-signed
certificates (so-called "insecure" mode, which in Minikube is enabled via the
[`--insecure-registry`](https://minikube.sigs.k8s.io/docs/handbook/registry/) flag),
then Reposure's "simple" registry can provision such a self-signed certificate for
you.

To do so Reposure relies on [cert-manager](https://github.com/jetstack/cert-manager),
which does a lot of the heavy lifting required for provisioning and updating
certificates. (The additional challenge of working with TLS certificates in cloud
environments is that IP addresses change, so that certificates either have to be
updated or tied to a DNS domain name, and then DNS management may be local and custom.)

So, assuming your container runtime is "insecure", you can start by installing
cert-manager:

    kubectl apply --filename=https://github.com/jetstack/cert-manager/releases/download/v1.3.1/cert-manager.yaml

And then install the "simple" registry with self-signed authentication like so:

    reposure simple install --authentication --wait -v

Finally, configure the registry:

    reposure registry create default --provider=simple --wait -v

(Note that if you are using the "simple" registry with authentication then you don't
need to install the operators with `--role=view`, because the "simple" registry stores
its certificate secrets within its namespace.)


Building the Self-Contained CSAR
--------------------------------

Let's use the included [Self-Contained](examples/self-contained/) example, which is based on
[this Kubernetes demo](https://github.com/paulbouwer/hello-kubernetes).

First you'll need to export the container image into a tarball so that you can use it as
a TOSCA artifact. The
[build-csar](examples/self-contained/scripts/save-container-image) script will handle that:

    examples/self-contained/scripts/save-container-image

Next you'll use the [build-csar](examples/self-contained/scripts/build-csar) script to package
the TOSCA topology template, profiles, and artifacts (including the container image
tarball you created above) into a CSAR:

    examples/self-contained/scripts/build-csar

The CSAR file should now sit in the "dist" directory.


Deploying "Self-Contained"
--------------------------

You can now deploy the CSAR to your cluster:

    turandot service deploy self-contained --file=dist/self-contained.csar -v

Follow the logs to see what Turandot is doing:

    turandot operator logs --follow

To list all deployed services:

    turandot service list

Note that the "Self-Contained" example includes a LoadBalancer Kubernetes service that would
allow you to access its web page from outside the cluster. If your cluster is not configured
with LoadBalancer support then the service will never get an IP address, and the TOSCA "url"
output for your "Self-Contained" service will show `http://<unknown>:80`, even when successfully
deployed.

If you're using Minikube, it comes with a
[primitive ingress solution](https://minikube.sigs.k8s.io/docs/commands/tunnel/) based on ssh
tunneling that can be useful for testing. To run it (it's blocking, so you might want to do
it in a separate terminal session):

    minikube tunnel

Once the tunnel is up, the LoadBalancer should get its IP address, and Turandot would soon
update the "url" output with the correct URL. You can then use curl or a web browser to access
it:

    xdg-open $(turandot service output self-contained url)
