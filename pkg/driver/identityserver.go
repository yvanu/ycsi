package driver

import "github.com/kubernetes-csi/drivers/pkg/csi-common"

type identityServer struct {
	*csicommon.DefaultIdentityServer
}

func newIdentityServer(d *csicommon.CSIDriver) *identityServer {
	return &identityServer{
		DefaultIdentityServer: csicommon.NewDefaultIdentityServer(d),
	}
}
