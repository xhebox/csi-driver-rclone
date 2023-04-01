/*
from csi-driver-host-path/pkg/driver

Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package driver

import (
	"errors"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/golang/glog"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"
)

func NewNonBlockingGRPCServer() *nonBlockingGRPCServer {
	return &nonBlockingGRPCServer{}
}

// NonBlocking server
type nonBlockingGRPCServer struct {
	wg      sync.WaitGroup
	server  *grpc.Server
	cleanup func()
}

func (s *nonBlockingGRPCServer) Start(endpoint string, ids csi.IdentityServer, cs csi.ControllerServer, ns csi.NodeServer) {

	s.wg.Add(1)

	go s.serve(endpoint, ids, cs, ns)

	return
}

func (s *nonBlockingGRPCServer) Wait() {
	s.wg.Wait()
}

func (s *nonBlockingGRPCServer) Stop() {
	s.server.GracefulStop()
	s.cleanup()
}

func (s *nonBlockingGRPCServer) ForceStop() {
	s.server.Stop()
	s.cleanup()
}

func (s *nonBlockingGRPCServer) serve(ep string, ids csi.IdentityServer, cs csi.ControllerServer, ns csi.NodeServer) {
	eps := strings.SplitN(ep, "://", 2)
	if eps[0] == "unix" {
		if err := os.Remove(eps[1]); err != nil && !errors.Is(err, os.ErrNotExist) {
			glog.Fatalf("Failed to remove unix socket: %v", err)
		}
		s.cleanup = func() {
			os.Remove(eps[1])
		}
	}
	listener, err := net.Listen(eps[0], eps[1])
	if err != nil {
		glog.Fatalf("Failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(logGRPC),
	}
	server := grpc.NewServer(opts...)
	s.server = server

	if ids != nil {
		csi.RegisterIdentityServer(server, ids)
	}
	if cs != nil {
		csi.RegisterControllerServer(server, cs)
	}
	if ns != nil {
		csi.RegisterNodeServer(server, ns)
	}

	glog.Infof("Listening for connections on address: %#v", listener.Addr())

	server.Serve(listener)

}

func logGRPC(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	pri := glog.Level(3)
	if info.FullMethod == "/csi.v1.Identity/Probe" {
		// This call occurs frequently, therefore it only gets log at level 5.
		pri = 5
	}
	glog.V(pri).Infof("GRPC call: %s", info.FullMethod)

	v5 := glog.V(5)
	if v5 {
		v5.Infof("GRPC request: %s", protosanitizer.StripSecrets(req))
	}
	resp, err := handler(ctx, req)
	if err != nil {
		// Always log errors. Probably not useful though without the method name?!
		glog.Errorf("GRPC error: %v", err)
	}

	if v5 {
		v5.Infof("GRPC response: %s", protosanitizer.StripSecrets(resp))
	}

	return resp, err
}
