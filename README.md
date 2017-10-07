Kubernetes hostpath provisioner
===============================

This is a Persistent Volume Claim (PVC) provisioner for Kubernetes.  It 
dynamically provisions hostPath volumes to provide storage for PVCs.  It is 
based on the
[demo hostpath-provisioner](https://github.com/kubernetes-incubator/external-storage/tree/master/docs/demo/hostpath-provisioner).

Unlike the demo provisioner, this version is intended to be suitable for 
production use.  Its purpose is to provision storage on network filesystems 
mounted on the host, rather than using Kubernetes' built-in network volume 
support.   This has some advantages over storage-specific provisioners:

* There is no need to expose PV credentials to users (cephfs-provisioner 
  requires this, for example).

* PVs can be provisioned on network storage not natively supported in 
  Kubernetes, e.g. `ceph-fuse`.

* The network storage configuration is centralised on the node (e.g., in 
  `/etc/fstab`); this means you can change the storage configuration, or even 
  completely change the storage type (e.g. NFS to CephFS) without having to 
  update every PV by hand.

There are also some disadvantages:

* Every node requires full access to the storage containing all PVs.  This may 
  defeat attempts to limit node access in Kubernetes, such as the Node 
  authorizor.

* Storage can no longer be introspected via standard Kubernetes APIs.

* Kubernetes cannot report storage-related errors such as failures to mount 
  storage; this information will not be available to users.

* Moving storage configuration from Kubernetes to the host will not work well 
  in environments where host access is limited, such as GKE.

We test and use this provisioner with CephFS and `ceph-fuse`, but in principal 
it should work with any network filesystem.  

You **cannot** use it without a network filesystem unless you can ensure all 
provisioned PVs will only be used on the host where they were provisioned; this 
is an inherent limitation of `hostPath`.

Unlike the demo hostpath-provisioner, there is no attempt to identify PVs by 
node name, because the intended use is with a network filesystem mounted on all 
hosts.

Deployment
----------

### Mount the network storage

First, mount the storage on each host.  You can do this any way you like; 
systemd mount units or `/etc/fstab` is the typical method.  The storage 
**must** be mounted at the same path on every host.

However you decide to provision the storage, you should set the mountpoint 
immutable:

```
# mkdir /mynfs
# chattr +i /mynfs
# mount -tnfs mynfsserver:/export/mynfs /mynfs
```

This ensures that nothing can be written to `/mynfs` if the storage is 
unmounted.  Without this protection, a failure to mount the storage could 
result in PVs being provisioned on the host instead of on the network storage.  
This would initially appear to work, but then lead to data loss.

Note that the immutable flag is set on the underlying mountpoint, *not* the
mounted filesystem.  Once the filesystem is mounted, the immutable mountpoint
is hidden and files can be created as normal.

### Create a StorageClass

The provisioner must be associated with a Kubernetes StorageClass:

```
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: cephfs
  annotations:
    storageclass.kubernetes.io/is-default-class: "true"
provisioner: torchbox.com/hostpath
parameters:
  pvDir: /ceph/pvs
```

* The `name` can be anything you like; you could name it after the storage type 
  (such as `cephfs-ssd` or `bignfs`), or give it a generic name like 
  `pvstorage`.

* If you *don't* want this to be the default storage class, delete the 
  `is-default-class` annotation; then this class will only be used in 
  explicitly requested in the PVC.

* Set `pvDir` to the root path on the host where volumes should be provisioned.  
  This must be on network storage, but does not need to be the root of the 
  storage or the mountpoint.

* Unless you're running multiple provisioners, leave `provisioner` at the 
  default `torchbox.com/hostpath`.  If you want to run multiple provisioners, 
  the value passed to `-name` when starting the provisioner must match the 
  value of `provisioner`.

### Start the provisioner

#### Out-of-cluster

For out-of-cluster use, just provide a kubeconfig file when starting:

```
# ./hostpath-provisioner -kubeconfig ~/.kube/config
```

In most cases the provisioner will need to run as root, but you can run it as a 
non-privileged user as long as it has access to create and delete PVs.  (For 
example, by giving it the `CAP_DAC_OVERRIDE` capability; using an ACL is not 
sufficient because a user could remove the ACL and prevent PVs from being 
deleted.)

### In-cluster

Use the sample `deployment.yaml` to deploy the provisioner.  You must edit the
`volumes` volume to point to the location of the network storage; in nearly all
cases, this should be the same as the `pvDir` in the storage class.

The sample deployment includes RBAC configuration, and creates a new service
account called `hostpath-provisioner`.  If you have restricted access to
`hostPath` volumes using Pod Security Policies, you must ensure this
serviceaccount can use hostPath.  (This is not required for consumers of the
created PVs, only the provisioner itself.)

### Create a PVC

Test the provisioner by creating a new PVC:

```
$ cat testpvc.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: testpvc
spec:
  accessModes:
  - ReadWriteMany
  resources:
    requests:
      storage: 50Gi
$ kubectl create -f testpvc.yaml 
persistentvolumeclaim "testpvc" created
```

The provisioner should detect the new PVC and provision a PV.  The PVC will then become Bound:

```
$ kubectl get pvc               
NAME      STATUS    VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
testpvc   Bound     pvc-145c785e-ab83-11e7-9432-4201ac1fd019   50Gi       RWX            cephfs         10s
```

You can now use the PVC in a pod.
