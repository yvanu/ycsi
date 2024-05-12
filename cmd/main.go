package main

import (
	"flag"
	"ycsi/pkg/driver"
)

var (
	endpoint = flag.String("endpoint", "unix://tmp/csi.sock", "csi endpoint")
	nodeId   = flag.String("nodeid", "", "node id")
)

func main() {
	flag.Parse()
	d := driver.NewCSIDriver(*nodeId, *endpoint)
	d.Run()
}
