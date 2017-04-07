# libcordis
libcordis is a simple yet flexible plugin architecture.
Currently written in golang, though will eventually be rewritten in C to minimize overhead and dependencies.
The API is all C, there are no other dependencies once the shared library is built.

## Usage:

The plugins are defined in a manifest named *appname*.manifest.json, loaded from the app's binary path, with the following structure:

```json
{
    "libcordis": {
        "interfaces": {
            "/interfaces/myplugin": {
                "library": "myplugin.so",
                "interface": "myplugin"
            }
        }
    }
}
```

* libcordis: the main libcordis configuration
* interfaces: a map of plugins, the keys are the libcordis paths to open the plugin, and the values are interface definitions.
* *interface definition*
  * library: the path to the plugin's shared object. This is relative to the app's interfaces directory, lib directory, or bin directory, depending on which exist.
  * interface: is the interface in the shared object to load, as explained in the below plugin API.

## libcordis API

*TBD*

## Plugin API

This is the API in the plugin shared object. *TBD*

## Build

```sh
go build -buildmode=c-shared \
    -ldflags '-s -w' \
    -o libcordis.so github.com/millerlogic/libcordis
```
