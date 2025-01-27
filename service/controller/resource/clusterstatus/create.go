package clusterstatus

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v5/pkg/controller/context/reconciliationcanceledcontext"
	"k8s.io/apimachinery/pkg/types"
	apiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/cluster-operator/v3/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	// Fetch the latest version of the CAPI Cluster CR first so that we can check
	// if it has its status already updated.
	var cr apiv1alpha3.Cluster
	{
		r.logger.Debugf(ctx, "finding cluster")

		cl, err := key.ToCluster(obj)
		if err != nil {
			return microerror.Mask(err)
		}

		err = r.k8sClient.CtrlClient().Get(ctx, types.NamespacedName{Name: cl.GetName(), Namespace: cl.GetNamespace()}, &cr)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "found cluster")
	}

	if cr.Status.ControlPlaneInitialized && cr.Status.InfrastructureReady {
		return nil
	}

	// Fetching the latest version of the common cluster CR, which is
	// infrastructure specific, e.g. AWSCluster CR. Once it contains the "Created"
	// status condition we want to ensure the Cluster CR status and set
	// InfrastructureReady to true.
	cc := r.newCommonClusterObjectFunc()
	{
		r.logger.Debugf(ctx, "finding infrastructure reference")

		err := r.k8sClient.CtrlClient().Get(ctx, key.ObjRefToNamespacedName(key.ObjRefFromCluster(cr)), cc)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "found infrastructure reference")
	}

	if !cc.GetCommonClusterStatus().HasCreatedCondition() {
		return nil
	}

	{
		r.logger.Debugf(ctx, "updating cluster status")

		cr.Status.ControlPlaneInitialized = true
		cr.Status.InfrastructureReady = true

		err := r.k8sClient.CtrlClient().Status().Update(ctx, &cr)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "updated cluster status")

		r.logger.Debugf(ctx, "canceling reconciliation")
		reconciliationcanceledcontext.SetCanceled(ctx)
	}

	return nil
}
