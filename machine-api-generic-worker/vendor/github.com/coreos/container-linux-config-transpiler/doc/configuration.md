# Configuration Specification #

A Container Linux Configuration, to be processed by ct, is a YAML document conforming to the following specification:

_Note: all fields are optional unless otherwise marked_

* **ignition** (object): metadata about the configuration itself.
  * **config** (objects): options related to the configuration.
    * **append** (list of objects): a list of the configs to be appended to the current config.
      * **source** (string, required): the URL of the config. Supported schemes are http and https. Note: When using http, it is advisable to use the verification option to ensure the contents haven't been modified.
      * **verification** (object): options related to the verification of the config.
        * **hash** (object): the hash of the config
          * **function** (string): the function used to hash the config. Supported functions are sha512.
          * **sum** (string): the resulting sum of the hash applied to the contents.
    * **replace** (object): the config that will replace the current.
      * **source** (string, required): the URL of the config. Supported schemes are http and https. Note: When using http, it is advisable to use the verification option to ensure the contents haven't been modified.
      * **verification** (object): options related to the verification of the config.
        * **hash** (object): the hash of the config
          * **function** (string): the function used to hash the config. Supported functions are sha512.
          * **sum** (string): the resulting sum of the hash applied to the contents.
  * **timeouts** (object): options relating to http timeouts when fetching files over http or https.
    * **httpResponseHeaders** (integer) the time to wait (in seconds) for the server's response headers (but not the body) after making a request. 0 indicates no timeout. Default is 10 seconds.
    * **httpTotal** (integer) the time limit (in seconds) for the operation (connection, request, and response), including retries. 0 indicates no timeout. Default is 0.
