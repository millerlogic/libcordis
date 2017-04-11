package main

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"path"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"
)

/*
#cgo linux LDFLAGS: -ldl

#include <stdlib.h>
#include <string.h>
#include <errno.h>
#include "loadlib.h"

#define LIBCORDIS_PATH_TMP 1
#define LIBCORDIS_PATH_HOME 2
#define LIBCORDIS_PATH_EXE 3
#define LIBCORDIS_PATH_APP 4
#define LIBCORDIS_PATH_CONFIG 5
#define LIBCORDIS_PATH_DATA 6
#define LIBCORDIS_PATH_CACHE 7

#define LIBCORDIS_INIT_LOAD_FILE     0x0001 // Load from file path.
#define LIBCORDIS_INIT_LOAD_STRING   0x0002 // Load from direct contents.
#define LIBCORDIS_INIT_JSON_MANIFEST 0x0010 // Generic manifest file, look for the "libcordis" object.
#define LIBCORDIS_INIT_JSON_MAIN     0x0020 // This is already the main "libcordis" object.


#define LIBCORDIS_OPEN_NOT_FOUND (-ENOENT)

#define LIBCORDIS_OPEN_ERROR_INIT (-70001)
#define LIBCORDIS_OPEN_ERROR_WRONGKIND (-70002)
#define LIBCORDIS_OPEN_ERROR_UNABLE_LOAD (-70003)
#define LIBCORDIS_OPEN_ERROR_NEED_WRITE (-70004)

#define LIBCORDIS_INIT_ERROR_FLAGS (-80001)
#define LIBCORDIS_INIT_ERROR_MANIFEST_LOAD (-80002)
#define LIBCORDIS_INIT_ERROR_MANIFEST_DATA (-80003)
#define LIBCORDIS_INIT_ERROR_INIT (-80004)

#define LIBCORDIS_OPEN_WRITE     0x0001
#define _LIBCORDIS_OPEN_WANT     0x0F00
#define LIBCORDIS_OPEN_INTERFACE 0x0100
#define LIBCORDIS_OPEN_FS        0x0200

typedef const char *_const_string;
*/
import "C"

var initdir string
var interfacesDir string

func init() {
	xinitdir, _ := os.Getwd()
	initdir = xinitdir
}

type Paths interface {
	GetPathTmp() string
	GetPathHome() string
	GetPathExe() string
	GetPathApp() string
	GetPathConfig() string
	GetPathData() string
	GetPathCache() string
}

func cstrbuf(s string, dest *C.char, destlen C.size_t) C.size_t {
	slen := C.size_t(len(s))
	if slen >= destlen {
		if s == "" {
			return 0 // Don't need space for empty.
		}
		return slen + 1
	}
	cs := unsafe.Pointer(C.CString(s))
	C.memcpy(unsafe.Pointer(dest), cs, slen+1)
	C.free(cs)
	return slen
}

//export libcordis_get_path
func libcordis_get_path(which C.int, dest *C.char, destlen C.size_t) C.size_t {
	var path string
	switch which {
	case C.LIBCORDIS_PATH_TMP:
		path = paths.GetPathTmp()
	case C.LIBCORDIS_PATH_HOME:
		path = paths.GetPathHome()
	case C.LIBCORDIS_PATH_EXE:
		path = paths.GetPathExe()
	case C.LIBCORDIS_PATH_APP:
		path = paths.GetPathApp()
	case C.LIBCORDIS_PATH_CONFIG:
		path = paths.GetPathConfig()
	case C.LIBCORDIS_PATH_DATA:
		path = paths.GetPathData()
	case C.LIBCORDIS_PATH_CACHE:
		path = paths.GetPathCache()
	}
	return cstrbuf(path, dest, destlen)
}

type ManifestInterface struct {
	Library   string `json:"library"`
	Interface string `json:"interface"`
	NoUnload  bool   `json:"no_unload"`
}

type ManifestLaunch struct {
	Run string `json:"run"`
}

type ManifestLibcordis struct {
	Interfaces map[string]*ManifestInterface `json:"interfaces"`
	Launch     map[string]*ManifestLaunch    `json:"launch"`
}

type Manifest struct {
	Libcordis *ManifestLibcordis `json:"libcordis"`
}

