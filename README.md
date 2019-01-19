# perfstatbeat

perfstatbeat is a beat based on metricbeat which collects performance metrics through the perfstat API of the AIX operating system and supplies them in a form, that is mostly compatible with the metricbeat system module.

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

- An AIX operating system instance. I'm using version 7.1. Support for Go should be even better on version 7.2 but I haven't try it yet.
- A recent version of `gcc-go` and its dependencies. I'm using the packages kindly provided by [BullFreeware](http://www.bullfreeware.com/search.php?package=gcc-go).
- Various open source packages, at least GNU `make`. Some tasks also require `find-utils` and Python. I'm using the packages kindly provided by [Michael Perzl](http://www.perzl.org/aix/).

### Runtime requirements

- `libgo` should be the same as in the build environment.
- `libgcc`, not necessarily at the same level or from the same source like `libgo`.

### Compiling and running

To compile your beat run `gmake`. Then you can run the following command to see the first output:

```
./perfstatbeat -e -d "*"
```

## Development

perfstatbeat is a beat based on metricbeat which was generated with metricbeat/metricset generator.

In case further modules are metricsets should be added, run:

```
make create-metricset
```

After updates to the fields or config files, always run

```
make collect
```

This updates all fields and docs with the most recent changes.

## Vendoring

perfstatbeat currently includes version 6.2.3 of beats in the `vendor` subfolder. Later version make use of Go modules that currently don't run on AIX.

## Packaging

The original packaging process makes use of containers and will obviously not work here. So currently the binary file has to be distributed manually. Maybe an RPM spec file will be provided later.

## Disclaimer

AIX is a registered trademark of the International Business Machines Corporation.

Metricbeat is a trademark of Elasticsearch BV.

perfstatbeat is not endorsed by any of these companies.
