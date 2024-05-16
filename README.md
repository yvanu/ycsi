# ycsi
- 总览：

本质上是实现一个grpc服务，该服务需要使用以下三个接口。

   - IdentityServer
```go
// IdentityServer is the server API for Identity service.
type IdentityServer interface {
	GetPluginInfo(context.Context, *GetPluginInfoRequest) (*GetPluginInfoResponse, error)
	GetPluginCapabilities(context.Context, *GetPluginCapabilitiesRequest) (*GetPluginCapabilitiesResponse, error)
	Probe(context.Context, *ProbeRequest) (*ProbeResponse, error)
}
```

   - ControllerServer
```go
// ControllerServer is the server API for Controller service.
type ControllerServer interface {
	CreateVolume(context.Context, *CreateVolumeRequest) (*CreateVolumeResponse, error)
	DeleteVolume(context.Context, *DeleteVolumeRequest) (*DeleteVolumeResponse, error)
	ControllerPublishVolume(context.Context, *ControllerPublishVolumeRequest) (*ControllerPublishVolumeResponse, error)
	ControllerUnpublishVolume(context.Context, *ControllerUnpublishVolumeRequest) (*ControllerUnpublishVolumeResponse, error)
	ValidateVolumeCapabilities(context.Context, *ValidateVolumeCapabilitiesRequest) (*ValidateVolumeCapabilitiesResponse, error)
	ListVolumes(context.Context, *ListVolumesRequest) (*ListVolumesResponse, error)
	GetCapacity(context.Context, *GetCapacityRequest) (*GetCapacityResponse, error)
	ControllerGetCapabilities(context.Context, *ControllerGetCapabilitiesRequest) (*ControllerGetCapabilitiesResponse, error)
	CreateSnapshot(context.Context, *CreateSnapshotRequest) (*CreateSnapshotResponse, error)
	DeleteSnapshot(context.Context, *DeleteSnapshotRequest) (*DeleteSnapshotResponse, error)
	ListSnapshots(context.Context, *ListSnapshotsRequest) (*ListSnapshotsResponse, error)
	ControllerExpandVolume(context.Context, *ControllerExpandVolumeRequest) (*ControllerExpandVolumeResponse, error)
	ControllerGetVolume(context.Context, *ControllerGetVolumeRequest) (*ControllerGetVolumeResponse, error)
	ControllerModifyVolume(context.Context, *ControllerModifyVolumeRequest) (*ControllerModifyVolumeResponse, error)
}
```

   - NodeServer
```go
// NodeServer is the server API for Node service.
type NodeServer interface {
	NodeStageVolume(context.Context, *NodeStageVolumeRequest) (*NodeStageVolumeResponse, error)
	NodeUnstageVolume(context.Context, *NodeUnstageVolumeRequest) (*NodeUnstageVolumeResponse, error)
	NodePublishVolume(context.Context, *NodePublishVolumeRequest) (*NodePublishVolumeResponse, error)
	NodeUnpublishVolume(context.Context, *NodeUnpublishVolumeRequest) (*NodeUnpublishVolumeResponse, error)
	NodeGetVolumeStats(context.Context, *NodeGetVolumeStatsRequest) (*NodeGetVolumeStatsResponse, error)
	NodeExpandVolume(context.Context, *NodeExpandVolumeRequest) (*NodeExpandVolumeResponse, error)
	NodeGetCapabilities(context.Context, *NodeGetCapabilitiesRequest) (*NodeGetCapabilitiesResponse, error)
	NodeGetInfo(context.Context, *NodeGetInfoRequest) (*NodeGetInfoResponse, error)
}
```

- 实现

csicommon包帮我们封装了一部分功能。<br />identityServer直接使用了ciscommon提供的DefaultIdentityServer。<br />controllerServer实现了createVolume,deleteVolume,validateVolumeCapablitites三个方法。

   - createVolume主要功能是在腾讯云上创建bucket/prefix
   - deleteVolume是上述反操作
   - validateVolumeCapablitites用于验证创建请求相关参数是否符合要求

nodeServer实现了nodePublishVolume,nodeUnpublishVolume两个方法

   - nodePublishVolume主要功能是将腾讯云上的目录挂载到节点上
   - nodeUnpublishVolume是上述反操作

实现了上面功能后相当于实现了一个简易的可以操作腾讯云的csi。

- 部署
   - provisioner
      - 监听PVC对象，判断是否需要动态创建存储卷：
         - 检查pvc的注解中的provisioner是否是自身
         - 调用createVolume接口，创建PV
      - 监听PV对象，如果需要删除则调用deleteVolume接口
   - attacher
      - 监听volumeAttachment对象，获取pv信息，
   - daemonset
      - 把ycsi通过daemonset形式部署到各个节点
- example
   - 创建storageclass
   - 创建pvc
   - 创建pod

