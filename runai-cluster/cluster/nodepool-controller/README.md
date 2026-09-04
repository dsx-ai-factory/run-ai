# nodepool-controller

The nodepool-controller is a controller responsible for the crd "nodepool".
A node pool is a group of nodes. The controller is reponsible to tag each node with the nodepool name it belongs to.
Also, it's responsible for starting a kai-scheduler for each nodepool.
And, of course, for updating the correct status for each nodepool.
