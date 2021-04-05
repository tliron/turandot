Hello World Example
===================

A stateless single-pod web workload comprising a deployment and a loadbalancer service.


Building the CSAR
-----------------

* [Package as CSAR file](examples/hello-world/scripts/build-csar)


Deploying
---------

    turandot service deploy hello-world --file=dist/hello-world.csar

Verifying
---------

    turandot service list

If you want to access the deployed web server from outside the cluster you will need to have
loadbalancing supported on your Kubernetes cluster. On Minikube you can just
[start a tunnel](https://minikube.sigs.k8s.io/docs/handbook/accessing/#using-minikube-tunnel) in a separate terminal session.

    minikube tunnel

If supported, the "url" output of the service template will show the allocated external address. To open from your default web
browser:

    xdg-open $(turandot service output hello-world url)