* **storage** (object): describes the desired state of the system's storage devices.
  * **disks** (list of objects): the list of disks to be configured and their options.
    * **device** (string, required): the absolute path to the device. Devices are typically referenced by the `/dev/disk/by-*` symlinks.
    * **wipe_table** (boolean): whether or not the partition tables shall be wiped. When true, the partition tables are erased before any further manipulation. Otherwise, the existing entries are left intact.
    * **partitions** (list of objects): the list of partitions and their configuration for this particular disk.
      * **label** (string): the PARTLABEL for the partition.
      * **number** (integer): the partition number, which dictates it's position in the partition table (one-indexed). If zero, use the next available partition slot.
      * **size** (string): the size of the partition with a unit (KiB, MiB, GiB). If zero, the partition will fill the remainder of the disk.
      * **start** (string): the start of the partition with a unit (KiB, MiB, GiB). If zero, the partition will be positioned at the earliest available part of the disk.
      * **type_guid** (string): the GPT [partition type GUID][part-types]. If omitted, the default will be 0FC63DAF-8483-4772-8E79-3D69D8477DE4 (Linux filesystem data). The keywords `linux_filesystem_data`, `raid_partition`, `swap_partition`, and `raid_containing_root` can also be used.
      * **guid** (string): the GPT unique partition GUID.
  * **raid** (list of objects): the list of RAID arrays to be configured.
    * **name** (string, required): the name to use for the resulting md device.
    * **level** (string, required): the redundancy level of the array (e.g. linear, raid1, raid5, etc.).
    * **devices** (list of strings, required): the list of devices (referenced by their absolute path) in the array.
    * **spares** (integer): the number of spares (if applicable) in the array.
  * **filesystems** (list of objects): the list of filesystems to be configured and/or used in the "files" section. Either "mount" or "path" needs to be specified.
    * **name** (string): the identifier for the filesystem, internal to Ignition. This is only required if the filesystem needs to be referenced in the "files" section.
    * **mount** (object): contains the set of mount and formatting options for the filesystem. A non-null entry indicates that the filesystem should be mounted before it is used by Ignition.
      * **device** (string, required): the absolute path to the device. Devices are typically referenced by the `/dev/disk/by-*` symlinks.
      * **format** (string, required): the filesystem format (ext4, btrfs, or xfs).
      * **wipeFilesystem** (boolean): whether or not to wipe the device before filesystem creation, see [Ignition's documentation on filesystems][ignition-fs-reuse] for more information.
      * **label** (string): the label of the filesystem.
      * **uuid** (string): the uuid of the filesystem.
      * **options** (list of strings): any additional options to be passed to the format-specific mkfs utility.
      * **create** (object, DEPRECATED): contains the set of options to be used when creating the filesystem. A non-null entry indicates that the filesystem shall be created.
        * **force** (boolean, DEPRECATED): whether or not the create operation shall overwrite an existing filesystem.
        * **options** (list of strings, DEPRECATED): any additional options to be passed to the format-specific mkfs utility.
    * **path** (string): the mount-point of the filesystem. A non-null entry indicates that the filesystem has already been mounted by the system at the specified path. This is really only useful for "/sysroot".
  * **files** (list of objects): the list of files, rooted in this particular filesystem, to be written.
    * **filesystem** (string, required): the internal identifier of the filesystem. This matches the last filesystem with the given identifier.
    * **path** (string, required): the absolute path to the file.
    * **contents** (object): options related to the contents of the file.
      * **inline** (string): the contents of the file.
      * **remote** (object): options related to the fetching of remote file contents.
        * **compression** (string): the type of compression used on the contents (null or gzip)
        * **url** (string): the URL of the file contents. Supported schemes are http and [data][rfc2397]. Note: When using http, it is advisable to use the verification option to ensure the contents haven't been modified.
        * **verification** (object): options related to the verification of the file contents.
          * **hash** (object): the hash of the config
            * **function** (string): the function used to hash the config. Supported functions are sha512.
            * **sum** (string): the resulting sum of the hash applied to the contents.
    * **mode** (integer): the file's permission mode.
    * **user** (object): specifies the file's owner.
      * **id** (integer): the user ID of the owner.
      * **name** (string): the user name of the owner.
    * **group** (object): specifies the group of the owner.
      * **id** (integer): the group ID of the owner.
      * **name** (string): the group name of the owner.
  * **directories** (list of objects): the list of directories to be created.
    * **filesystem** (string, required): the internal identifier of the filesystem in which to create the directory. This matches the last filesystem with the given identifier.
    * **path** (string, required): the absolute path to the directory.
    * **mode** (integer): the directory's permission mode.
    * **user** (object): specifies the directory's owner.
      * **id** (integer): the user ID of the owner.
      * **name** (string): the user name of the owner.
    * **group** (object): specifies the group of the owner.
      * **id** (integer): the group ID of the owner.
      * **name** (string): the group name of the owner.
  * **links** (list of objects): the list of links to be created
    * **filesystem** (string, required): the internal identifier of the filesystem in which to write the link. This matches the last filesystem with the given identifier.
    * **path** (string, required): the absolute path to the link
    * **user** (object): specifies the symbolic link's owner.
      * **id** (integer): the user ID of the owner.
      * **name** (string): the user name of the owner.
    * **group** (object): specifies the group of the owner.
      * **id** (integer): the group ID of the owner.
      * **name** (string): the group name of the owner.
    * **target** (string, required): the target path of the link
    * **hard** (boolean): a symbolic link is created if this is false, a hard one if this is true.
* **systemd** (object): describes the desired state of the systemd units.
  * **units** (list of objects): the list of systemd units.
    * **name** (string, required): the name of the unit. This must be suffixed with a valid unit type (e.g. "thing.service").
    * **enable** (boolean, DEPRECATED): whether or not the service shall be enabled. When true, the service is enabled. In order for this to have any effect, the unit must have an install section.
    * **enabled** (boolean): whether or not the service shall be enabled. When true, the service is enabled. When false, the service is disabled. When omitted, the service is unmodified. In order for this to have any effect, the unit must have an install section.
    * **mask** (boolean): whether or not the service shall be masked. When true, the service is masked by symlinking it to `/dev/null`.
    * **contents** (string): the contents of the unit.
    * **dropins** (list of objects): the list of drop-ins for the unit.
      * **name** (string, required): the name of the drop-in. This must be suffixed with ".conf".
      * **contents** (string): the contents of the drop-in.
* **networkd** (object): describes the desired state of the networkd files.
  * **units** (list of objects): the list of networkd files.
    * **name** (string, required): the name of the file. This must be suffixed with a valid unit type (e.g. "00-eth0.network").
    * **contents** (string): the contents of the networkd file.
* **passwd** (object): describes the desired additions to the passwd database.
  * **users** (list of objects): the list of accounts to be added.
    * **name** (string, required): the username for the account.
    * **password_hash** (string): the encrypted password for the account.
    * **ssh_authorized_keys** (list of strings): a list of SSH keys to be added to the user's authorized_keys.
    * **uid** (integer): the user ID of the account.
    * **gecos** (string): the GECOS field of the account.
    * **home_dir** (string): the home directory of the account.
    * **no_create_home** (boolean): whether or not to create the user's home directory. This only has an effect if the account doesn't exist yet.
    * **primary_group** (string): the name of the primary group of the account.
    * **groups** (list of strings): the list of supplementary groups of the account.
    * **no_user_group** (boolean): whether or not to create a group with the same name as the user. This only has an effect if the account doesn't exist yet.
    * **no_log_init** (boolean): whether or not to add the user to the lastlog and faillog databases. This only has an effect if the account doesn't exist yet.
    * **shell** (string): the login shell of the new account.
    * **system** (bool): whether or not to make the account a system account. This only has an effect if the account doesn't exist yet.
    * **create** (object, DEPRECATED): contains the set of options to be used when creating the user. A non-null entry indicates that the user account shall be created.
      * **uid** (integer, DEPRECATED): the user ID of the new account.
      * **gecos** (string, DEPRECATED): the GECOS field of the new account.
      * **home_dir** (string, DEPRECATED): the home directory of the new account.
      * **no_create_home** (boolean, DEPRECATED): whether or not to create the user's home directory.
      * **primary_group** (string, DEPRECATED): the name or ID of the primary group of the new account.
      * **groups** (list of strings, DEPRECATED): the list of supplementary groups of the new account.
      * **no_user_group** (boolean, DEPRECATED): whether or not to create a group with the same name as the user.
      * **no_log_init** (boolean, DEPRECATED): whether or not to add the user to the lastlog and faillog databases.
      * **shell** (string, DEPRECATED): the login shell of the new account.
  * **groups** (list of objects): the list of groups to be added.
    * **name** (string, required): the name of the group.
    * **gid** (integer): the group ID of the new group.
    * **password_hash** (string): the encrypted password of the new group.
* **etcd**
  * **version** (string): the version of etcd to be run
  * **_other options_** (string): this section accepts any valid etcd options for the version of etcd specified. For a comprehensive list, please consult etcd's documentation. Note all options here should be in snake_case, not spine-case.
* **flannel**
  * **version** (string): the version of flannel to be run
  * **network_config** (string): the flannel configuration to be written into etcd before flannel starts.
  * **_other options_** (string): this section accepts any valid flannel options for the version of flannel specified. For a comprehensive list, please consult flannel's documentation. Note all options here should be in snake_case, not spine-case.
* **docker**
  * **flags** (list of strings): additional flags to pass to the docker daemon when it is started
* **update**
  * **group** (string): the update group to follow. Most users will want one of: stable, beta, alpha.
  * **server** (string): the server to fetch updates from.
* **locksmith**
  * **reboot_strategy** (string): the reboot strategy for locksmithd to follow. Must be one of: reboot, etcd-lock, off.
  * **window_start** (string, required if window-length isn't empty): the start of the window that locksmithd can reboot the machine during
  * **window_length** (string, required if window-start isn't empty): the duration of the window that locksmithd can reboot the machine during
  * **group** (string): the locksmith etcd group to be part of for reboot control
  * **etcd_endpoints** (string): the endpoints of etcd locksmith should use
  * **etcd_cafile** (string): the tls CA file to use when communicating with etcd
  * **etcd_certfile** (string): the tls cert file to use when communicating with etcd
  * **etcd_keyfile** (string): the tls key file to use when communicating with etcd

[part-types]: http://en.wikipedia.org/wiki/GUID_Partition_Table#Partition_type_GUIDs
[rfc2397]: https://tools.ietf.org/html/rfc2397
[ignition-fs-reuse]: https://github.com/coreos/ignition/blob/master/doc/operator-notes.md#filesystem-reuse-semantics
