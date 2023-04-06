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
	"errors"
	"fmt"
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	"golang.org/x/net/context"
)

func (d *driver) ListVolumes(ctx context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	vols := &csi.ListVolumesResponse{
		Entries: []*csi.ListVolumesResponse_Entry{},
	}
	rs, err := d.remoteList()
	if err != nil {
		return nil, err
	}
	for _, r := range rs {
		ri, err := d.remoteAbout(r, "")
		if err != nil {
			return nil, err
		}
		vols.Entries = append(vols.Entries, &csi.ListVolumesResponse_Entry{
			Volume: &csi.Volume{
				VolumeId:      r,
				CapacityBytes: ri.Get("total").Int(),
			},
		})
	}
	glog.V(5).Infof("Volumes are: %+v", *vols)
	return vols, nil
}

func (d *driver) ControllerGetVolume(ctx context.Context, req *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	ri, err := d.remoteAbout(req.VolumeId, "")
	if err != nil {
		return nil, err
	}
	vol := &csi.ControllerGetVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      req.VolumeId,
			CapacityBytes: ri.Get("total").Int(),
		},
		Status: &csi.ControllerGetVolumeResponse_VolumeStatus{},
	}
	glog.V(5).Infof("Volume is: %+v", *vol)
	return vol, nil
}

func (d *driver) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	supportedModes := []csi.VolumeCapability_AccessMode_Mode{
		csi.VolumeCapability_AccessMode_SINGLE_NODE_READER_ONLY,
		csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
		csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY,
		csi.VolumeCapability_AccessMode_MULTI_NODE_SINGLE_WRITER,
	}

	msg := &strings.Builder{}
	for _, c := range req.VolumeCapabilities {
		if _, ok := c.GetAccessType().(*csi.VolumeCapability_Mount); !ok {
			fmt.Fprintf(msg, "[%s]: volume must be mount\n", req.VolumeId)
			continue
		}
		found := false
		mode := c.AccessMode.Mode
		for _, m := range supportedModes {
			if mode == m {
				found = true
			}
		}
		if !found {
			fmt.Fprintf(msg, "[%s]: unsupported AccessMode %s\n", req.VolumeId, c.AccessMode.Mode)
			continue
		}
	}
	if msg.Len() == 0 {
		caps := []*csi.VolumeCapability{}
		for _, m := range supportedModes {
			caps = append(caps, &csi.VolumeCapability{
				AccessType: &csi.VolumeCapability_Mount{},
				AccessMode: &csi.VolumeCapability_AccessMode{Mode: m},
			})
		}
		return &csi.ValidateVolumeCapabilitiesResponse{Confirmed: &csi.ValidateVolumeCapabilitiesResponse_Confirmed{VolumeCapabilities: caps}}, nil
	} else {
		return &csi.ValidateVolumeCapabilitiesResponse{Message: msg.String()}, nil
	}
}

func (d *driver) ControllerGetCapabilities(context.Context, *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	cl := []csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_LIST_VOLUMES,
		csi.ControllerServiceCapability_RPC_GET_VOLUME,
	}
	var caps []*csi.ControllerServiceCapability
	for _, c := range cl {
		caps = append(caps, &csi.ControllerServiceCapability{
			Type: &csi.ControllerServiceCapability_Rpc{
				Rpc: &csi.ControllerServiceCapability_RPC{
					Type: c,
				},
			},
		})
	}
	return &csi.ControllerGetCapabilitiesResponse{Capabilities: caps}, nil
}

func (d *driver) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	return nil, errors.New("not supported")
}

func (hp *driver) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	return nil, errors.New("not supported")
}

func (d *driver) GetCapacity(ctx context.Context, req *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	return nil, errors.New("not supported")
}

func (d *driver) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	return nil, errors.New("unsupported")
}

func (d *driver) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	return nil, errors.New("unsupported")
}

func (d *driver) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	return nil, errors.New("unsupported")
}

func (d *driver) ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	return nil, errors.New("unsupported")
}

func (d *driver) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (resp *csi.CreateVolumeResponse, finalErr error) {
	return nil, errors.New("not supported")
}

func (d *driver) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	return nil, errors.New("not supported")
}
