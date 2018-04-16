# Examples

Here you can find a bunch of simple examples for using ct, with some explanations about what they do. The examples here are in no way comprehensive, for a full list of all the options present in ct check out the [configuration specification][spec].

## Users and groups

```yaml container-linux-config
passwd:
  users:
    - name: core
      password_hash: "$6$43y3tkl..."
      ssh_authorized_keys:
        - key1
```

This example modifies the existing `core` user, giving it a known password hash (this will enable login via password), and setting its ssh key.

```yaml container-linux-config
passwd:
  users:
    - name: user1
      password_hash: "$6$43y3tkl..."
      ssh_authorized_keys:
        - key1
        - key2
    - name: user2
      ssh_authorized_keys:
        - key3
```

This example will create two users, `user1` and `user2`. The first user has a password set and two ssh public keys authorized to log in as the user. The second user doesn't have a password set (so log in via password will be disabled), but have one ssh key.

```yaml container-linux-config
passwd:
  users:
    - name: user1
      password_hash: "$6$43y3tkl..."
      ssh_authorized_keys:
        - key1
      home_dir: /home/user1
      no_create_home: true
      groups:
        - wheel
        - plugdev
      shell: /bin/zsh
```

This example creates one user, `user1`, with the password hash `$6$43y3tkl...`, and sets up one ssh public key for the user. The user is also given the home directory `/home/user1`, but it's not created, the user is added to the `wheel` and `plugdev` groups, and the user's shell is set to `/bin/zsh`.

## Storage and files

### Files

```yaml container-linux-config
storage:
  files:
    - path: /opt/file1
      filesystem: root
      contents:
        inline: Hello, world!
      mode: 0644
      user:
        id: 500
      group:
        id: 501
```

This example creates a file at `/opt/file` with the contents `Hello, world!`, permissions 0644 (so readable and writable by the owner, and only readable by everyone else), and the file is owned by user uid 500 and gid 501.

```yaml container-linux-config
storage:
  files:
    - path: /opt/file2
      filesystem: root
      contents:
        remote:
          url: http://example.com/file2
          compression: gzip
          verification:
            hash:
              function: sha512
              sum: 4ee6a9d20cc0e6c7ee187daffa6822bdef7f4cebe109eff44b235f97e45dc3d7a5bb932efc841192e46618f48a6f4f5bc0d15fd74b1038abf46bf4b4fd409f2e
      mode: 0644
```

This example fetches a gzip-compressed file from `http://example.com/file2`, makes sure that it matches the provided sha512 hash, and writes it to `/opt/file2`.

### Filesystems

```yaml container-linux-config
storage:
  filesystems:
    - name: filesystem1
      mount:
        device: /dev/disk/by-partlabel/ROOT
        format: btrfs
        wipe_filesystem: true
        label: ROOT
```

This example formats the root filesystem to be `btrfs`, and names it `filesystem1` (primarily for use in the `files` section).

## systemd units

```yaml container-linux-config
systemd:
  units:
    - name: etcd-member.service
      dropins:
        - name: conf1.conf
          contents: |
            [Service]
            Environment="ETCD_NAME=infra0"
```

This example adds a drop-in for the `etcd-member` unit, setting the name for etcd to `infra0` with an environment variable. More information on systemd dropins can be found in [the docs][dropins].

```yaml container-linux-config
systemd:
  units:
    - name: hello.service
      enabled: true
      contents: |
        [Unit]
        Description=A hello world unit!
        Type=oneshot

        [Service]
        ExecStart=/usr/bin/echo "Hello, World!"

        [Install]
        WantedBy=multi-user.target
```

This example creates a new systemd unit called hello.service, enables it so it will run on boot, and defines the contents to simply echo `"Hello, World!"`.

## networkd units

```yaml container-linux-config
networkd:
  units:
    - name: static.network
      contents: |
        [Match]
        Name=enp2s0

        [Network]
        Address=192.168.0.15/24
        Gateway=192.168.0.1
```

This example creates a networkd unit to set the IP address on the `enp2s0` interface to the static address `192.168.0.15/24`, and sets an appropriate gateway. More information on networkd units in CoreOS can be found in [the docs][networkd].

## etcd

```yaml container-linux-config:norender
etcd:
  version:                     "3.0.15"
  name:                        "{HOSTNAME}"
  advertise_client_urls:       "http://{PRIVATE_IPV4}:2379"
  initial_advertise_peer_urls: "http://{PRIVATE_IPV4}:2380"
  listen_client_urls:          "http://0.0.0.0:2379"
  listen_peer_urls:            "http://{PRIVATE_IPV4}:2380"
  initial_cluster:             "{HOSTNAME}=http://{PRIVATE_IPV4}:2380"
```

This example will create a dropin for the `etcd-member` systemd unit, configuring it to use the specified version and adding all the specified options. This will also enable the `etcd-member` unit.

This is referencing dynamic data that isn't known until an instance is booted. For more information on how this works, please take a look at the [referencing dynamic data][dynamic-data] document.

## Updates and Locksmithd

```yaml container-linux-config
update:
  group:  "beta"
locksmith:
  reboot_strategy: "etcd-lock"
  window_start:    "Sun 1:00"
  window_length:   "2h"
```

This example configures the Container Linux instance to be a member of the beta group, configures locksmithd to acquire a lock in etcd before rebooting for an update, and only allows reboots during a 2 hour window starting at 1 AM on Sundays.

[spec]: configuration.md
[dropins]: https://coreos.com/os/docs/latest/using-systemd-drop-in-units.html
[networkd]: https://coreos.com/os/docs/latest/network-config-with-networkd.html
[dynamic-data]: dynamic-data.md
