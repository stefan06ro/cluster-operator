package chartconfig

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/cluster"
	"github.com/giantswarm/cluster-operator/pkg/label"
)

const (
	chartConfigAPIVersion           = "core.giantswarm.io"
	chartConfigKind                 = "ChartConfig"
	chartConfigVersionBundleVersion = "0.1.0"
)

// GetDesiredState returns all desired ChartConfigs for managed guest resources.
func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	clusterGuestConfig, err := r.toClusterGuestConfigFunc(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	clusterConfig, err := prepareClusterConfig(r.baseClusterConfig, clusterGuestConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	desiredChartConfigs := make([]*v1alpha1.ChartConfig, 0)
	{
		chartConfig := newKubeStateMetricsChartConfig(clusterConfig, r.projectName)
		desiredChartConfigs = append(desiredChartConfigs, chartConfig)
	}
	{
		chartConfig := newNodeExporterChartConfig(clusterConfig, r.projectName)
		desiredChartConfigs = append(desiredChartConfigs, chartConfig)
	}

	// Only enable Ingress Controller for Azure.
	if r.provider == label.ProviderAzure {
		chartConfig := newIngressControllerChartConfig(clusterConfig, r.projectName)
		desiredChartConfigs = append(desiredChartConfigs, chartConfig)
	}

	// Only enable External DNS for Azure.
	if r.provider == label.ProviderAzure {
		chartConfig := newExternalDNSChartConfig(clusterConfig, r.projectName)
		desiredChartConfigs = append(desiredChartConfigs, chartConfig)
	}

	return desiredChartConfigs, nil
}

func newIngressControllerChartConfig(clusterConfig cluster.Config, projectName string) *v1alpha1.ChartConfig {
	chartName := "kubernetes-nginx-ingress-controller-chart"
	channelName := "0-1-stable"
	releaseName := "nginx-ingress-controller"
	labels := newChartConfigLabels(clusterConfig, releaseName, projectName)

	return &v1alpha1.ChartConfig{
		TypeMeta: apismetav1.TypeMeta{
			Kind:       chartConfigKind,
			APIVersion: chartConfigAPIVersion,
		},
		ObjectMeta: apismetav1.ObjectMeta{
			Name:   chartName,
			Labels: labels,
		},
		Spec: v1alpha1.ChartConfigSpec{
			Chart: v1alpha1.ChartConfigSpecChart{
				Name:      chartName,
				Channel:   channelName,
				Namespace: apismetav1.NamespaceSystem,
				Release:   releaseName,
			},
			VersionBundle: v1alpha1.ChartConfigSpecVersionBundle{
				Version: chartConfigVersionBundleVersion,
			},
		},
	}
}

func newExternalDNSChartConfig(clusterConfig cluster.Config, projectName string) *v1alpha1.ChartConfig {
	chartName := "kubernetes-external-dns-chart"
	channelName := "0-1-stable"
	releaseName := "external-dns"
	labels := newChartConfigLabels(clusterConfig, releaseName, projectName)

	return &v1alpha1.ChartConfig{
		TypeMeta: apismetav1.TypeMeta{
			Kind:       chartConfigKind,
			APIVersion: chartConfigAPIVersion,
		},
		ObjectMeta: apismetav1.ObjectMeta{
			Name:   chartName,
			Labels: labels,
		},
		Spec: v1alpha1.ChartConfigSpec{
			Chart: v1alpha1.ChartConfigSpecChart{
				Name:      chartName,
				Channel:   channelName,
				Namespace: apismetav1.NamespaceSystem,
				Release:   releaseName,
			},
			VersionBundle: v1alpha1.ChartConfigSpecVersionBundle{
				Version: chartConfigVersionBundleVersion,
			},
		},
	}
}

func newKubeStateMetricsChartConfig(clusterConfig cluster.Config, projectName string) *v1alpha1.ChartConfig {
	chartName := "kubernetes-kube-state-metrics-chart"
	channelName := "0-1-stable"
	releaseName := "kube-state-metrics"
	labels := newChartConfigLabels(clusterConfig, releaseName, projectName)

	return &v1alpha1.ChartConfig{
		TypeMeta: apismetav1.TypeMeta{
			Kind:       chartConfigKind,
			APIVersion: chartConfigAPIVersion,
		},
		ObjectMeta: apismetav1.ObjectMeta{
			Name:   chartName,
			Labels: labels,
		},
		Spec: v1alpha1.ChartConfigSpec{
			Chart: v1alpha1.ChartConfigSpecChart{
				Name:      chartName,
				Channel:   channelName,
				Namespace: apismetav1.NamespaceSystem,
				Release:   releaseName,
			},
			VersionBundle: v1alpha1.ChartConfigSpecVersionBundle{
				Version: chartConfigVersionBundleVersion,
			},
		},
	}
}

func newNodeExporterChartConfig(clusterConfig cluster.Config, projectName string) *v1alpha1.ChartConfig {
	chartName := "kubernetes-node-exporter-chart"
	channelName := "0-1-stable"
	releaseName := "node-exporter"
	labels := newChartConfigLabels(clusterConfig, releaseName, projectName)

	return &v1alpha1.ChartConfig{
		TypeMeta: apismetav1.TypeMeta{
			Kind:       chartConfigKind,
			APIVersion: chartConfigAPIVersion,
		},
		ObjectMeta: apismetav1.ObjectMeta{
			Name:   chartName,
			Labels: labels,
		},
		Spec: v1alpha1.ChartConfigSpec{
			Chart: v1alpha1.ChartConfigSpecChart{
				Name:      chartName,
				Channel:   channelName,
				Namespace: apismetav1.NamespaceSystem,
				Release:   releaseName,
			},
			VersionBundle: v1alpha1.ChartConfigSpecVersionBundle{
				Version: chartConfigVersionBundleVersion,
			},
		},
	}
}

func newChartConfigLabels(clusterConfig cluster.Config, appName, projectName string) map[string]string {
	return map[string]string{
		label.App:          appName,
		label.Cluster:      clusterConfig.ClusterID,
		label.ManagedBy:    projectName,
		label.Organization: clusterConfig.Organization,
		label.ServiceType:  label.ServiceTypeManaged,
	}
}