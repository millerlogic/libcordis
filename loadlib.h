#ifndef _LOADLIB_H_
#define _LOADLIB_H_

#include <stdlib.h>


size_t _loadlib(const char *path);

void _unloadlib(size_t hlib);

int _servelib(size_t hlib, const char *name, int sockfd);

int _clientcount(size_t hlib, const char *name);


#endif
