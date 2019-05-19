# Perfstatbeat

Perfstatbeat is a beat based on Metricbeat which collects performance metrics through the perfstat API of the AIX operating system and supplies them in a form, that is mostly compatible with the Metricbeat system module.

## Current state

These modules and metrics have already been implemented:

- `system.load`

    Same fields as the upstream module.

- `system.diskio`

    - Disk name, volume group name and unique disk identifier (UDID).
    - IO count, size and time per read and write (i.e. average IO time can be calculated).

This does not seem to be much, but it's both a working metric that uses a global interface of the perfstat API, as well as one that uses a component-specific interface that is an example of implementing others.

## Getting started

### Build requirements

- An AIX operating system instance.
- A recent version of `gcc-go` and its dependencies. I've build the software on these combinations:
    - AIX 7.2 and the Go packages of the [AIX Toolbox for Linux Applications](https://www.ibm.com/developerworks/aix/library/aix-toolbox/)
    - AIX 7.1 and the Go packages kindly provided by [BullFreeware](http://www.bullfreeware.com/search.php?package=gcc-go)
- GNU `make`, e.g. from one of the sources above or packages kindly provided by [Michael Perzl](http://www.perzl.org/aix/).

### Runtime requirements

- `libgo` should come from the same Go packages as the build environment.
- `libgcc`, not necessarily at the same level or from the same source like `libgo`.

### Compiling and running

To compile your beat run `gmake`. Then you can run the following command to see the first output:

```
./perfstatbeat -e -d "*"
```

Note, that I've seen linker problems when the `ar` command of GNU binutils is beeing used insted of the native one. This could happen, if `/opt/freeware/bin` shows up in front of `/usr/bin` in the `PATH` environment.

## Development

Perfstatbeat is a beat based on Metricbeat which was generated with metricbeat/metricset generator.

Have a look at the [upstream documentation](https://www.elastic.co/guide/en/beats/devguide/6.2/metricbeat-developer-guide.html) to start with development.

### Requirements

- Internet access or an outbound proxy server (use `export https_proxy=http://[user:passwd@]proxy.server:port`).
- `python` and `python-pip` also GNU `coreutils` and `findutils`, e.g. from the sources listed above.
- `virtualenv`, which might be installed with `pip install virtualenv`.

Some make steps depend on GNU specific options of the `cp` and `find` commands, i.e. the have to be used instead of the native ones. However, putting `/opt/freeware/bin` to the beginning of the `PATH` environment could lead to the linker error described above. This can be solved, for example, as follows:

```
mkdir -p ~/bin
ln -s /opt/freeware/bin/cp /opt/freeware/bin/find ~/bin/
export PATH=~/bin:$PATH
```

Furthermore these steps will need to load some Python modules from the upstream sources, so set up the `PYTHONPATH` environment like this:
```
export PYTHONPATH=vendor/github.com/elastic/beats/metricbeat/scripts
```

### Extending Perfstatbeat

In case further modules and metricsets should be added, run:

```
gmake create-metricset
```

This will create the necessary boilerplate code. Check the [upstream documentation](https://www.elastic.co/guide/en/beats/devguide/6.2/creating-metricsets.html) to see how you can proceed.

In the existing modules (e.g. [load](https://github.com/WuerthIT/perfstatbeat/blob/master/module/system/load/load.go)) you can see how to access the perfstat API from Go. [cgo](https://golang.org/cmd/cgo/) contains information on how to call C code from Go routines.

You can find the description of the perfstat API in the [IBM Knowledge Center](https://www.ibm.com/support/knowledgecenter/en/ssw_aix_72/com.ibm.aix.prftools/idprftools_perfstat.htm). The AIX operating system contains some code examples written in C inside `/usr/samples/libperfstat`.

After updates to the fields or config files, always run

```
gmake collect
```

This updates all fields and docs with the most recent changes.

## Vendoring

Perfstatbeat currently includes version 6.2.4 of beats in the `vendor` subfolder with [some minor modifications](https://github.com/WuerthIT/beats/releases/tag/v6.2.4-support_aix) on libraries inside their `vendor` directory. Later versions make use of Go modules that are not available on the AIX operation system currently.

## Packaging

The original packaging process makes use of containers and will obviously not work here. So currently the binary file has to be distributed manually. Maybe an RPM spec file will be provided later.

## Disclaimer

AIX is a registered trademark of the International Business Machines Corporation.

Metricbeat is a trademark of Elasticsearch BV.

Perfstatbeat is not endorsed by any of these companies.
