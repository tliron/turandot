Helm Example
============

This example demonstrates how a Helm chart can be included as an artifact within a CSAR.


Building the CSAR
-----------------

* [Package the Helm chart](examples/helm/scripts/build-chart)
* [Package as CSAR file](examples/helm/scripts/build-csar)



Deploying
---------
As in the hello-world example, ensure that kubectl is pointing to a k8 service, e.g.

    minikube start --addons=registry ...
    kubectl create namespace workspace
    kubectl config set-context --current --namespace=workspace
    turandot operator install --site=central --role=view --wait -v
    reposure registry create default --provider=minikube --wait -v

then use turandot to deploy the service

    turandot service deploy helm-hello-world --file=dist/helm-hello-world.csar

Note that though Helm is used to create the Kubernetes manifests, Turandot is controlling them.
The `helm` command line tool will not be able to list the release. Use the `turandot service` command
instead.

Verifying
---------
As in the hello-world example, confirm operation using

  Run minikube tunnel in a separate session

    minikube tunnel
  
Then get the external IP address

    kubectl get service hello-world-helm

and use it to confirm operation

    curl <external address>:8080


How the Helm Chart Was Created
------------------------------

We used `helm create hello-world`. Unfortunately, this example chart assumes privileged permissions on
the host that are not allowed out-of-the-box on OpenShift. To ensure it would work on OpenShift we
made two changes:

* In `value.yaml` we changed `image.repository` to `bitnami/nginx`, which is a containerization of NGINX
  that does not require priveleged permissions
  ([documentation](https://hub.docker.com/r/bitnami/nginx)).
* In `templates/deployment.yaml` we changed the port from 80 to 8080 (port 80 requires privileged
  permissions).
