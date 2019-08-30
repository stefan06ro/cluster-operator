package v19

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/appresource"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/clusterclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/resource/k8s/configmapresource"
	"github.com/giantswarm/operatorkit/resource/k8s/secretresource"
	"github.com/giantswarm/operatorkit/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/resource/wrapper/retryresource"
	"github.com/giantswarm/tenantcluster"
	"github.com/spf13/afero"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v19/controllercontext"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v19/key"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v19/resources/app"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v19/resources/certconfig"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v19/resources/chartconfig"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v19/resources/chartoperator"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v19/resources/clusterconfigmap"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v19/resources/clusterid"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v19/resources/clusterstatus"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v19/resources/configmap"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v19/resources/cpnamespace"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v19/resources/encryptionkey"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v19/resources/kubeconfig"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v19/resources/operatorversions"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v19/resources/tcnamespace"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v19/resources/tenantclients"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v19/resources/tiller"
	"github.com/giantswarm/cluster-operator/service/controller/clusterapi/v19/resources/workercount"
)

// ClusterResourceSetConfig contains necessary dependencies and settings for
// Cluster API's Cluster controller ResourceSet configuration.
type ClusterResourceSetConfig struct {
	ApprClient    *apprclient.Client
	CertsSearcher certs.Interface
	ClusterClient *clusterclient.Client
	CMAClient     clientset.Interface
	FileSystem    afero.Fs
	G8sClient     versioned.Interface
	K8sClient     kubernetes.Interface
	Logger        micrologger.Logger
	Tenant        tenantcluster.Interface

	APIIP              string
	CalicoAddress      string
	CalicoPrefixLength string
	CertTTL            string
	ClusterIPRange     string
	DNSIP              string
	Provider           string
	RegistryDomain     string
}

