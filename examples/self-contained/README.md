Self-Contained Example
======================

A stateless single-pod web workload comprising a deployment and a loadbalancer service.

This example demonstrates how a container image can be included as an artifact within a CSAR.


Building the CSAR
-----------------

* [Save container images as tarball](scripts/save-container-image)
* [Package as CSAR file](scripts/build-csar)


Deploying
---------

    turandot service deploy self-contained --file=dist/self-contained.csar

If you want to access the deployed web server from outside the cluster you will need to have
loadbalancing supported on your Kubernetes cluster. On Minikube you can just
[start a tunnel](https://minikube.sigs.k8s.io/docs/handbook/accessing/#using-minikube-tunnel).

If supported, the "url" output of the service template will work. To open from your default web
browser:

    xdg-open $(turandot service output self-contained url)
