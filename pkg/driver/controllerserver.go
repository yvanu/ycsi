package driver

import (
	"context"
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"path"
	"ycsi/pkg/s3"
)

// 在远端创建卷 这里先实现的腾讯云上创建目录
type controllerServer struct {
	*csicommon.DefaultControllerServer
}

func (c *controllerServer) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	return &csi.ControllerPublishVolumeResponse{}, nil
}

func (c *controllerServer) ControllerUnpublishVolume(ctx context.Context, request *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	return &csi.ControllerUnpublishVolumeResponse{}, nil
}

func (c *controllerServer) ControllerExpandVolume(ctx context.Context, request *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c *controllerServer) ControllerGetVolume(ctx context.Context, request *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c *controllerServer) ControllerModifyVolume(ctx context.Context, request *csi.ControllerModifyVolumeRequest) (*csi.ControllerModifyVolumeResponse, error) {
	//TODO implement me
	panic("implement me")
}

func newControllerServer(d *csicommon.CSIDriver) *controllerServer {
	return &controllerServer{DefaultControllerServer: csicommon.NewDefaultControllerServer(d)}
}

func (c *controllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	params := req.GetParameters()
	capacityBytes := req.GetCapacityRange().GetRequiredBytes()
	mounterType := params["mounter"] // s3fs
	volumeID := sanitizeVolumeID(req.GetName())

	bucketName := params["bucket"]
	prefix := volumeID
	volumeID = path.Join(bucketName, prefix)

	if err := c.Driver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		glog.V(6).Info("没有创建删除卷的能力")
		return nil, err
	}
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID is required")
	}
	if req.GetVolumeCapabilities() == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capabilities are required")
	}
	glog.V(6).Infof("收到创建volume: %s的请求", volumeID)
	glog.V(6).Infof("%+v\n", req)
	meta := &s3.FSMeta{
		BucketName:    bucketName,
		Prefix:        prefix,
		Mounter:       mounterType,
		CapacityBytes: capacityBytes,
		FSPath:        defaultFsPath,
	}
	// 创建s3 client
	client, err := s3.NewClientFromSecret(req.GetSecrets())
	if err != nil {
		return nil, status.Error(codes.Internal, "创建s3客户端失败")
	}
	exist, err := client.BucketExists(bucketName)
	if err != nil {
		glog.V(6).Infof(fmt.Sprintf("判断bucket是否存在失败, err: %s", err.Error()))
		return nil, status.Error(codes.Internal, fmt.Sprintf("判断bucket是否存在失败, err: %s", err.Error()))
	}
	if exist {
		meta, err := client.GetFSMeta(bucketName, prefix)
		if err == nil {
			if capacityBytes > meta.CapacityBytes {
				return nil, status.Error(codes.InvalidArgument, "存在相同的volume， 但是大小不一样")
			}
		}
	} else {
		if err := client.CreateBucket(bucketName); err != nil {
			return nil, status.Error(codes.Internal, "创建bucket失败")
		}
	}
	if err = client.CreatePrefix(bucketName, path.Join(prefix, defaultFsPath)); err != nil && prefix != "" {
		return nil, status.Error(codes.Internal, fmt.Sprintf("创建目录:%s失败, %v", bucketName, err))
	}
	if err = client.SetFSMeta(meta); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("设置bucket的元信息失败"))
	}
	glog.V(6).Infof("创建volume: %s", volumeID)
	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      volumeID,
			VolumeContext: req.GetParameters(),
			CapacityBytes: capacityBytes,
		},
	}, nil
}

func (c *controllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	volumeId := req.GetVolumeId()
	bucketName, prefix := volumeIDToBucketPrefix(volumeId)
	if len(volumeId) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID is required")
	}
	if err := c.Driver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		glog.V(6).Info("没有创建删除卷的能力")
		return nil, err
	}
	glog.V(6).Infof("删除卷: %s", volumeId)
	client, err := s3.NewClientFromSecret(req.GetSecrets())
	if err != nil {
		return nil, status.Error(codes.Internal, "创建s3客户端失败")
	}
	meta, err := client.GetFSMeta(bucketName, prefix)
	if err != nil {
		glog.V(6).Infof("fsmeta文件不存在，忽略删除请求")
		return &csi.DeleteVolumeResponse{}, nil
	}
	var deleteErr error
	// 不做删除bucket的操作
	if prefix != "" {
		err = client.RemovePrefix(bucketName, prefix)
		if err != nil {
			deleteErr = fmt.Errorf("删除prefix失败: %s", err.Error())
		}
	}
	if deleteErr != nil {
		// 删除失败了，确保fsmeta存在，否则会脱离管控
		err = client.SetFSMeta(meta)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("设置bucket的元信息失败"))
		}
		return nil, deleteErr
	}
	return &csi.DeleteVolumeResponse{}, nil
}

func (c *controllerServer) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	volumeId := req.GetVolumeId()
	if len(volumeId) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID is required")
	}
	if req.GetVolumeCapabilities() == nil {
		return nil, status.Error(codes.InvalidArgument, "volume capabilities is required")
	}

	// 验证请求中一些文件参数是否正确
	bucketName, prefix := volumeIDToBucketPrefix(req.GetVolumeId())

	client, err := s3.NewClientFromSecret(req.GetSecrets())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize S3 client: %s", err)
	}
	exists, err := client.BucketExists(bucketName)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("bucket of volume with id %s does not exist", req.GetVolumeId()))
	}

	if _, err := client.GetFSMeta(bucketName, prefix); err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("fsmeta of volume with id %s does not exist", req.GetVolumeId()))
	}

	supportedAccessMode := &csi.VolumeCapability_AccessMode{
		Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
	}
	for _, cap := range req.GetVolumeCapabilities() {
		if cap.GetAccessMode() != supportedAccessMode {
			return &csi.ValidateVolumeCapabilitiesResponse{
				Message: "只支持单节点写入",
			}, nil
		}
	}
	return &csi.ValidateVolumeCapabilitiesResponse{
		Confirmed: &csi.ValidateVolumeCapabilitiesResponse_Confirmed{
			VolumeCapabilities: []*csi.VolumeCapability{
				{
					AccessMode: supportedAccessMode,
				},
			},
		},
	}, nil
}
