*This is an early release. Some features are not yet fully implemented.*

Turandot
========

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Latest Release](https://img.shields.io/github/release/tliron/turandot.svg)](https://github.com/tliron/turandot/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/tliron/turandot)](https://goreportcard.com/report/github.com/tliron/turandot)

Compose and orchestrate [Kubernetes](https://kubernetes.io/) workloads using
[TOSCA](https://www.oasis-open.org/committees/tosca/).

Want to dive in?

Check out the included [examples](examples/) to understand what you can do with Turandot, and then
head to the [tutorial](TUTORIAL.md) to get up and running.


Get It
------

[![Download](assets/media/download.png "Download")](https://github.com/tliron/turandot/releases)


Features
--------

**Complex workloads**: Turandot targets complex, large-scale, and multi-cluster workloads.
Many examples are found in the field of Network Function Virtualization (NFV) (e.g.
[MANO](https://en.wikipedia.org/wiki/Network_function_virtualization#Management_and_orchestration_%28MANO%29)).
Included with Turandot is a comprehensive example of a multi-cluster
[telephony network service](https://github.com/tliron/turandot/tree/main/examples/telephony-network-service/)
modeled entirely in TOSCA.

**Diverse workloads**: Workloads can comprise both standard and
[custom](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
Kubernetes resources, as well as their
[operators](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/). They can be deployed
on a single cluster or on multi-cluster clouds. Virtual machines are supported via
[KubeVirt](https://kubevirt.io/).

**Service composition**: Turandot implements TOSCA substitution mappings via policy-based service
composition based on service templates selected from a repository. 

**Plugins**: [Helm charts](https://helm.sh/) and external orchestrators, such as
[Ansible](https://www.ansible.com/), are supported via custom artifacts encapsulated as TOSCA
types.


Rationale
---------

**Design-time**: TOSCA's extensibility via an object-oriented grammar is  to Kubernetes's
extensibility via custom resource definitions and operators. TOSCA's added value for Kubernetes is
in providing a composable and validated graph of resource interrelations, effectively imbuing
Kubernetes resources with architectural intent.

**Run-time**: Turandot manages resources *together* as single, coherent workloads, even across
cluster boundaries, ensuring consistency and integration as well as allowing for cascading policies
for allocation, composition, networking, security, etc.


How It Works
------------

Turandot is an in-cluster Kubernetes operator that:

1. Handles custom resources called "services".
   ([here is the CRD](assets/kubernetes/custom-resource-definition.yaml))
2. Can work with an internal (built-in) or external repositories to retrieve CSAR-packaged service
   templates. A CSAR
   ([Cloud Service Archive](https://docs.oasis-open.org/tosca/TOSCA-Simple-Profile-YAML/v1.3/os/TOSCA-Simple-Profile-YAML-v1.3-os.html#_Toc26969474))
   is a zip file containing a TOSCA service template, TOSCA profiles, and other files ("artifacts")
   required for orchestration (see #5, below).
3. Uses [Puccini](https://puccini.cloud/) to compile the CSAR-packaged service templates into the
   [Clout](https://puccini.cloud/clout/) intermediary format.
4. Renders the Clout to Kubernetes resources and schedules them as integrated workloads.
5. Deploys and activates artifacts packaged in the CSAR file. This includes container images (as
   well as KubeVirt virtual machine images) and cloud-native configuration tools, such as scripts,
   playbooks, recipes, etc., as well as Kubernetes operators. These configuration tools have access
   to the entire workload topology, allowing them to essentially configure themselves.
6. Are some of the resources remote? Turandot will delegate orchestration to Turandot operators in
   remote clusters (see multi-cluster workloads, below).

The Turandot operator can be controlled using the stateless `turandot` utility, e.g.:

    turandot service deploy my-service --file=my-service-template.csar

⮕ [Documentation](turandot/)

Note that this utility is merely a convenience, not a requirement. You can use your existing
Kubernetes tools to interact with the "service" custom resources.


Cloud-Native Self-Orchestration
-------------------------------

Self-orchestration is coordinated by setting a "mode" for the entire workflow, following a pattern
we call the Town-Crier Model. This mode is a *proclamation*: a modal, asynchronous, system-wide
event, which answers the question "What should we be doing now?" Service template designers can
attach actions to certain modes, modeled as TOSCA interfaces that use Kubernetes command streaming
(or SSH for KubeVirt virtual machines). The guiding assumption is that components know their own
status, needs, and obstacles better than a centralized orchestrator ever could, and so the best
approach is to optimize for coordination rather than dictation. The end result is that the total
state of the system is *emergent* rather than imposed.

Turandot acts as the proclamation controller, or "town crier", ensuring that the interfaces for the
current mode are continuously polled for all running components and collating their success/failure
statuses, even across multi-cluster boundaries. The Clout functions as the "town hall": it is where
components can continuously store and share configuration data with the entire topology.


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


Putting It All Together: The Cycle of Life
------------------------------------------

**Day -1: Modeling.** TOSCA is used to create "profiles" of reusable, composable types, which
together provide a validated and validating model for the target domain. TOSCA profiles vastly
simplify the work of the service template designer. For example, our telephony network service
example uses profiles for Kubernetes, KubeVirt, network services (including data planes), and
telephony.

**Day 0: Design.** Solution architects compose service templates from the models provided by the
TOSCA profiles, either by writing the TOSCA manually, or by using a wonderful graphical TOSCA IDE
(that is yet to be created). The templates are tested in lab and staging environments using
CI/CD-style automation.

**Day 1: Operations handoff.** The service templates are ready to be instantiated in production.
A ticket from an operations support system (OSS) initiates the transfer to a managed multi-cluster
cloud. Turandot is deployed to the target clusters (or automatically delegated from central
clusters) and takes it from there.

**Day 2+: Cloud-native operations.** Once they are up and running the services should orchestrate
themselves by adapting to changing internal and external conditions, as well as triggered and manual
actions from operations. Changes include scaling, healing, migration, as well as more elaborate
transformations. The Turandot operator will continue to monitor these changes and update the Clout.
Components can refer to the Clout as "single source of truth" to see the complete topology in order
to make self-orchestration decisions, as well as checking against policies to which they must or can
adhere. Machine learning and AI can be applied to the Clout in order to make the best possible
runtime orchestration decisions.


FAQ
---

### Is Turandot a lifecycle manager (LCM) for Kubernetes workloads?

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

### Is Turandot a replacement for full-blown NFV orchestrators like [ONAP](https://www.onap.org/)?

Absolutely not. Turandot's scope is purposely limited and focused only on managing Kubernetes
workloads. The point is not to replace full-blown orchestrators but rather to make their job much
easier by allowing them delegate the actual work of orchestrating Kubernetes workloads to Kubernetes
itself, thus completing the cloud-native paradigm. All the orchestrator would need to do is tell
Turandot to deploy a workload packaged as a CSAR file, and to provide it with inputs and to process
its outputs. The orchestrator would not have to concern itself with the complex internal composition
of these workloads.

### Why doesn't Turandot include a workflow engine?

Workflow engines are unreliable in any cloud environment but are an especially bad fit for
Kubernetes. In Kubernetes a single container is, by design, controlled by multiple levels of
operators, e.g. changes to the Pod, ReplicaSet, and Deployment resources can cause a container to
restart and lose state at any moment. An event or message may very well be invalid as soon as it is
triggered or sent.

And so Turandot introduces and embraces the Town-Crier Model. The "town" of components will be
*continuously* attempting to achieve the current mode (a modal event). Turandot is merely a
coordinator, not an orchestrator per se. In other words, we encourage *self-orchestration*.

That said, Turandot does not stop you from using a workflow engine if you reall need or want one.
You can delegate to it via the mode interfaces or have it running as an entirely separate system.

For a workflow solution that is well integrated with Kubernetes consider
[Argo Workflows](https://argoproj.github.io/projects/argo), which extends the scheduling
functionality of [Kubernetes jobs](https://kubernetes.io/docs/concepts/workloads/controllers/job/)
to allow for declarative dependency graphs.

(Note: We are working on an TOSCA profile for Argo, which will include a workflow example.)

### Why does Turandot include a built-in repository? Shouldn't the repository be managed externally?

Surely, for production systems a robust repository is necessary. Turandot can work with various
repository backends, as well as any container image repository adhering to the
[OCI](https://www.opencontainers.org/) or Docker standards, e.g.
[Quay](https://www.projectquay.io/) and [Harbor](https://goharbor.io/). Indeed, the internal
repository is implemented via the reference Docker repository. (Note that Turandot can store and
retrieve CSAR files from such repositories even though they are not container images.)

The built-in repository does not have to be used in production, but it can be useful as a local
cache in cases in which the main repositories are slow to access or if access is unreliable, e.g. on
cloud edge datacenters.

### Why use TOSCA and CSARs instead of packaged Helm charts?

Turandot comes with a Helm profile that allows you to package one or more Helm charts inside the
CSAR or install them from an external chart repository. See the
[example](https://github.com/tliron/turandot/tree/main/examples/helm/). This feature allows you to
combine the advantages of TOSCA and Turandot with existing Helm packaging efforts. Up to Helm version
3, Helm had an in-cluster controller named Tiller. At version 3 it was removed, leaving Helm entirely
devoted to text templating. Turandot can be understood in this context as a super-charged replacement
for Tiller. 

All that said, it is worth considering abandoning Helm entirely and converting your charts into pure
TOSCA CSARs. A Helm chart is essentially a collection of text templates for low-level Kubernetes
YAML resource manifests stored in a bespoke
[repository format](https://helm.sh/docs/topics/chart_repository/). Text templating is a rather
miserable mechanism for generating YAML, and it's hard to use it to model reusable types. By
contrast, TOSCA is a strictly-typed object-oriented language that supports inheritance and
topological composition, making it vastly superior for modeling complex cloud workloads. TOSCA and
CSAR are industry-supported standards.

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
