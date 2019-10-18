package app

import (
	"context"
	"fmt"
	"strconv"

	g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/annotation"
	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/project"
	"github.com/giantswarm/cluster-operator/pkg/v21/key"
	awskey "github.com/giantswarm/cluster-operator/service/controller/aws/v21/key"
	azurekey "github.com/giantswarm/cluster-operator/service/controller/azure/v21/key"
	kvmkey "github.com/giantswarm/cluster-operator/service/controller/kvm/v21/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) ([]*g8sv1alpha1.App, error) {
	clusterConfig, err := r.getClusterConfigFunc(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	configMaps, err := r.getConfigMaps(ctx, clusterConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	secrets, err := r.getSecrets(ctx, clusterConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var apps []*g8sv1alpha1.App

	for _, appSpec := range r.newAppSpecs() {
		userConfig := newUserConfig(clusterConfig, appSpec, configMaps, secrets)
		apps = append(apps, r.newApp(clusterConfig, appSpec, userConfig))
	}

	return apps, nil
}

func (r *Resource) getConfigMaps(ctx context.Context, clusterConfig v1alpha1.ClusterGuestConfig) (map[string]corev1.ConfigMap, error) {
	configMaps := map[string]corev1.ConfigMap{}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding configMaps in namespace %#q", clusterConfig.ID))

	list, err := r.k8sClient.CoreV1().ConfigMaps(clusterConfig.ID).List(metav1.ListOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	for _, cm := range list.Items {
		configMaps[cm.Name] = cm
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d configMaps in namespace %#q", len(configMaps), clusterConfig.ID))

	return configMaps, nil
}

func (r *Resource) getSecrets(ctx context.Context, clusterConfig v1alpha1.ClusterGuestConfig) (map[string]corev1.Secret, error) {
	secrets := map[string]corev1.Secret{}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding secrets in namespace %#q", clusterConfig.ID))

	list, err := r.k8sClient.CoreV1().Secrets(clusterConfig.ID).List(metav1.ListOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	for _, s := range list.Items {
		secrets[s.Name] = s
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d secrets in namespace %#q", len(secrets), clusterConfig.ID))

	return secrets, nil
}

func (r *Resource) newApp(clusterConfig v1alpha1.ClusterGuestConfig, appSpec key.AppSpec, userConfig g8sv1alpha1.AppSpecUserConfig) *g8sv1alpha1.App {
	return &g8sv1alpha1.App{
		TypeMeta: metav1.TypeMeta{
			Kind:       "App",
			APIVersion: "application.giantswarm.io",
		},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				annotation.ForceHelmUpgrade: strconv.FormatBool(appSpec.UseUpgradeForce),
			},
			Labels: map[string]string{
				label.App:                appSpec.App,
				label.AppOperatorVersion: "1.0.0",
				label.Cluster:            clusterConfig.ID,
				label.ManagedBy:          project.Name(),
				label.Organization:       clusterConfig.Owner,
				label.ServiceType:        label.ServiceTypeManaged,
			},
			Name:      appSpec.App,
			Namespace: clusterConfig.ID,
		},
		Spec: g8sv1alpha1.AppSpec{
			Catalog:   appSpec.Catalog,
			Name:      appSpec.Chart,
			Namespace: appSpec.Namespace,
			Version:   appSpec.Version,

			Config: g8sv1alpha1.AppSpecConfig{
				ConfigMap: g8sv1alpha1.AppSpecConfigConfigMap{
					Name:      key.ClusterConfigMapName(clusterConfig),
					Namespace: clusterConfig.ID,
				},
			},

			KubeConfig: g8sv1alpha1.AppSpecKubeConfig{
				Context: g8sv1alpha1.AppSpecKubeConfigContext{
					Name: key.KubeConfigSecretName(clusterConfig),
				},
				InCluster: false,
				Secret: g8sv1alpha1.AppSpecKubeConfigSecret{
					Name:      key.KubeConfigSecretName(clusterConfig),
					Namespace: clusterConfig.ID,
				},
			},

			UserConfig: userConfig,
		},
	}
}

func (r *Resource) newAppSpecs() []key.AppSpec {
	switch r.provider {
	case "aws":
		return append(key.CommonAppSpecs(), awskey.AppSpecs()...)
	case "azure":
		return append(key.CommonAppSpecs(), azurekey.AppSpecs()...)
	case "kvm":
		return append(key.CommonAppSpecs(), kvmkey.AppSpecs()...)
	default:
		return key.CommonAppSpecs()
	}
}

func newUserConfig(clusterConfig v1alpha1.ClusterGuestConfig, appSpec key.AppSpec, configMaps map[string]corev1.ConfigMap, secrets map[string]corev1.Secret) g8sv1alpha1.AppSpecUserConfig {
	userConfig := g8sv1alpha1.AppSpecUserConfig{}

	_, ok := configMaps[key.AppUserConfigMapName(appSpec)]
	if ok {
		configMapSpec := g8sv1alpha1.AppSpecUserConfigConfigMap{
			Name:      key.AppUserConfigMapName(appSpec),
			Namespace: clusterConfig.ID,
		}

		userConfig.ConfigMap = configMapSpec
	}

	_, ok = secrets[key.AppUserSecretName(appSpec)]
	if ok {
		secretSpec := g8sv1alpha1.AppSpecUserConfigSecret{
			Name:      key.AppUserSecretName(appSpec),
			Namespace: clusterConfig.ID,
		}

		userConfig.Secret = secretSpec
	}

	return userConfig
}