var launch map[string]*ManifestLaunch
var interfaces map[string]*ManifestInterface

func isInit() bool {
	return interfaces != nil
}

func initLibFrom(flags int, arg string) int {
	if isInit() {
		return C.LIBCORDIS_INIT_ERROR_INIT
	}
	// Interfaces dir:
	appdir := paths.GetPathApp()
	ifdir := path.Join(appdir, "../interfaces")
	libdir := path.Join(appdir, "../lib")
	if _, err := os.Stat(ifdir); err == nil {
		interfacesDir = ifdir
	} else if _, err := os.Stat(libdir); err == nil {
		interfacesDir = libdir
	} else {
		interfacesDir = appdir
	}
	// Manifest:
	var manifestStream io.Reader
	if (flags & C.LIBCORDIS_INIT_LOAD_FILE) != 0 {
		manifestpath := arg
		if manifestpath == "" {
			manifestpath = paths.GetPathExe() + ".manifest.json"
		}
		fmanifest, err := os.Open(manifestpath)
		if err != nil {
			return C.LIBCORDIS_INIT_ERROR_MANIFEST_LOAD
		}
		defer fmanifest.Close()
		manifestStream = fmanifest
	} else if (flags & C.LIBCORDIS_INIT_LOAD_STRING) != 0 {
		manifestStream = strings.NewReader(arg)
	} else {
		return C.LIBCORDIS_INIT_ERROR_FLAGS
	}
	var manifestLibcordis *ManifestLibcordis
	if (flags & C.LIBCORDIS_INIT_JSON_MANIFEST) != 0 {
		var manifest Manifest
		decodeErr := json.NewDecoder(manifestStream).Decode(&manifest)
		if decodeErr != nil {
			return C.LIBCORDIS_INIT_ERROR_MANIFEST_LOAD
		}
		if manifest.Libcordis == nil {
			return C.LIBCORDIS_INIT_ERROR_MANIFEST_DATA
		}
		manifestLibcordis = manifest.Libcordis
	} else if (flags & C.LIBCORDIS_INIT_JSON_MAIN) != 0 {
		var main ManifestLibcordis
		decodeErr := json.NewDecoder(manifestStream).Decode(&main)
		if decodeErr != nil {
			return C.LIBCORDIS_INIT_ERROR_MANIFEST_LOAD
		}
		manifestLibcordis = &main
	} else {
		return C.LIBCORDIS_INIT_ERROR_FLAGS
	}
	// Validate:
	for lname, linfo := range manifestLibcordis.Launch {
		if lname == "" || linfo.Run == "" {
			return C.LIBCORDIS_INIT_ERROR_MANIFEST_DATA
		}
	}
	for ldep, depinfo := range manifestLibcordis.Interfaces {
		if ldep == "" || depinfo.Library == "" || depinfo.Interface == "" {
			return C.LIBCORDIS_INIT_ERROR_MANIFEST_DATA
		}
	}
	// Ok:
	launch = manifestLibcordis.Launch
	interfaces = manifestLibcordis.Interfaces
	return 0
}

// Set flags to 0.
// Returns 0 on success, or a LIBCORDIS_INIT_ERROR_* value.
//export libcordis_init
func libcordis_init(flags C.int) C.int {
	if (flags & 0x00FF) != 0 {
		return C.LIBCORDIS_INIT_ERROR_FLAGS
	}
	return C.int(initLibFrom(int(flags|C.LIBCORDIS_INIT_LOAD_FILE|C.LIBCORDIS_INIT_JSON_MANIFEST), ""))
}

//export libcordis_init_from
func libcordis_init_from(flags C.int, arg C._const_string) C.int {
	goarg := ""
	if arg != nil {
		goarg = C.GoString(arg)
	}
	return C.int(initLibFrom(int(flags), goarg))
}

func loadlib(path string) (uintptr, error) {
	cstr := C.CString(path)
	hlib := C._loadlib(cstr)
	C.free(unsafe.Pointer(cstr))
	if hlib == 0 {
		return 0, errors.New("Unable to load library")
	}
	return uintptr(hlib), nil
}

func unloadlib(hlib uintptr) {
	C._unloadlib(C.size_t(hlib))
}

