package machinedeploymentstatus

import (
	"context"
	"fmt"

	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/v5/pkg/controller/context/finalizerskeptcontext"
	"github.com/giantswarm/operatorkit/v5/pkg/controller/context/reconciliationcanceledcontext"
	"github.com/giantswarm/operatorkit/v5/pkg/controller/context/resourcecanceledcontext"
	"k8s.io/apimachinery/pkg/types"
	apiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/cluster-operator/v3/pkg/label"
	"github.com/giantswarm/cluster-operator/v3/service/controller/key"
	"github.com/giantswarm/cluster-operator/v3/service/internal/basedomain"
	"github.com/giantswarm/cluster-operator/v3/service/internal/nodecount"
	"github.com/giantswarm/cluster-operator/v3/service/internal/recorder"
	"github.com/giantswarm/cluster-operator/v3/service/internal/tenantclient"
)

const (
	Name = "machinedeploymentstatus"
)

type Config struct {
	Event     recorder.Interface
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger
	NodeCount nodecount.Interface
}

type Resource struct {
	event     recorder.Interface
	k8sClient k8sclient.Interface
	logger    micrologger.Logger
	nodeCount nodecount.Interface
}

func New(config Config) (*Resource, error) {
	if config.Event == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Event must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.NodeCount == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.NodeCount must not be empty", config)
	}

	r := &Resource{
		event:     config.Event,
		k8sClient: config.K8sClient,
		logger:    config.Logger,
		nodeCount: config.NodeCount,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) ensure(ctx context.Context, obj interface{}) error {
	cr := &apiv1alpha3.MachineDeployment{}
	{
		md, err := key.ToMachineDeployment(obj)
		if err != nil {
			return microerror.Mask(err)
		}

		err = r.k8sClient.CtrlClient().Get(ctx, types.NamespacedName{Name: md.Name, Namespace: md.Namespace}, cr)
		if err != nil {
			return microerror.Mask(err)
		}
	}
	workerCount, err := r.nodeCount.WorkerCount(ctx, cr)
	if tenantclient.IsNotAvailable(err) {
		r.logger.LogCtx(
			ctx,
			"level", "debug",
			"message", fmt.Sprintf("not getting worker nodes for tenant cluster %#q", key.ClusterID(cr)),
			"reason", "tenant cluster api not available yet",
		)
		r.logger.Debugf(ctx, "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil
	} else if basedomain.IsNotFound(err) {
		// in case of a cluster deletion AWSCluster CR does not exist anymore, handle basedomain error gracefully
		r.logger.Debugf(ctx, "not getting basedomain for tenant client %#q", key.ClusterID(cr))
		r.logger.Debugf(ctx, "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}
	{
		r.logger.Debugf(ctx, "checking if status of machine deployment needs to be updated")

		replicasChanged := cr.Status.Replicas != workerCount[cr.Labels[label.MachineDeployment]].Nodes
		readyReplicasChanged := cr.Status.ReadyReplicas != workerCount[cr.Labels[label.MachineDeployment]].Ready

		if !replicasChanged && !readyReplicasChanged {
			r.logger.Debugf(ctx, "status of machine deployment does not need to be updated")
			return nil
		}

		r.logger.Debugf(ctx, "status of machine deployment needs to be updated")
	}

	{
		cr.Status.Replicas = workerCount[cr.Labels[label.MachineDeployment]].Nodes
		cr.Status.ReadyReplicas = workerCount[cr.Labels[label.MachineDeployment]].Ready
	}

	{
		r.logger.Debugf(ctx, "updating status of machine deployment")

		err := r.k8sClient.CtrlClient().Status().Update(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "updated status of machine deployment")
		r.event.Emit(ctx, cr, "MachineDeploymentUpdated",
			fmt.Sprintf("updated status of machine deployment, changed replicas %d -> %d", cr.Status.Replicas, cr.Status.ReadyReplicas),
		)

		if key.IsDeleted(cr) {
			r.logger.Debugf(ctx, "keeping finalizers")
			finalizerskeptcontext.SetKept(ctx)
		}

		r.logger.Debugf(ctx, "canceling reconciliation")
		reconciliationcanceledcontext.SetCanceled(ctx)
	}

	return nil
}
