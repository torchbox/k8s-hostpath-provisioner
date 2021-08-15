module github.com/xplodwild/k8s-hostpath-provisioner

go 1.16

require (
	github.com/golang/glog v0.0.0-20210429001901-424d2337a529
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/pkg/xattr v0.4.3
	github.com/prometheus/client_golang v1.11.0 // indirect
	github.com/prometheus/procfs v0.7.2 // indirect
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/api v0.22.0
	k8s.io/apimachinery v0.22.0
	k8s.io/client-go v11.0.0+incompatible
	sigs.k8s.io/sig-storage-lib-external-provisioner/v7 v7.0.1
)

replace k8s.io/api => k8s.io/api v0.20.1

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.20.1

replace k8s.io/apimachinery => k8s.io/apimachinery v0.21.0-alpha.0

replace k8s.io/apiserver => k8s.io/apiserver v0.20.1

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.20.1

replace k8s.io/client-go => k8s.io/client-go v0.20.1

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.20.1

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.20.1

replace k8s.io/code-generator => k8s.io/code-generator v0.20.2-rc.0

replace k8s.io/component-base => k8s.io/component-base v0.20.1

replace k8s.io/component-helpers => k8s.io/component-helpers v0.20.1

replace k8s.io/controller-manager => k8s.io/controller-manager v0.20.1

replace k8s.io/cri-api => k8s.io/cri-api v0.20.2-rc.0

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.20.1

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.20.1

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.20.1

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.20.1

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.20.1

replace k8s.io/kubectl => k8s.io/kubectl v0.20.1

replace k8s.io/kubelet => k8s.io/kubelet v0.20.1

replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.20.1

replace k8s.io/metrics => k8s.io/metrics v0.20.1

replace k8s.io/mount-utils => k8s.io/mount-utils v0.20.2-rc.0

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.20.1

replace k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.20.2-0.20201218181215-7a0b5ef74f21

replace k8s.io/sample-controller => k8s.io/sample-controller v0.20.2-0.20201218175324-ef87ddc8da22
