Hello World
===========

A stateless single-pod web workload comprising a deployment and a loadbalancer service.

This example demonstrates how a container image can be included as an artifact within a CSAR.


Building the CSAR
-----------------

* [Save container images as tarball](scripts/save-container-image)
* [Package as CSAR file](scripts/build-csar)


Deploying
---------

    turandot service deploy hello-world --file=dist/hello-world.csar

If you want to access the deployed web server from outside the cluster you will need to have
loadbalancing supported on your Kubernetes cluster. On Minikube you can just
[start a tunnel](https://minikube.sigs.k8s.io/docs/handbook/accessing/#using-minikube-tunnel).

[Here](scripts/url) is a script to get the URL of the web server. To open the page on your default
web browser in one command:

    xdg-open $(examples/hello-world/scripts/url)
