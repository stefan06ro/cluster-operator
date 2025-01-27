package key

import (
	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	apiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
)

func ObjRefFromCluster(cl apiv1alpha3.Cluster) corev1.ObjectReference {
	return *cl.Spec.InfrastructureRef
}

func ObjRefFromG8sControlPlane(cp infrastructurev1alpha3.G8sControlPlane) corev1.ObjectReference {
	return cp.Spec.InfrastructureRef
}

func ObjRefFromMachineDeployment(md apiv1alpha3.MachineDeployment) corev1.ObjectReference {
	return md.Spec.Template.Spec.InfrastructureRef
}

func ObjRefToNamespacedName(ref corev1.ObjectReference) types.NamespacedName {
	return types.NamespacedName{
		Name:      ref.Name,
		Namespace: ref.Namespace,
	}
}
