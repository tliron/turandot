*This is an early release. There are still important features missing.*

Turandot
========

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Latest Release](https://img.shields.io/github/release/tliron/turandot.svg)](https://github.com/tliron/turandot/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/tliron/turandot)](https://goreportcard.com/report/github.com/tliron/turandot)

Orchestrate and compose [Kubernetes](https://kubernetes.io/) workloads using
[TOSCA](https://www.oasis-open.org/committees/tosca/).

Supports policy-based service composition based on service templates stored in an inventory. 

Workloads can comprise both standard and
[custom](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
Kubernetes resources, as well as their
[operators](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/). They can be deployed
on a single cluster or on multi-cluster clouds. Virtual machines are supported via
[KubeVirt](https://kubevirt.io/).

[Helm](https://helm.sh/) charts and external orchestrators, such as
[Ansible](https://www.ansible.com/), are supported via custom artifacts encapsulated as TOSCA
types.

See the included [examples](examples/).

Turandot targets complex, large-scale workloads. Moreover, it intends to handle the
orchestration aspect of
[NFV (Network Function Virtualization) MANO (Management and Orchestration)](https://en.wikipedia.org/wiki/Network_function_virtualization#Management_and_orchestration_%28MANO%29),
which is a crucial component for deploying heterogeneous network services on clouds at scale.
Included is a comprehensive example of a multi-cluster
[telephony network service](examples/telephony-network-service/) modeled entirely in TOSCA.

Get It
------

[![Download](assets/media/download.png "Download")](https://github.com/tliron/turandot/releases)


Rationale
---------

**Design-time**: TOSCA's extensibility via an object-oriented grammar is analogous to Kubernetes's
extensibility via custom resource definitions and operators. TOSCA's added value is in providing a
composable and validated graph of resource interrelations, effectively imbuing Kubernetes resources
with architectural intent.

**Run-time**: Turandot manages resources *together* as single, coherent workloads—whether we call
them "applications" or "services"—even across cluster boundaries, ensuring consistency and
integration as well as allowing for cascading policies for allocation, composition, networking,
security, etc.


How It Works
------------

The core is a Kubernetes operator that:

1. Can work with an internal (built-in) or external inventories to retrieve CSAR-packaged service
   templates. A CSAR
   ([Cloud Service Archive](https://docs.oasis-open.org/tosca/TOSCA-Simple-Profile-YAML/v1.3/os/TOSCA-Simple-Profile-YAML-v1.3-os.html#_Toc26969474)) is a zip file containing a TOSCA service template, TOSCA profiles,
   and other files ("artifacts") required for orchestration (see #4, below).
2. Uses [Puccini](https://puccini.cloud/) to compile the CSAR-packaged service templates into the
   [Clout](https://puccini.cloud/clout/) intermediary format.
3. Renders the Clout to Kubernetes resources and schedules them as integrated workloads.
4. Deploys and activates artifacts packaged in the CSAR file. This includes container images (as
   well as KubeVirt virtual machine images) and cloud-native configuration tools, such as scripts,
   playbooks, recipes, etc., as well as Kubernetes operators. These configuration tools have access
   to the entire workload topology, allowing them to essentially configure themselves.
5. Can delegate orchestration to Turandot operators in remote clusters (see multi-cluster workloads,
   below).

The Turandot operator can be controlled using the `turandot` utility, e.g.:

    turandot service deploy my-service --file=my-service-template.csar

⮕ [Documentation](turandot/)


The Cycle of Life
-----------------

**Day -1: Modeling.** TOSCA is used to create "profiles" of reusable, composable types, which
together provide a validated and validating model for the target domain. TOSCA profiles vastly
simplify the work of the service template designer. For example, our telephony network service
example uses profiles for Kubernetes, KubeVirt, network services (including data planes), and
telephony.

**Day 0: Design.** Solution architects compose the models provided by the TOSCA profiles into
service templates, either by writing the TOSCA manually, or by using a wonderful graphical TOSCA IDE
(that is yet to be created!). The templates are tested in lab and staging environments using
CI/CD-style automation.

**Day 1: Operations Handoff.** The service templates are ready to be instantiated in production.
A ticket from an operations support system (OSS) initiates the transfer to a managed multi-cluster
cloud. Turandot is installed on the target clusters (or available as a delegate from central
clusters) and takes it from there.

**Day 2+: Cloud-native Operations.** Once they are up and running the services should orchestrate
themselves by adapting to changing internal and external conditions, as well as triggered and manual
actions from operations. Changes include scaling, healing, migration, as well as more elaborate
transformations. The Turandot operator will continue to monitor these changes and update the Clout.
Components can refer to the Clout as "single source of truth" to see the complete topology in order
to make self-orchestration decisions, as well as checking against policies to which they must or can
adhere. Machine learning and AI can be applied to the Clout in order to make the best possible
runtime orchestration decisions.


Multi-Cluster Workloads
-----------------------

What if your workload crosses the boundaries of a single Kubernetes cluster?

Each cluster will have its own Turandot operator that manages resources *only* for that cluster,
however the Clout will always contain a view of *all* resources, ensuring workload integration.
Each operator can delegate work to specific other operators, according to composition policy.
This network of operators essentially turns your multi-cluster environment into a single cloud.

Note that allowing operators to network with each other across cluster boundaries is beyond the
scope of Turandot, however you can definitely use Turandot to orchestrate this control plane itself.
Often this will be a [SDN solution](https://en.wikipedia.org/wiki/Software-defined_networking), such
as shared virtual LANs across SD-WAN connections, using a combination of
[Kubernetes CNI providers](https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/network-plugins/),
[Multus](https://github.com/intel/multus-cni), [Cilium](https://cilium.io/),
[Network Service Mesh](https://networkservicemesh.io/), custom proxies, etc. Indeed, one size does
not fit all, which is why Turandot insists on not having an opinion.


Namespaced or Cluster Mode
--------------------------

The Turandot operator can work in either "namespaced mode", in which it can only manage resources in
the namespace in which it is installed, or "cluster mode", in which it can manage all namespaces.

Cluster mode requires elevated permissions, and as such may not be applicable in multi-tenancy
scenarios. A more secure configuration is to have Turandot installed only in supported namespaces
within a cluster and to allow secure delegation between them, in effect treating it like the
multi-cluster scenario (see above).


FAQ
---

### Is this a lifecycle manager (LCM) for Kubernetes workloads?

No, or not exactly. In Kubernetes, LCM is hardcoded behind the scheduling paradigm. Of course work
is done by built-in and custom controllers to provision containers, wire up the networking, run init
containers and sidecars, attach storage blocks, etc., but from an orchestration perspective LCM is
largely reduced to a simple binary: either the resource is scheduled or it isn't.

Individual resources can be updated, and this can have cascading effects on other resources, but
these effects are event-driven, not necessary sequential, and are certainly not "workflows" or
atomic transactions that can be rolled back. Changes are expected to be dynamic, asynchronous, and
"eventual". In other words: the total state of the workload is emergent rather than imposed.

This is so different from "legacy" LCM that it's probably best not to use that term in this
scenario. Kubernetes introduces a new, cloud-native orchestration paradigm.

### Why is there a built-in inventory? Shouldn't the inventory be managed externally?

Surely, for production systems a robust inventory is necessary. Turandot can work with various
inventory backends, as well as any container image repository adhering to the
[OCI](https://www.opencontainers.org/) or Docker standards, e.g.
[Quay](https://github.com/quay/quay) and [Harbor](https://goharbor.io/). Indeed, the internal
repository is a simple Docker repository. Note that Turandot can store and retrieves CSAR files from
such repositories even though they are not container images.

The built-in inventory does not have to be used in production, but it can be useful as a local cache
if the repositories are slow to access or if access is unreliable, e.g. on cloud edge datacenters.

### Why use TOSCA and CSARs instead of packaged Helm charts?

Turandot comes with a Helm profile that allows you to package one or more Helm charts inside the
CSAR or install them from an external chart repository. See the [example](examples/helm/). This
feature allows you to combine the advantages of TOSCA and Turandot with existing Helm packaging
efforts.

However, it is worth considering converting your Helm charts into pure TOSCA CSARs.  

A Helm chart is essentially a collection of text templates for low-level Kubernetes YAML resource
manifests stored in a bespoke [repository format](https://helm.sh/docs/topics/chart_repository/). Up
to Helm version 3, it had an in-cluster controller named Tiller. At version 3 it was removed,
leaving Helm entirely devoted to text templating.

Text templating is a rather miserable mechanism for generating YAML, and it's hard to use it to
model reusable types. By contrast, TOSCA is a strictly-typed object-oriented language that supports
inheritance and topological composition, making it vastly superior for modeling complex cloud
workloads. TOSCA is an industry-supported standard created exactly for this purpose.

### Why is it called "Turandot"?

"Turandot" is [the last opera](https://en.wikipedia.org/wiki/Turandot) by composer
[Giacomo Puccini](https://en.wikipedia.org/wiki/Giacomo_Puccini), likely inspired by Count Carlo
Gozzi's *commedia dell'arte*
[play of the same name](https://en.wikipedia.org/wiki/Turandot_(Gozzi)). Its final aria,
[Nessun Dorma](https://en.wikipedia.org/wiki/Nessun_dorma), is one of the
[most well-known of all arias](https://www.youtube.com/watch?v=cWc7vYjgnTs). Puccini is also famous
for his Tosca opera. See, everything is connected.

Turandot, the name of the protagonist of the opera, comes from Persian *Turandokht*, meaning
"daughter of Turan", *Turan* being an older name for
[much of what we now call Central Asia](https://en.wikipedia.org/wiki/Turan). Turan in turn is named
for its legendary ruler, *Tūr* (meaning "brave"), a prince of the ancient
[Shahnameh epic](https://en.wikipedia.org/wiki/Shahnameh).

There is some disagreement over whether the final "t" should be pronounced or not, as it likely
wasn't pronounced by Puccini himself. All you should know is that if you pronounce it incorrectly
this software will not work well for you.
