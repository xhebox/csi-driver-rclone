/*
MIT License

Copyright (c) 2023 xhe

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package driver

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	"golang.org/x/net/context"
)

func (d *driver) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	resp := &csi.NodeGetInfoResponse{
		NodeId: d.config.NodeID,
	}
	return resp, nil
}

func (d *driver) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	return &csi.NodeStageVolumeResponse{}, nil
}

func (d *driver) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	return &csi.NodeUnstageVolumeResponse{}, nil
}

func (d *driver) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	if _, e := os.Stat(req.TargetPath); e != nil {
		if errors.Is(e, os.ErrNotExist) {
			e = os.MkdirAll(req.TargetPath, 0755)
		}
		if e != nil {
			return &csi.NodePublishVolumeResponse{}, e
		}
	}
	args := []string{"mount"}
	p := "/"
	if e, ok := req.VolumeContext["path"]; ok {
		p = e
	}
	args = append(args, fmt.Sprintf(":%s:%s", req.VolumeContext["type"], p), req.TargetPath)
	if req.Readonly {
		args = append(args, "--read-only")
	}
	for k, v := range req.VolumeContext {
		if k == "path" || k == "type" || k == "wait" {
			continue
		}
		args = append(args, "--"+k, v)
	}
	var b *bytes.Buffer
	var err error
	ch := make(chan struct{})
	go func() {
		b, err = d.exec(args...)
		if err != nil {
			err = fmt.Errorf("err: %+v\nout: %s", err, b.String())
		}
		ch <- struct{}{}
	}()
	wait, ve := time.ParseDuration(req.VolumeContext["wait"])
	if ve != nil {
		wait = 10 * time.Second
	}
	select {
	case <-ch:
		glog.V(5).Infof("publish volume: %s", err)
		return &csi.NodePublishVolumeResponse{}, err
	case <-time.After(wait):
		return &csi.NodePublishVolumeResponse{}, nil
	}
}

func (d *driver) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	var err error
	b := &bytes.Buffer{}
	if _, e := os.Stat(req.TargetPath); e == nil {
		cmd := exec.Command("umount", req.TargetPath)
		cmd.Stdout = b
		cmd.Stderr = b
		err = cmd.Run()
	}
	if err != nil {
		err = fmt.Errorf("err: %+v\nout: %s", err, b.String())
		glog.V(5).Infof("unpublish volume: %s", err)
	}
	return &csi.NodeUnpublishVolumeResponse{}, err
}

func (d *driver) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	cl := []csi.NodeServiceCapability_RPC_Type{}
	var caps []*csi.NodeServiceCapability
	for _, c := range cl {
		caps = append(caps, &csi.NodeServiceCapability{
			Type: &csi.NodeServiceCapability_Rpc{
				Rpc: &csi.NodeServiceCapability_RPC{
					Type: c,
				},
			},
		})
	}
	return &csi.NodeGetCapabilitiesResponse{Capabilities: caps}, nil
}

func (d *driver) NodeGetVolumeStats(ctx context.Context, in *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	return nil, fmt.Errorf("unsupported")
}

// NodeExpandVolume is only implemented so the driver can be used for e2e testing.
func (hp *driver) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return nil, fmt.Errorf("unsupported")
}
