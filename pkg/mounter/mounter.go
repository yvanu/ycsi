package mounter

import (
	"errors"
	"fmt"
	"github.com/golang/glog"
	"k8s.io/utils/mount"
	"os"
	"os/exec"
	"time"
	"ycsi/pkg/s3"
)

type Mounter interface {
	Mount(source, target string) error
}

func NewMounter(meta *s3.FSMeta, config *s3.Config) Mounter {
	return newS3fsMounter(meta, config)
}

func fuseMount(path string, command string, args []string) error {
	cmd := exec.Command(command, args...)
	glog.V(3).Infof("Mounting fuse with command: %s and args: %s", command, args)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Error fuseMount command: %s\nargs: %s\noutput", command, args)
	}

	return waitForMount(path, 10*time.Second)
}

func FuseUnmount(path string) error {
	if err := mount.New("").Unmount(path); err != nil {
		return err
	}
	// todo 加上检测是否取消挂载完成
	return nil
}

func waitForMount(path string, timeout time.Duration) error {
	var elapsed time.Duration
	var interval = 10 * time.Millisecond
	for {
		notMount, err := mount.New("").IsLikelyNotMountPoint(path)
		if err != nil {
			return err
		}
		if !notMount {
			return nil
		}
		time.Sleep(interval)
		elapsed = elapsed + interval
		if elapsed >= timeout {
			return errors.New("timeout waiting for mount")
		}
	}
}
