package driver

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"k8s.io/utils/mount"
	"os"
	"strings"
)

func sanitizeVolumeID(volumeID string) string {
	volumeID = strings.ToLower(volumeID)
	if len(volumeID) > 63 {
		h := sha1.New()
		io.WriteString(h, volumeID)
		volumeID = hex.EncodeToString(h.Sum(nil))
	}
	return volumeID
}

func volumeIDToBucketPrefix(volumeID string) (string, string) {
	splitVolumeID := strings.Split(volumeID, "/")
	if len(splitVolumeID) > 1 {
		return splitVolumeID[0], splitVolumeID[1]
	}

	return volumeID, ""
}

func checkMount(targetPath string) (bool, error) {
	notMnt, err := mount.New("").IsLikelyNotMountPoint(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(targetPath, 0750); err != nil {
				return false, err
			}
			notMnt = true
		} else {
			return false, err
		}
	}
	return notMnt, nil
}
