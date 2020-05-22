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
