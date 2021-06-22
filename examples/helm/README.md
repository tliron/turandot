Helm Example
============

This example demonstrates a Helm chart included as an artifact within a CSAR.

To deploy this example you can generally follow the instructions for the
[Hello World example](../../TUTORIAL.md), though make sure to build the chart
artifact first. E.g.:

    examples/scripts/build-chart
    examples/scripts/build-csar
    turandot service deploy helm --file=dist/helm.csar

As long as you have LoadBalancer ingress support on your cluster (such as
Minikube's "tunnel"), then you can then use curl or a web browser to access the
deployed service:

    IP=$(kubectl get services --selector=app.kubernetes.io/name=helm --output=jsonpath={.items[0].status.loadBalancer.ingress[0].ip})
    xdg-open http://$IP:8080


Helper Scripts
--------------

* [Package the Helm chart](scripts/build-chart)
* [Package as CSAR file](scripts/build-csar)


How the Helm Chart Was Created
------------------------------

We used `helm create hello-world`. Unfortunately, this example chart assumes privileged containers on
the host that are not allowed out-of-the-box on OpenShift. To ensure it would work on OpenShift we
made two changes:

* In `value.yaml` we changed `image.repository` to `bitnami/nginx`, which points to an NGINX container
  image that does not require a priveleged container
  ([documentation](https://hub.docker.com/r/bitnami/nginx)).
* In `templates/deployment.yaml` we changed the port from 80 to 8080 (again, because port 80 requires
  a privileged container).
