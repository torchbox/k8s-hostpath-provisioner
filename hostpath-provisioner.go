/* vim:set sw=8 ts=8 noet:
 *
 * Copyright 2016 The Kubernetes Authors.
 * Copyright 2017 Torchbox Ltd.
 * Copyright 2021 Vladimir Yumatov
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"time"

	"syscall"

	"github.com/golang/glog"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/sig-storage-lib-external-provisioner/v7/controller"
)

/* Our constants */
const (
	resyncPeriod     = 15 * time.Second
	provisionerName  = "roruk/hostpath"
	provisionerIDAnn = "roruk/hostpath-provisioner-id"
)

/* Our provisioner class, which implements the controller API. */
type hostPathProvisioner struct {
	//client   kubernetes.Interface /* Kubernetes client for accessing the cluster during provision */
	identity string /* Our unique provisioner identity */
}

/* Storage the parsed configuration from the storage class */
type hostPathParameters struct {
	pvDir string /* On-disk path of the PV root */
}

/*
 * Create a new provisioner from a given client and identity.
 */
func NewHostPathProvisioner(id string) controller.Provisioner {
	return &hostPathProvisioner{
		identity: id,
	}
}

var _ controller.Provisioner = &hostPathProvisioner{}

/*
 * Provision: create the physical on-disk path for this PV and return a new
 * volume referencing it as a hostPath.  The volume is annotated with our
 * provisioner id, so multiple provisioners can run on the same cluster.
 */
func (p *hostPathProvisioner) Provision(ctx context.Context, options controller.ProvisionOptions) (*v1.PersistentVolume, controller.ProvisioningState, error) {
	/*
	 * Extract the PV capacity as bytes.  We can use this to set CephFS
	 * quotas.
	 */
	capacity := options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]
	volbytes := capacity.Value()
	if volbytes <= 0 {
		return nil, controller.ProvisioningFinished, fmt.Errorf("storage capacity must be >= 0 (not %+v)", capacity.String())
	}

	/*
	 * Fetch the PV root directory from the PV storage class.
	 */
	/* Create the on-disk directory. */
	volumePath := path.Join(options.StorageClass.Parameters["pvDir"], options.PVName)
	if err := os.MkdirAll(volumePath, 0777); err != nil {
		glog.Errorf("failed to mkdir %s: %s", volumePath, err)
		return nil, controller.ProvisioningFinished, err
	}
	if err := os.Chmod(volumePath, 0777); err != nil {
		glog.Errorf("failed to chmod %s, %s", volumePath, err)
		return nil, controller.ProvisioningFinished, err
	}
	glog.Infof("successfully chmoded %s", volumePath)

	/* The actual PV we will create */
	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: options.PVName,
			Annotations: map[string]string{
				provisionerIDAnn: p.identity,
			},
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: *options.StorageClass.ReclaimPolicy,
			AccessModes:                   options.PVC.Spec.AccessModes,
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)],
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Path: volumePath,
				},
			},
		},
	}

	glog.Infof("successfully created hostpath volume %s (%s)",
		options.PVName, volumePath)

	return pv, controller.ProvisioningFinished, nil
}

/*
 * Delete: remove a PV from the disk by deleting its directory.
 */
func (p *hostPathProvisioner) Delete(ctx context.Context, volume *v1.PersistentVolume) error {
	/* Ensure this volume was provisioned by us */
	ann, ok := volume.Annotations[provisionerIDAnn]
	if !ok {
		glog.Infof("not removing volume <%s>: identity annotation <%s> missing",
			volume.Name, provisionerIDAnn)
		return errors.New("identity annotation not found on PV")
	}
	glog.Infof("Remove volume %s", volume.Name)
	if ann != p.identity {
		glog.Infof("not removing volume <%s>: identity annotation <%s> does not match ours <%s>",
			volume.Name, p.identity, provisionerIDAnn)
		return &controller.IgnoredError{Reason: "identity annotation on PV does not match ours"}
	}

	volumePath := volume.Spec.HostPath.Path
	if err := os.RemoveAll(volumePath); err != nil {
		glog.Errorf("failed to remove PV %s (%s): %v",
			volume.Name, volumePath, err)
		return err
	}

	return nil
}

var (
	master     = flag.String("master", "", "Master URL")
	kubeconfig = flag.String("kubeconfig", "", "Absolute path to the kubeconfig")
	name       = flag.String("name", "", "Provisioner name")
	id         = flag.String("id", "", "Unique provisioner identity")
)

func main() {
	syscall.Umask(0)

	flag.Parse()
	flag.Set("logtostderr", "true")

	/* Configure the client based on our command line. */
	var config *rest.Config
	var err error
	if *master != "" || *kubeconfig != "" {
		glog.Infof("using out-of-cluster configuration")
		config, err = clientcmd.BuildConfigFromFlags(*master, *kubeconfig)
	} else {
		glog.Infof("using in-cluster configuration; use -master or -kubeconfig to change")
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		glog.Fatalf("failed to create config: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Fatalf("failed to create client: %v", err)
	}

	/*
	 * Default provisioner id to the name; the user can override with the
	 * -id option.
	 */
	prID := provisionerName
	if *id != "" {
		prID = *id
	}

	prName := provisionerName
	if *name != "" {
		prName = *name
	}

	/*
	 * Create the provisioner, which has a standard interface (Provision,
	 * Delete) used by the controller to notify us what to do.
	 */
	hostPathProvisioner := NewHostPathProvisioner(prID)

	/* Start the controller */
	pc := controller.NewProvisionController(
		clientset,
		prName,
		hostPathProvisioner)

	pc.Run(context.Background())
}
