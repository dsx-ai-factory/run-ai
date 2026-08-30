# shared-objects-controller

The shared-objects-controller will control the objects the need to be copied to scoped namespaces.

The first usage will be to control the datavolume crds: copy the pointed pvc to the shared scopes (the namespaces of the projects/departments that need to access the data).