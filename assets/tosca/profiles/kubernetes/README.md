Kubernetes Profile
==================

Node Types
----------

Kubernetes Metadata
-------------------

Capability Type Metadata
------------------------

`turandot.apiVersion`: Kubernetes resource API version. This metadata value triggers a Kubernetes
resource generation.

`turandot.kind`: Kubernetes resource kind. If not specified will default to the capability type name.

`turandot.move[.#]`: Move a value or branch in the resource. The format is "<from path>-><to path>".
Adding "." to the key is optional and used for ensuring the order of move operations, as the keys
are sorted alphabetically before processing.

`turandot.copy[.#]`: Copy a value or branch in the resource. The format is "<from path>-><to path>".
Adding "." to the key is optional and used for ensuring the order of move operations, as the keys
are sorted alphabetically before processing. Note that moves are processed before copies.

Data Type Metadata
------------------

Note that TOSCA scalar-unit types will always be converted to numbers (floats or integers as
appropriate).

Property Metadata
-----------------

`puccini.information:turandot.ignore`: When "true" (a string) will *not* export the property to the
Kubernetes resource. 

Attribute Metadata
------------------

`puccini.information:turandot.mapping`: The value is the path within the Kubernetes resource from which to
map the attribute value. Note that both "status." and "spec." can be used.
