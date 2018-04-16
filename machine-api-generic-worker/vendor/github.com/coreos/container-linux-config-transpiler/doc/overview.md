# Config transpiler overview

The Config Transpiler, ct, is the utility responsible for transforming a user-provided Container Linux Configuration into an [Ignition][ignition] configuration. The resulting Ignition config can then be provided to a Container Linux machine when it first boots in order to provision it.

The Container Linux Config is intended to be human-friendly, and is thus in YAML. The syntax is rather forgiving, and things like references and multi-line strings are supported.

The resulting Ignition config is very much not intended to be human-friendly. It is an artifact produced by ct that users should simply pass along to their machines. JSON was chosen over a binary format to make the process more transparent and to allow power users to inspect/modify what ct produces, but it would have worked fine if the result from ct had not been human readable at all.

[ignition]: https://github.com/coreos/ignition

## Why a two-step process?

There are a couple factors motivating the decision to not incorporate support for Container Linux Configs directly into the boot process of Container Linux (as in, the ability to provide a Container Linux Config directly to a booting machine, instead of an Ignition config).

- By making users run their configs through ct before they attempt to boot a machine, issues with their configs can be caught before any machine attempts to boot. This will save users time, as they can much more quickly find problems with their configs. Were users to provide Container Linux Configs directly to machines at first boot, they would need to find a way to extract the Ignition logs from a machine that may have failed to boot, which can be a slow and tedious process.
- YAML parsing is a complex process that in the past has been rather error-prone. By only doing JSON parsing in the boot path, we can guarantee that the utilities necessary for a machine to boot are simpler and more reliable. We want to allow users to use YAML however, as it's much more human-friendly than JSON, hence the decision to have a tool separate from the boot path to "transpile" YAML configurations to machine-appropriate JSON ones.

## Tell me more about Ignition

[Ignition][ignition] is the utility inside of a Container Linux image that is responsible for setting up a machine. It takes in a configuration, written in JSON, that instructs it to do things like add users, format disks, and install systemd units. The artifacts that ct produces are Ignition configs. All of this should be an implementation detail however, users are encouraged to write Container Linux Configs for ct, and to simply pass along the produced JSON file to their machines.

## How similar are Container Linux Configs and Ignition configs?

Some features in Container Linux Configs and Ignition configs are identical.  Both support listing users for creation, systemd unit dropins for installation, and files for writing.

All of the differences stem from the fact that Ignition configs are distribution agnostic. An Ignition config can't just tell Ignition to enable etcd, because Ignition doesn't know what etcd is. The config must tell Ignition what systemd unit to enable, and provide a systemd dropin to configure etcd.

ct on the other hand _does_ understand the specifics of Container Linux. A user can merely specify an etcd version and some etcd options, and ct knows that there's a unit called `etcd-member` already on the system it can enable. It knows what options are supported by etcd, so it can sanity check them for the user. It can then generate an appropriate systemd dropin with the user's options, and provide Ignition the level of verbosity it needs, that would be tedious for a human to create.
