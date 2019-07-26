package tenantclients

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/tenantcluster"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v18/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v18/key"
)

const (
	Name = "tenantclientsv18"
)

type Config struct {
	CMAClient     clientset.Interface
	Logger        micrologger.Logger
	Tenant        tenantcluster.Interface
	ToClusterFunc func(v interface{}) (v1alpha1.Cluster, error)
}

type Resource struct {
	cmaClient     clientset.Interface
	logger        micrologger.Logger
	tenant        tenantcluster.Interface
	toClusterFunc func(v interface{}) (v1alpha1.Cluster, error)
}

func New(config Config) (*Resource, error) {
	if config.CMAClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CMAClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Tenant == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Tenant must not be empty", config)
	}
	if config.ToClusterFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ToClusterFunc must not be empty", config)
	}

	r := &Resource{
		cmaClient:     config.CMAClient,
		logger:        config.Logger,
		tenant:        config.Tenant,
		toClusterFunc: config.ToClusterFunc,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) ensure(ctx context.Context, obj interface{}) error {
	cr, err := r.toClusterFunc(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var g8sClient versioned.Interface
	var k8sClient kubernetes.Interface
	{
		g8sClient, err = r.tenant.NewG8sClient(ctx, key.ClusterID(&cr), key.ClusterAPIEndpoint(cr))
		if tenantcluster.IsTimeout(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "timeout fetching certificates")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil

		} else if err != nil {
			return microerror.Mask(err)
		}

		k8sClient, err = r.tenant.NewK8sClient(ctx, key.ClusterID(&cr), key.ClusterAPIEndpoint(cr))
		if tenantcluster.IsTimeout(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "timeout fetching certificates")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil

		} else if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		cc.Client.TenantCluster.G8s = g8sClient
		cc.Client.TenantCluster.K8s = k8sClient
	}

	return nil
}