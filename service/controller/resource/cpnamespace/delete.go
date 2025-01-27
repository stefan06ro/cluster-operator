package cpnamespace

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v5/pkg/resource/crud"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/v3/service/controller/key"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	ns, err := toNamespace(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if ns != nil {
		r.logger.Debugf(ctx, "deleting namespace %#q in control plane", ns.Name)

		err = r.k8sClient.CoreV1().Namespaces().Delete(ctx, ns.Name, metav1.DeleteOptions{})
		if apierrors.IsNotFound(err) {
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "deleted namespace %#q in control plane", ns.Name)
	} else {
		r.logger.Debugf(ctx, "did not delete namespace %#q in control plane", key.ClusterID(&cr))
	}

	return nil
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*crud.Patch, error) {
	delete, err := r.newDeleteChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := crud.NewPatch()
	patch.SetDeleteChange(delete)

	return patch, nil
}

func (r *Resource) newDeleteChange(ctx context.Context, obj, currentState, desiredState interface{}) (*corev1.Namespace, error) {
	currentNamespace, err := toNamespace(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return currentNamespace, nil
}
