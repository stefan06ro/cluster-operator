package keepforinfrarefs

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v5/pkg/controller/context/finalizerskeptcontext"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/giantswarm/cluster-operator/v3/service/controller/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	or, err := r.toObjRef(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	// Due to unforeseen reasons it might happen that the infrastructure
	// reference is partially or even fully empty. In such cases we cancel the
	// resource gracefully in order to prevent errors blocking the deletion.
	{
		if or.Name == "" {
			r.logger.Debugf(ctx, "infrastructure reference misses name")
			r.logger.Debugf(ctx, "canceling resource")
			return nil
		}
		if or.Namespace == "" {
			r.logger.Debugf(ctx, "infrastructure reference misses namespace")
			r.logger.Debugf(ctx, "canceling resource")
			return nil
		}
	}

	// Here we fetch the provider specific CR defined as infrastructure reference
	// in the CAPI type. We use an unstructured object and therefore need to set
	// the api version and kind accordingly. If we would not do that the
	// controller-runtime client cannot find the right object.
	{
		r.logger.Debugf(ctx, "finding infrastructure reference")

		ir := &unstructured.Unstructured{}
		ir.SetAPIVersion(or.APIVersion)
		ir.SetKind(or.Kind)

		err = r.k8sClient.CtrlClient().Get(ctx, key.ObjRefToNamespacedName(or), ir)
		if apierrors.IsNotFound(err) {
			// At this point the runtime object linked in the infrastructure reference
			// does not exist anymore, which means the deletion of the parent can
			// continue now.
			r.logger.Debugf(ctx, "did not find infrastructure reference")
			r.logger.Debugf(ctx, "continue deletion of parent runtime object")
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "found infrastructure reference")
		r.logger.Debugf(ctx, "keeping finalizers")
		finalizerskeptcontext.SetKept(ctx)
	}

	return nil
}
