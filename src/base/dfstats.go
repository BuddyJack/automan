package base

import (
	"io/ioutil"
	"bufio"
	"bytes"
	"io"
	"strings"
	"unsafe"
	"syscall"
	"../model"
	"strconv"
)

func listMountPoint() ([][3]string, error) {
	contents, err := ioutil.ReadFile("/proc/mounts")
	if err != nil {
		return nil, err
	}

	ret := make([][3]string, 0)
	reader := bufio.NewReader(bytes.NewBuffer(contents))
	for {
		line, err := reader.ReadBytes('\n')
		if io.EOF == err {
			err = nil
			break
		} else if err != nil {
			return nil, err
		}
		fields := strings.Fields(*(*string)(unsafe.Pointer(&line)))
		fs_spec := fields[0]
		fs_file := fields[1]
		fs_vfstype := fields[2]
		if _, exist := FSSPEC_IGNORE[fs_spec]; exist {
			continue
		}

		if _, exist := FSTYPE_IGNORE[fs_vfstype]; exist {
			continue
		}

		if strings.HasPrefix(fs_vfstype, "fuse") {
			continue
		}

		if IgnoreFsFile(fs_file) {
			continue
		}
		//replace one device mounted by '/home/****' and '/home', choose shorter one
		if strings.HasPrefix(fs_spec, "/dev") {
			deviceFound := false
			for idx := range ret {
				if ret[idx][0] == fs_spec {
					deviceFound = true
					if len(fs_file) < len(ret[idx][1]) {
						ret[idx][1] = fs_file
					}
				}
			}
			if !deviceFound {
				ret = append(ret, [3]string{fs_spec, fs_file, fs_vfstype})
			}
		} else {
			ret = append(ret, [3]string{fs_spec, fs_file, fs_vfstype})
		}
	}
	return ret, nil
}

func buildDeviceUsage(_fs_spec, _fs_file, _fs_type string) (*DeviceUsage, error) {
	ret := &DeviceUsage{FsSpec: _fs_spec, FsFile: _fs_file, FsVfstype: _fs_type}
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(_fs_file, &fs)
	if nil != err {
		return nil, err
	}
	used := fs.Blocks - fs.Bfree
	ret.BlocksAll = uint64(fs.Bsize) * fs.Blocks
	ret.BlocksFree = uint64(fs.Bsize) * fs.Bfree
	ret.BlocksUsed = uint64(fs.Bsize) * used
	if 0 == fs.Blocks {
		ret.BlocksFreePercent = float64(0)
		ret.BlocksUsedPercent = float64(0)
	} else {
		ret.BlocksUsedPercent = float64(used) * 100.0 / float64(used+fs.Bavail)
		ret.BlocksFreePercent = 100.0 - ret.BlocksUsedPercent
	}
	return ret, nil
}

func (deviceUsage *DeviceUsage) Metrics() (metrics []*model.MetricValue) {
	mountPoints, err := listMountPoint()
	if nil != err {
		return nil
	}
	for idx := range mountPoints {
		du, err := buildDeviceUsage(mountPoints[idx][0], mountPoints[idx][1], mountPoints[idx][2])
		if nil != err {
			continue
		}
		metrics = append(metrics, &model.MetricValue{Endpoint: "df", Metric: "df.available.percent", Value: strconv.FormatFloat(du.BlocksFreePercent, 'f', 2, 64), Tags: map[string]string{"mount": du.FsFile,}})
	}
	return
}
