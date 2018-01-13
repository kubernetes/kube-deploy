# Referencing dynamic data

## Overview

Sometimes it can be useful to refer to data in a Container Linux Config that isn't known until a machine boots, like its network address. This can be accomplished with [coreos-metadata][coreos-metadata]. coreos-metadata is a very basic utility that fetches information about the current machine and makes it available for consumption. By making it a dependency of services which requires this information, systemd will ensure that coreos-metadata has successfully completed before starting these services. These services can then simply source the fetched information and let systemd perform the environment variable expansions.

As of version 0.2.0, ct has support for making this easy for users. In specific sections of a config, users can enter in dynamic data between `{}`, and ct will handle enabling the coreos-metadata service and using the information it provides.

The available information varies by provider, and is expressed in different variables by coreos-metadata. If this feature is used a `--provider` flag must be passed to ct. Currently, the `etcd` and `flannel` sections are the only ones which support this feature.

[coreos-metadata]: https://github.com/coreos/coreos-metadata/

## Supported data by provider

This is the information available in each provider.

|                    | `HOSTNAME` | `PRIVATE_IPV4` | `PUBLIC_IPV4` | `PRIVATE_IPV6` | `PUBLIC_IPV6` |
|--------------------|------------|----------------|---------------|----------------|---------------|
| Azure              |            | ✓              | ✓             |                |               |
| Digital Ocean      | ✓          | ✓              | ✓             | ✓              | ✓             |
| EC2                | ✓          | ✓              | ✓             |                |               |
| GCE                | ✓          | ✓              | ✓             |                |               |
| Packet             | ✓          | ✓              | ✓             |                | ✓             |
| OpenStack-Metadata | ✓          | ✓              | ✓             |                |               |
| Vagrant-Virtualbox | ✓          | ✓              |               |                |               |

## Behind the scenes

For a more in-depth walk through of how this feature works, let's look at the etcd example from the [examples document][examples].

```yaml
etcd:
  version:                     "3.0.15"
  name:                        "{HOSTNAME}"
  advertise_client_urls:       "http://{PRIVATE_IPV4}:2379"
  initial_advertise_peer_urls: "http://{PRIVATE_IPV4}:2380"
  listen_client_urls:          "http://0.0.0.0:2379"
  listen_peer_urls:            "http://{PRIVATE_IPV4}:2380"
  initial_cluster:             "{HOSTNAME}=http://{PRIVATE_IPV4}:2380"
```

If we give this example to ct with the `--platform=ec2` tag, it produces the following drop-in:

```
[Unit]
Requires=coreos-metadata.service
After=coreos-metadata.service

[Service]
EnvironmentFile=/run/metadata/coreos
Environment="ETCD_IMAGE_TAG=v3.0.15"
ExecStart=
ExecStart=/usr/lib/coreos/etcd-wrapper $ETCD_OPTS \
  --name="${COREOS_EC2_HOSTNAME}" \
  --listen-peer-urls="http://${COREOS_EC2_IPV4_LOCAL}:2380" \
  --listen-client-urls="http://0.0.0.0:2379" \
  --initial-advertise-peer-urls="http://${COREOS_EC2_IPV4_LOCAL}:2380" \
  --initial-cluster="${COREOS_EC2_HOSTNAME}=http://${COREOS_EC2_IPV4_LOCAL}:2380" \
  --advertise-client-urls="http://${COREOS_EC2_IPV4_LOCAL}:2379"
```

This drop-in specifies that etcd should run after the coreos-metadata service, and it uses `/run/metadata/coreos` as an `EnvironmentFile`. This enables the coreos-metadata service, and puts the information it discovers into environment variables. These environment variables are then expanded by systemd when the service starts, inserting the dynamic data into the command-line flags to etcd.

[examples]: examples.md
