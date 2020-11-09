Helm Example
============

This example demonstrates how a Helm chart can be included as an artifact within a CSAR.


Building the CSAR
-----------------

* [Package the Helm chart](scripts/build-chart)
* [Package as CSAR file](scripts/build-csar)


Deploying
---------

    turandot service deploy helm-hello-world --file=dist/helm-hello-world.csar

Note that though Helm is used to create the Kubernetes manifests, Turandot is controlling them.
The `helm` command line tool will not be able to list the release. Use the `turandot service` command
instead.


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
