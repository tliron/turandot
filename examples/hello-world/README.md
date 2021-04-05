Hello World Example
===================

A stateless single-pod web workload comprising a deployment and a loadbalancer service.

This example demonstrates how a container image can be included as an artifact within a CSAR.


Building the CSAR
-----------------

* [Save container images as tarball](examples/hello-world/scripts/save-container-image)
* [Package as CSAR file](examples/hello-world/scripts/build-csar)


Deploying
---------

    turandot service deploy hello-world --file=dist/hello-world.csar

Verifying
---------

The command

    turandot service list

will show 'mode normal hellow-world: failed' until an external address is allocated.

If you want to access the deployed web server from outside the cluster you will need to have
loadbalancing supported on your Kubernetes cluster. On Minikube you can just
[start a tunnel](https://minikube.sigs.k8s.io/docs/handbook/accessing/#using-minikube-tunnel) in a separate terminal session.

As in the hello-world example, confirm operation using

  Run minikube tunnel in a separate session

    minikube tunnel

If supported, the "url" output of the service template will show the allocated external address. To open from your default web
browser:

    xdg-open $(turandot service output hello-world url)
