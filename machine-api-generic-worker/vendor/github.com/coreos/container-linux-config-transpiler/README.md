# Container Linux Config Transpiler

The Config Transpiler ("ct" for short) is the utility responsible for transforming a human-friendly Container Linux Config into a JSON file. This resulting file can be provided to a Container Linux machine when it first boots to provision the machine.

## Documentation

If you're looking to begin writing configs for your Container Linux machines, check out the [getting started][get-started] documentation.

The [configuration][config] documentation is a comprehensive resource specifying what options can be in a Container Linux Config.

For a more in-depth view of ct and why it exists, take a look at the [Overview][overview] document.

Please use the [bug tracker][issues] to report bugs.

[ignition]: https://github.com/coreos/ignition
[issues]: https://issues.coreos.com
[overview]: doc/overview.md
[get-started]: doc/getting-started.md
[config]: doc/configuration.md

## Examples

There are plenty of small, self-contained examples [in the documentation][examples].

[examples]: doc/examples.md

## Installation

### Prebuilt binaries

The easiest way to get started using ct is to download one of the binaries from the [releases page on GitHub][releases].

[releases]: https://github.com/coreos/container-linux-config-transpiler/releases

### Building from source

To build from source you'll need to have the go compiler installed on your system.

```shell
git clone --branch v0.3.1 https://github.com/coreos/container-linux-config-transpiler
cd container-linux-config-transpiler
make build
```

The `ct` binary will be placed in `./bin/`.

Note: Review releases for new branch versions.

## Related projects

- [https://github.com/coreos/ignition](https://github.com/coreos/ignition)
- [https://github.com/coreos/coreos-metadata/](https://github.com/coreos/coreos-metadata/)
- [https://github.com/coreos/matchbox](https://github.com/coreos/matchbox)