func servelib(hlib uintptr, name string, sockfd int) int {
	cstr := C.CString(name)
	ret := C._servelib(C.size_t(hlib), cstr, C.int(sockfd))
	C.free(unsafe.Pointer(cstr))
	return int(ret)
}

func clientcount(hlib uintptr, name string) int {
	cstr := C.CString(name)
	ret := C._clientcount(C.size_t(hlib), cstr)
	C.free(unsafe.Pointer(cstr))
	return int(ret)
}

type Interface struct {
	hlib uintptr
	*ManifestInterface
	refcount uint32 // One per active known client, plus other references.
	manual   int32  // atomic boolean if refcount isn't the total.
}

func errorToErrno(err error) int {
	if perr, ok := err.(*os.PathError); ok {
		xerr := perr.Err
		if errno, ok := xerr.(syscall.Errno); ok {
			return int(errno)
		}
	} else if errno, ok := err.(syscall.Errno); ok {
		return int(errno)
	}
	return 0
}

func dupfd(f *os.File) int {
	fd, _, errno := syscall.Syscall(syscall.SYS_FCNTL, f.Fd(), syscall.F_DUPFD_CLOEXEC, 0)
	if errno != 0 {
		return int(-errno)
	}
	return int(fd)
}

var depsLock sync.Mutex
var depsLoaded = make(map[string]*Interface)

// Returns an open error (negative), or 0 on success.
// The returned lib has refcount incremented by 1 so that a cleanup can't sneak in.
func getDepLibLoad(npath string) (dep *Interface, openError int) {
	depsLock.Lock()
	defer depsLock.Unlock()
	dep, ok := depsLoaded[npath]
	if ok {
		atomic.AddUint32(&dep.refcount, 1)
		return dep, 0
	}
	depinfo := interfaces[npath]
	if depinfo != nil {
		libpath := depinfo.Library
		if !path.IsAbs(libpath) {
			libpath = path.Join(interfacesDir, libpath)
		}
		hlib, err := loadlib(libpath)
		if err != nil {
			return nil, C.LIBCORDIS_OPEN_ERROR_UNABLE_LOAD
		}
		dep = &Interface{hlib, depinfo, 1, 0} // refcount=1
		depsLoaded[npath] = dep
		return dep, 0
	}
	return nil, OPEN_NOT_FOUND
}

const OPEN_NOT_FOUND = -C.ENOENT

// When libcordis_open is called, the chain of functions are checked in order.
// Any chain function returns OPEN_NOT_FOUND it will continue to the next function.
// openLock is held while the chain functions are called.
var openChain = []func(path string, flags int) int{openInterface, openFS}
var openLock sync.RWMutex

func openInterface(path string, flags int) int {
	want := flags & C._LIBCORDIS_OPEN_WANT
	depinfo := interfaces[path]
	if depinfo == nil {
		return OPEN_NOT_FOUND
	}
	if want != 0 && (want&C.LIBCORDIS_OPEN_INTERFACE) == 0 {
		return C.LIBCORDIS_OPEN_ERROR_WRONGKIND
	}
	if (flags & C.LIBCORDIS_OPEN_WRITE) == 0 {
		return C.LIBCORDIS_OPEN_ERROR_NEED_WRITE
	}
	dep, openerr := getDepLibLoad(path) // Increments refcount by 1.
	if openerr != 0 {
		return openerr
	}
	defer atomic.AddUint32(&dep.refcount, ^uint32(0)) // decrement
	socktype := syscall.SOCK_STREAM | syscall.SOCK_CLOEXEC
	fds, err := syscall.Socketpair(syscall.AF_LOCAL, socktype, 0)
	if err != nil {
		errno := errorToErrno(err)
		if errno != 0 {
			return -errno
		}
		return OPEN_NOT_FOUND
	}
	atomic.AddUint32(&dep.refcount, 1)
	go func() {
		ret := servelib(dep.hlib, dep.Interface+"_interface", fds[0])
		if ret == 1 {
			atomic.StoreInt32(&dep.manual, 1)
		}
		atomic.AddUint32(&dep.refcount, ^uint32(0)) // decrement
		if ret < 0 {
			if ret == -2 {
				os.Stderr.WriteString("Interface '" + dep.Interface + "' not found\n")
			} else {
				os.Stderr.WriteString("Interface '" + dep.Interface + "' returned a failure\n")
			}
		}
	}()
	return fds[1]
}

