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
            "/interfaces/myinterface": {
                "library": "myplugin.so",
                "interface": "myinterface"
            }
        }
    }
}
```

* libcordis: the main libcordis configuration
* interfaces: a map of plugins. The keys are the libcordis paths to open the plugin, this can be anything as long as nothing conflicts, but /interfaces/*interface-name* is recommended. The values are interface definitions, explained next.
* *interface definition*
  * library: the path to the plugin's shared object. This is relative to the app's interfaces directory, lib directory, or bin directory, depending on which exist.
  * interface: is the interface in the shared object to load, as explained in the below plugin API.

An app opens interfaces to read and write in order to communicate with the plugin.

## libcordis C API

This is the API to be used by the application in order to load plugins. Include ```libcordis.h```

## Plugin C API

This is the API in the plugin shared object. *TBD*

```
int libcordis_init(int flags)
```
Initialize libcordis. Set flags to 0.

```
libcordis_open(const char *path, int flags)
```
Open a libcordis plugin. The path is a libcordis plugin path, which are the keys in the libcordis.interfaces map from the manifest file. The path can also be a local file system path to open a file.
Flags:
```
LIBCORDIS_OPEN_WRITE - request write permission.
LIBCORDIS_OPEN_INTERFACE - only open interfaces, do not open other things.
LIBCORDIS_OPEN_FS - only open file system files, do not open other things.
```
Returns a file descriptor to the interface, or a negative error value:
```
LIBCORDIS_OPEN_ERROR_INIT - did not initialize libcordis.
LIBCORDIS_OPEN_ERROR_WRONGKIND - wrong kind of path, when LIBCORDIS_OPEN_INTERFACE/FS used.
LIBCORDIS_OPEN_ERROR_UNABLE_LOAD - the library failed to load.
LIBCORDIS_OPEN_ERROR_NEED_WRITE - write access is required, such as for interfaces.
LIBCORDIS_OPEN_NOT_FOUND (-ENOENT) - path not found.
-errno - otherwise a negative errno value can be returned.
```
If a file descriptor is returned, it must be closed when no longer in use.

```
void libcordis_cleanup()
```
Attempt to clean up unused interfaces and libraries. Can be called at any point, it will not clean anything still in use.
If a library should not ever be unloaded for some reason, no_unload can be set to true in the interface definition.

```
size_t libcordis_get_path(int which, char* buf, size_t buflen)
```
Helper function to get various useful paths for the platform.
which:
```
LIBCORDIS_PATH_TMP - temporary directory, such as /tmp
LIBCORDIS_PATH_HOME - user's home directory, or the current dir if none
LIBCORDIS_PATH_EXE - full path to the executable
LIBCORDIS_PATH_APP - path to the app's directory
LIBCORDIS_PATH_CONFIG - config dir, such as ~/.config
LIBCORDIS_PATH_DATA - data dir, such as ~/.local/share
LIBCORDIS_PATH_CACHE - cache dir, such as ~/.cache
```
buf is the buffer to copy the path into.
buflen is the length of bytes at buf.
Returns number of bytes in the returned string; otherwise the number of bytes needed to hold the path. 0 is returned on failure.

## Build

```sh
go build -buildmode=c-shared \
    -ldflags '-s -w' \
    -o libcordis.so github.com/millerlogic/libcordis
```
