package driver

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
)

type driver struct {
	csiDriver *csicommon.CSIDriver
	endpoint  string

	ids *identityServer
	cs  *controllerServer
	ns  *nodeServer
}

func NewCSIDriver(nodeID, endpoint string) *driver {
	csiDriver := csicommon.NewCSIDriver(driverName, version, nodeID)
	csiDriver.AddControllerServiceCapabilities(
		[]csi.ControllerServiceCapability_RPC_Type{
			csi.ControllerServiceCapability_RPC_LIST_VOLUMES,
			csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
			csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
		})
	csiDriver.AddVolumeCapabilityAccessModes([]csi.VolumeCapability_AccessMode_Mode{csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER})
	driver := &driver{endpoint: endpoint, csiDriver: csiDriver}
	return driver
}

func (d *driver) Run() {
	d.ids = newIdentityServer(d.csiDriver)
	d.cs = newControllerServer(d.csiDriver)
	d.ns = newNodeServer(d.csiDriver)

	s := csicommon.NewNonBlockingGRPCServer()
	s.Start(d.endpoint, d.ids, d.cs, d.ns)
	s.Wait()
}
