package clusterconfigmap

import (
	"reflect"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	// Name is the identifier of the resource.
	Name = "clusterconfigmapv17"
)

// Config represents the configuration used to create a new clusterConfigMap resource.
type Config struct {
	// Dependencies.
	GetClusterConfigFunc     func(obj interface{}) (v1alpha1.ClusterGuestConfig, error)
	GetClusterObjectMetaFunc func(obj interface{}) (metav1.ObjectMeta, error)
	K8sClient                kubernetes.Interface
	Logger                   micrologger.Logger

	// Settings.
	ProjectName string
}

// Resource implements the clusterConfigMap resource.
type StateGetter struct {
	// Dependencies.
	getClusterConfigFunc     func(obj interface{}) (v1alpha1.ClusterGuestConfig, error)
	getClusterObjectMetaFunc func(obj interface{}) (metav1.ObjectMeta, error)
	k8sClient                kubernetes.Interface
	logger                   micrologger.Logger

	// Settings.
	projectName string
}

// New creates a new configured clusterConfigMap resource.
func New(config Config) (*StateGetter, error) {
	// Dependencies.
	if config.GetClusterConfigFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.GetClusterConfigFunc must not be empty", config)
	}
	if config.GetClusterObjectMetaFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.GetClusterObjectMetaFunc must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	// Settings
	if config.ProjectName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ProjectName must not be empty", config)
	}

	r := &StateGetter{
		// Dependencies.
		getClusterConfigFunc:     config.GetClusterConfigFunc,
		getClusterObjectMetaFunc: config.GetClusterObjectMetaFunc,
		k8sClient:                config.K8sClient,
		logger:                   config.Logger,

		// Settings
		projectName: config.ProjectName,
	}

	return r, nil
}

func (r *StateGetter) Name() string {
	return Name
}

// equals asseses the equality of ConfigMaps with regards to distinguishing
// fields.
func equals(a, b *corev1.ConfigMap) bool {
	if a.Name != b.Name {
		return false
	}
	if a.Namespace != b.Namespace {
		return false
	}
	if !reflect.DeepEqual(a.Annotations, b.Annotations) {
		return false
	}
	if !reflect.DeepEqual(a.Data, b.Data) {
		return false
	}
	if !reflect.DeepEqual(a.Labels, b.Labels) {
		return false
	}

	return true
}

func toConfigMap(v interface{}) (*corev1.ConfigMap, error) {
	if v == nil {
		return nil, nil
	}

	configMap, ok := v.(*corev1.ConfigMap)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &corev1.ConfigMap{}, v)
	}

	return configMap, nil
}