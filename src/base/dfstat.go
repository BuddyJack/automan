package base

import "strings"

var FSSPEC_IGNORE = map[string]struct{}{
	"none":  struct{}{},
	"nodev": struct{}{},
}

var FSTYPE_IGNORE = map[string]struct{}{
	"cgroup":     struct{}{},
	"debugfs":    struct{}{},
	"devpts":     struct{}{},
	"devtmpfs":   struct{}{},
	"rpc_pipefs": struct{}{},
	"rootfs":     struct{}{},
}

var FSFILE_PREFIX_IGNORE = []string{
	"/sys",
	"/net",
	"/misc",
	"/proc",
	"/lib",
}

func IgnoreFsFile(fs_file string) bool {
	for _,prefix := range FSFILE_PREFIX_IGNORE{
		if strings.HasPrefix(fs_file,prefix){
			return true
		}
	}
	return false
}

type DeviceUsage struct {
	FsSpec            string
	FsFile            string
	FsVfstype         string
	BlocksAll         uint64
	BlocksUsed        uint64
	BlocksFree        uint64
	BlocksUsedPercent float64
	BlocksFreePercent float64
}
