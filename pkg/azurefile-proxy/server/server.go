/*
Copyright 2021 The Kubernetes Authors.

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

package server

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"google.golang.org/grpc"
	"k8s.io/klog/v2"
	mount_utils "k8s.io/mount-utils"
	"sigs.k8s.io/azurefile-csi-driver/pkg/azurefile"
	mount_azurefile "sigs.k8s.io/azurefile-csi-driver/pkg/azurefile-proxy/pb"
	volumehelper "sigs.k8s.io/azurefile-csi-driver/pkg/util"
)

type MountServer struct {
	mount_azurefile.UnimplementedMountServiceServer

	mounter mount_utils.Interface
}

// NewMountServer returns a new Mountserver
func NewMountServiceServer() *MountServer {
	mountServer := &MountServer{
		mounter: mount_utils.New(""),
	}
	return mountServer
}

// MountAzureFile mounts an AzureFile share to given location
func (server *MountServer) MountAzureFile(_ context.Context,
	req *mount_azurefile.MountAzureFileRequest,
) (resp *mount_azurefile.MountAzureFileResponse, err error) {

	source := req.GetSource()
	target := req.GetTarget()
	fstype := req.GetFstype()
	options := req.GetMountOptions()
	sensitiveOptions := req.GetSensitiveOptions()
	klog.V(2).Infof("received mount request: source: %s, target: %s, fstype: %s, options: %s", source, target, fstype, strings.Join(options, ","))

	execFunc := func() error {
		return server.mounter.MountSensitive(source, target, fstype, options, sensitiveOptions)
	}

	mountTimeoutInSec := azurefile.MountTimeoutInSec - 2
	timeoutFunc := func() error {
		return fmt.Errorf("mount operation timed out after %d seconds: source=%s, target=%s", mountTimeoutInSec, source, target)
	}

	if err = volumehelper.WaitUntilTimeout(time.Duration(mountTimeoutInSec)*time.Second, execFunc, timeoutFunc); err != nil {
		klog.Error("azurefile mount failed: with error:", err.Error())
		return nil, fmt.Errorf("azurefile mount failed: %v", err)
	}

	klog.V(2).Infof("azurefile successfully mounted")
	return &mount_azurefile.MountAzureFileResponse{}, nil
}

func (server *MountServer) UnmountAzureFile(_ context.Context,
	req *mount_azurefile.UnmountAzureFileRequest,
) (resp *mount_azurefile.UnmountAzureFileResponse, err error) {
	target := req.GetTarget()
	klog.V(2).Infof("Received unmount request for target: %s", target)
	err = mount_utils.CleanupMountPoint(target, server.mounter, true)
	if err != nil {
		klog.Errorf("azurefile unmount of path %s failed: with error: %v", target, err)
		return nil, fmt.Errorf("azurefile unmount of path %s failed: %v", target, err)
	}
	klog.V(2).Infof("azurefile %s successfully unmounted", target)
	return &mount_azurefile.UnmountAzureFileResponse{}, nil
}

func RunGRPCServer(
	mountServer mount_azurefile.MountServiceServer,
	enableTLS bool,
	listener net.Listener,
) error {
	serverOptions := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			grpcprom.NewServerMetrics().UnaryServerInterceptor(),
		),
	}

	grpcServer := grpc.NewServer(serverOptions...)

	mount_azurefile.RegisterMountServiceServer(grpcServer, mountServer)

	klog.V(2).Infof("Start GRPC server at %s, TLS = %t", listener.Addr().String(), enableTLS)
	return grpcServer.Serve(listener)
}
