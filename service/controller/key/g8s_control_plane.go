package key

import (
	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/microerror"
)

func G8sControlPlaneReplicas(cr infrastructurev1alpha3.G8sControlPlane) int {
	return cr.Spec.Replicas
}

func ToG8sControlPlane(v interface{}) (infrastructurev1alpha3.G8sControlPlane, error) {
	if v == nil {
		return infrastructurev1alpha3.G8sControlPlane{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &infrastructurev1alpha3.G8sControlPlane{}, v)
	}

	p, ok := v.(*infrastructurev1alpha3.G8sControlPlane)
	if !ok {
		return infrastructurev1alpha3.G8sControlPlane{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &infrastructurev1alpha3.G8sControlPlane{}, v)
	}

	c := p.DeepCopy()

	return *c, nil
}
