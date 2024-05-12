package driver

import (
	"context"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"ycsi/pkg/mounter"
	"ycsi/pkg/s3"
)

type nodeServer struct {
	*csicommon.DefaultNodeServer
}

func (n *nodeServer) NodeStageVolume(ctx context.Context, request *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (n *nodeServer) NodeUnstageVolume(ctx context.Context, request *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (n *nodeServer) NodeExpandVolume(ctx context.Context, request *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (n *nodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	glog.V(6).Infof("%+v\n", req)
	volumeId := req.GetVolumeId()
	targetPath := req.GetTargetPath()
	// 临时挂载点
	stagingTargetPath := req.GetStagingTargetPath()
	bucketName, prefix := volumeIDToBucketPrefix(volumeId)
	if req.GetVolumeCapability() == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capability not provided")
	}
	if len(volumeId) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID not provided")
	}
	//if len(stagingTargetPath) == 0 {
	//	return nil, status.Error(codes.InvalidArgument, "Staging target path not provided")
	//}
	if len(targetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path not provided")
	}
	// 检查是不是挂载点
	notMnt, err := checkMount(targetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !notMnt {
		return &csi.NodePublishVolumeResponse{}, nil
	}
	deviceId := ""
	if req.GetVolumeContext() != nil {
		deviceId = req.GetVolumeContext()["device_id"]
	}
	readOnly := req.GetReadonly()
	attrib := req.GetVolumeContext()
	mountFlags := req.GetVolumeCapability().GetMount().GetMountFlags()
	glog.V(6).Infof("target %v\ndevice %v\nreadonly %v\nvolumeId %v\nattributes %v\nmountflags %v\n",
		targetPath, deviceId, readOnly, volumeId, attrib, mountFlags)

	client, err := s3.NewClientFromSecret(req.GetSecrets())
	if err != nil {
		return nil, status.Error(codes.Internal, "连接s3失败")
	}
	meta, err := client.GetFSMeta(bucketName, prefix)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	mounter := mounter.NewMounter(meta, client.Config)
	if err := mounter.Mount(stagingTargetPath, targetPath); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	glog.V(6).Infof("卷: %s成功挂载到目标:%s", volumeId, targetPath)
	return &csi.NodePublishVolumeResponse{}, nil
}

func (n *nodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	volumeId := req.GetVolumeId()
	targetPath := req.GetTargetPath()
	if len(volumeId) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID not provided")
	}
	if len(targetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path not provided")
	}
	if err := mounter.FuseUnmount(targetPath); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	glog.V(6).Infof("卷: %s成功取消挂载", volumeId)
	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func newNodeServer(d *csicommon.CSIDriver) *nodeServer {
	return &nodeServer{DefaultNodeServer: csicommon.NewDefaultNodeServer(d)}
}
