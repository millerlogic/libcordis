/* Created by "go tool cgo" - DO NOT EDIT. */

/* package github.com/millerlogic/libcordis */

/* Start of preamble from import "C" comments.  */


#line 17 "/home/ndev/workspace/go/src/github.com/millerlogic/libcordis/libcordis.go"



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

#define LIBCORDIS_INTERFACE_JSON_DATA 0x0001

typedef const char *_const_string;

#line 1 "cgo-generated-wrapper"


/* End of preamble from import "C" comments.  */


/* Start of boilerplate cgo prologue.  */
#line 1 "cgo-gcc-export-header-prolog"

#ifndef GO_CGO_PROLOGUE_H
#define GO_CGO_PROLOGUE_H

typedef signed char GoInt8;
typedef unsigned char GoUint8;
typedef short GoInt16;
typedef unsigned short GoUint16;
typedef int GoInt32;
typedef unsigned int GoUint32;
typedef long long GoInt64;
typedef unsigned long long GoUint64;
typedef GoInt64 GoInt;
typedef GoUint64 GoUint;
typedef __SIZE_TYPE__ GoUintptr;
typedef float GoFloat32;
typedef double GoFloat64;
typedef float _Complex GoComplex64;
typedef double _Complex GoComplex128;

/*
  static assertion to make sure the file is being used on architecture
  at least with matching size of GoInt.
*/
typedef char _check_for_64_bit_pointer_matching_GoInt[sizeof(void*)==64/8 ? 1:-1];

typedef struct { const char *p; GoInt n; } GoString;
typedef void *GoMap;
typedef void *GoChan;
typedef struct { void *t; void *v; } GoInterface;
typedef struct { void *data; GoInt len; GoInt cap; } GoSlice;

#endif

/* End of boilerplate cgo prologue.  */

#ifdef __cplusplus
extern "C" {
#endif


extern size_t libcordis_get_path(int p0, char* p1, size_t p2);

// Set flags to 0.
// Returns 0 on success, or a LIBCORDIS_INIT_ERROR_* value.

extern int libcordis_init(int p0);

extern int libcordis_init_from(int p0, _const_string p1);

// Returns file descriptor on success, or a negative error value.
// Errors are: one of LIBCORDIS_OPEN_ERROR_*, or a negative errno value.
// Returns LIBCORDIS_OPEN_ERROR_WRONGKIND if a specific kind is requested but does not satisfy it.

extern int libcordis_open(_const_string p0, int p1);

// Unloads any interfaces and libraries that are no longer in use.
// An interface is considered not in use if it has no clients.
// If no_unload is specified for an interface, it, along with the library, will never unload.
// Returns how many interfaces were unloaded.

extern int libcordis_cleanup();

// Call before fork.

extern void libcordis_before_fork();

// Call after a fork in the parent.

extern void libcordis_after_fork();

#ifdef __cplusplus
}
#endif
