Hello World Example
===================

A stateless single-pod web workload comprising a deployment and a loadbalancer service.

Requirements
------------
Ensure to met all the requirements as explained in [QUICKSTART](../QUICKSTART.md), for minikube these can be summarized as: 

    minikube start --addons=registry ...
    kubectl create namespace workspace
    kubectl config set-context --current --namespace=workspace
    turandot operator install --site=central --role=view --wait -v
    reposure registry create default --provider=minikube --wait -v

Building the CSAR
-----------------

* [Package as CSAR file](scripts/build-csar)

    examples/hello-world/scripts/build-csar

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