// NewClusterResourceSet returns a configured Cluster API's Cluster controller
// ResourceSet.
func NewClusterResourceSet(config ClusterResourceSetConfig) (*controller.ResourceSet, error) {
	var err error

	var appGetter appresource.StateGetter
	{
		c := app.Config{
			G8sClient: config.G8sClient,
			Logger:    config.Logger,

			Provider: config.Provider,
		}

		appGetter, err = app.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var appResource controller.Resource
	{
		c := appresource.Config{
			G8sClient: config.G8sClient,
			Logger:    config.Logger,

			Name:        app.Name,
			StateGetter: appGetter,
		}

		ops, err := appresource.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		appResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var certConfigResource controller.Resource
	{
		c := certconfig.Config{
			G8sClient: config.G8sClient,
			Logger:    config.Logger,

			APIIP:    config.APIIP,
			CertTTL:  config.CertTTL,
			Provider: config.Provider,
		}

		ops, err := certconfig.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		certConfigResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var chartConfigResource controller.Resource
	{
		c := chartconfig.Config{
			Logger: config.Logger,

			Provider: config.Provider,
		}

		ops, err := chartconfig.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		chartConfigResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var chartOperatorResource controller.Resource
	{
		c := chartoperator.Config{
			ApprClient: config.ApprClient,
			FileSystem: config.FileSystem,
			Logger:     config.Logger,

			DNSIP:          config.DNSIP,
			RegistryDomain: config.RegistryDomain,
		}

		ops, err := chartoperator.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		chartOperatorResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterConfigMapGetter configmapresource.StateGetter
	{
		c := clusterconfigmap.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			DNSIP: config.DNSIP,
		}

		clusterConfigMapGetter, err = clusterconfigmap.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterConfigMapResource controller.Resource
	{
		c := configmapresource.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			Name:        clusterconfigmap.Name,
			StateGetter: clusterConfigMapGetter,
		}

		ops, err := configmapresource.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		clusterConfigMapResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterIDResource controller.Resource
	{
		c := clusterid.Config{
			CMAClient:                   config.CMAClient,
			CommonClusterStatusAccessor: &key.AWSClusterStatusAccessor{},
			G8sClient:                   config.G8sClient,
			Logger:                      config.Logger,
		}

		clusterIDResource, err = clusterid.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterStatusResource controller.Resource
	{
		c := clusterstatus.Config{
			Accessor:  &key.AWSClusterStatusAccessor{},
			CMAClient: config.CMAClient,
			G8sClient: config.G8sClient,
			Logger:    config.Logger,
		}

		clusterStatusResource, err = clusterstatus.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var configMapResource controller.Resource
	{
		c := configmap.Config{
			Logger: config.Logger,

			CalicoAddress:      config.CalicoAddress,
			CalicoPrefixLength: config.CalicoPrefixLength,
			ClusterIPRange:     config.ClusterIPRange,
			DNSIP:              config.DNSIP,
			Provider:           config.Provider,
			RegistryDomain:     config.RegistryDomain,
		}

		ops, err := configmap.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		configMapResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var cpNamespaceResource controller.Resource
	{
		c := cpnamespace.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		ops, err := cpnamespace.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		cpNamespaceResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var encryptionKeyGetter secretresource.StateGetter
	{
		c := encryptionkey.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		encryptionKeyGetter, err = encryptionkey.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var encryptionKeyResource controller.Resource
	{
		c := secretresource.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			Name:        encryptionkey.Name,
			StateGetter: encryptionKeyGetter,
		}

		ops, err := secretresource.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		encryptionKeyResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var kubeConfigGetter secretresource.StateGetter
	{
		c := kubeconfig.Config{
			CertsSearcher: config.CertsSearcher,
			K8sClient:     config.K8sClient,
			Logger:        config.Logger,
		}

		kubeConfigGetter, err = kubeconfig.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var kubeConfigResource controller.Resource
	{
		c := secretresource.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			Name:        kubeconfig.Name,
			StateGetter: kubeConfigGetter,
		}

		ops, err := secretresource.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		kubeConfigResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tcNamespaceResource controller.Resource
	{
		c := tcnamespace.Config{
			Logger: config.Logger,
		}

		ops, err := tcnamespace.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		tcNamespaceResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var operatorVersionsResource controller.Resource
	{
		c := operatorversions.Config{
			ClusterClient: config.ClusterClient,
			Logger:        config.Logger,
		}

		operatorVersionsResource, err = operatorversions.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tenantClientsResource controller.Resource
	{
		c := tenantclients.Config{
			Logger:        config.Logger,
			Tenant:        config.Tenant,
			ToClusterFunc: key.ToCluster,
		}

		tenantClientsResource, err = tenantclients.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tillerResource controller.Resource
	{
		c := tiller.Config{
			Logger: config.Logger,
		}

		tillerResource, err = tiller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var workerCountResource controller.Resource
	{
		c := workercount.Config{
			Logger: config.Logger,

			ToClusterFunc: key.ToCluster,
		}

		workerCountResource, err = workercount.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []controller.Resource{
		// Following resources manage resources controller context information.
		clusterIDResource,
		operatorVersionsResource,
		tenantClientsResource,
		workerCountResource,
		clusterStatusResource,

		// Following resources manage resources in the control plane.
		cpNamespaceResource,
		encryptionKeyResource,
		certConfigResource,
		clusterConfigMapResource,
		kubeConfigResource,
		appResource,

		// Following resources manage resources in the tenant cluster.
		tcNamespaceResource,
		tillerResource,
		chartOperatorResource,
		configMapResource,
		chartConfigResource,
	}

	// Wrap resources with retry and metrics.
	{
		c := retryresource.WrapConfig{
			Logger: config.Logger,
		}

		resources, err = retryresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	{
		c := metricsresource.WrapConfig{}
		resources, err = metricsresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	initCtxFunc := func(ctx context.Context, obj interface{}) (context.Context, error) {
		ctx = controllercontext.NewContext(ctx, controllercontext.Context{})
		return ctx, nil
	}

	handlesFunc := func(obj interface{}) bool {
		cr, err := key.ToCluster(obj)
		if err != nil {
			return false
		}

		if key.OperatorVersion(&cr) == VersionBundle().Version {
			return true
		}

		return false
	}

	var resourceSet *controller.ResourceSet
	{
		c := controller.ResourceSetConfig{
			Handles:   handlesFunc,
			InitCtx:   initCtxFunc,
			Logger:    config.Logger,
			Resources: resources,
		}

		resourceSet, err = controller.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return resourceSet, nil
}

func toCRUDResource(logger micrologger.Logger, ops controller.CRUDResourceOps) (*controller.CRUDResource, error) {
	c := controller.CRUDResourceConfig{
		Logger: logger,
		Ops:    ops,
	}

	r, err := controller.NewCRUDResource(c)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return r, nil
}