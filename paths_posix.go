package main

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"
)

type posixPaths struct {
}

var _env []string
var _envOnce sync.Once

func _getenv(key string) string {
	x := os.Getenv(key)
	if x != "" {
		return x
	}
	// Work around shared object loader issue with musl.
	// https://github.com/golang/go/issues/13492
	_envOnce.Do(func() {
		f, err := os.Open("/proc/self/environ")
		if err == nil {
			all, err := ioutil.ReadAll(f)
			if err == nil {
				_env = strings.Split(string(all), "\000")
			}
		}
	})
	for _, e := range _env {
		if len(e) > len(key) && e[len(key)] == '=' {
			if strings.HasPrefix(e, key) {
				return e[len(key)+1:]
			}
		}
	}
	return ""
}

func (p *posixPaths) GetPathTmp() string {
	return os.TempDir()
}

func (p *posixPaths) GetPathHome() string {
	x := _getenv("HOME")
	if x == "" {
		//os.Stderr.WriteString("HOME env var is empty\n")
		return initdir
	}
	return x
}

func (p *posixPaths) GetPathExe() string {
	x, _ := os.Executable()
	return x
}

func (p *posixPaths) GetPathApp() string {
	exe := p.GetPathExe()
	if exe == "" {
		return initdir
	}
	return path.Dir(exe)
}

func (p *posixPaths) GetPathConfig() string {
	x := _getenv("XDG_CONFIG_HOME")
	if x == "" {
		return path.Join(p.GetPathHome(), ".config")
	}
	return x
}

func (p *posixPaths) GetPathData() string {
	x := _getenv("XDG_DATA_HOME")
	if x == "" {
		return path.Join(p.GetPathHome(), ".local/share")
	}
	return x
}

func (p *posixPaths) GetPathCache() string {
	x := _getenv("XDG_CONFIG_HOME")
	if x == "" {
		return path.Join(p.GetPathHome(), ".cache")
	}
	return x
}

var paths Paths = &posixPaths{}
