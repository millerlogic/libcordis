#define _GNU_SOURCE
#include <dlfcn.h>

#include <stdio.h>

#include "loadlib.h"


size_t _loadlib(const char *path)
{
    void *hlib = dlopen(path, RTLD_LAZY);
    if(!hlib)
        fputs(dlerror(), stderr);
    return (size_t)hlib;
}


void _unloadlib(size_t hlib)
{
    dlclose((void*)hlib);
}


int _servelib(size_t hlib, const char *name, int sockfd, int flags, const char *data)
{
    int (*ifacefunc)(int sockfd, int flags, const char *data) = dlsym((void*)hlib, name);
    if(!ifacefunc)
        return -2;
    int ret = ifacefunc(sockfd, flags, data);
    if(ret == -2)
        return -1;
    return ret;
}

int _clientcount(size_t hlib, const char *name)
{
    int (*ccfunc)(void) = dlsym((void*)hlib, name);
    if(!ccfunc)
        return -2;
    int ret = ccfunc();
    if(ret == -2)
        return -1;
    return ret;
}

