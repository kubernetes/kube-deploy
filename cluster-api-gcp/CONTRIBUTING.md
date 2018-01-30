# Contributing Guidelines

## Developing

When making changes to the machine controller, it's generally a good idea to delete any existing cluster created with an older version of the cluster-api.

```
$ ./cluster-api-gcp delete
```

After making changes to the machine controller or the actuator, you need to follow these two steps:

1. Rebuild the machine-controller image.

	```
	$ cd machine-controller
	$ make push fix-image-permissions
	```

2. Rebuild cluster-api-gcp

	```
	$ go build
	```

The new `cluster-api-gcp` will have your changes.