func openFS(path string, flags int) int {
	fi, err := os.Stat(path)
	want := flags & C._LIBCORDIS_OPEN_WANT
	if want != 0 && (want&C.LIBCORDIS_OPEN_FS) == 0 {
		// User wants something that's not a FS file,
		// so check if the file exists to determine what to return.
		if err == nil {
			return C.LIBCORDIS_OPEN_ERROR_WRONGKIND
		}
	}
	if err != nil {
		errno := errorToErrno(err)
		if errno != 0 {
			return -errno
		}
		return OPEN_NOT_FOUND
	}
	if fi.IsDir() {
		return -C.EISDIR
	}
	var f *os.File
	if (fi.Mode() & os.ModeSocket) != 0 {
		if (flags & C.LIBCORDIS_OPEN_WRITE) == 0 {
			return C.LIBCORDIS_OPEN_ERROR_NEED_WRITE
		}
		var conn net.Conn
		conn, err = net.Dial("unix", path)
		if err != nil {
			errno := errorToErrno(err)
			if errno != 0 {
				return -errno
			}
			return OPEN_NOT_FOUND
		}
		defer conn.Close()
		if connf, ok := conn.(interface {
			File() (f *os.File, err error)
		}); ok {
			f, err = connf.File()
			if err != nil {
				return -C.EINVAL
			}
		} else {
			return C.LIBCORDIS_OPEN_ERROR_WRONGKIND
		}
	} else {
		openflag := os.O_RDONLY
		if (flags & C.LIBCORDIS_OPEN_WRITE) != 0 {
			openflag = os.O_RDWR
		}
		f, err = os.OpenFile(path, openflag, 0666)
		if err != nil {
			errno := errorToErrno(err)
			if errno != 0 {
				return -errno
			}
			return OPEN_NOT_FOUND
		}
	}
	defer f.Close()
	return dupfd(f)
}

func open(path string, flags int) int {
	if path == "" || path[0] != '/' {
		return OPEN_NOT_FOUND
	}
	openLock.Lock()
	defer openLock.Unlock()
	for _, openFunc := range openChain {
		ret := openFunc(path, flags)
		if ret != OPEN_NOT_FOUND {
			return ret
		}
	}
	return OPEN_NOT_FOUND
}

// Returns file descriptor on success, or a negative error value.
// Errors are: one of LIBCORDIS_OPEN_ERROR_*, or a negative errno value.
// Returns LIBCORDIS_OPEN_ERROR_WRONGKIND if a specific kind is requested but does not satisfy it.
//export libcordis_open
func libcordis_open(cpath C._const_string, flags C.int) C.int {
	if !isInit() {
		return C.LIBCORDIS_OPEN_ERROR_INIT
	}
	path := C.GoString(cpath)
	return C.int(open(path, int(flags)))
}

func cleanup() int {
	depsLock.Lock()
	defer depsLock.Unlock()
	result := 0
	for path, dep := range depsLoaded {
		if dep.NoUnload {
			continue
		}
		count := uint64(atomic.LoadUint32(&dep.refcount))
		//os.Stderr.WriteString("refcount = " + strconv.FormatInt(int64(count), 10) + "\n")
		if count == 0 {
			if atomic.LoadInt32(&dep.manual) != 0 {
				cc := clientcount(dep.hlib, dep.Interface+"_count")
				//os.Stderr.WriteString("interface_count = " + strconv.FormatInt(int64(cc), 10) + "\n")
				if cc < 0 {
					// Error getting count, so we can't assume there's no clients.
					count += uint64(1)
				} else {
					count += uint64(cc)
				}
			}
			if count == 0 {
				delete(depsLoaded, path)
				unloadlib(dep.hlib)
				result++
			}
		}
	}
	return result
}

// Unloads any interfaces and libraries that are no longer in use.
// An interface is considered not in use if it has no clients.
// If no_unload is specified for an interface, it, along with the library, will never unload.
// Returns how many interfaces were unloaded.
//export libcordis_cleanup
func libcordis_cleanup() C.int {
	return C.int(cleanup())
}

func main() {
}
