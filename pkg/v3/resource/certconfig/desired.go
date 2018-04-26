package certconfig

import (
	"context"
	"strings"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/versionbundle"
	apimetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-operator/pkg/cluster"
	"github.com/giantswarm/cluster-operator/pkg/label"
	"github.com/giantswarm/cluster-operator/pkg/v3/key"
)

const (
	certAPIVersion = "core.giantswarm.io"
	certKind       = "CertConfig"
	certOperatorID = "cert-operator"

	loopbackIP = "127.0.0.1"

	// systemMastersOrganization is the RBAC kubernetes admin group.
	systemMastersOrganization = "system:masters"
)

var (
	kubeAltNames = []string{
		"kubernetes",
		"kubernetes.default",
		"kubernetes.default.svc",
		"kubernetes.default.svc.cluster.local",
	}
)

// GetDesiredState returns all desired CertConfigs for managed certificates.
func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	clusterGuestConfig, err := r.toClusterGuestConfigFunc(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	clusterConfig, err := prepareClusterConfig(r.baseClusterConfig, clusterGuestConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	desiredCertConfigs := make([]*v1alpha1.CertConfig, 0)
	{
		certConfig := newAPICertConfig(clusterConfig, certs.APICert, r.projectName)
		desiredCertConfigs = append(desiredCertConfigs, certConfig)
	}
	{
		certConfig := newCalicoCertConfig(clusterConfig, certs.CalicoCert, r.projectName)
		desiredCertConfigs = append(desiredCertConfigs, certConfig)
	}
	{
		certConfig := newClusterOperatorAPICertConfig(clusterConfig, certs.ClusterOperatorAPICert, r.projectName)
		desiredCertConfigs = append(desiredCertConfigs, certConfig)
	}
	{
		certConfig := newEtcdCertConfig(clusterConfig, certs.EtcdCert, r.projectName)
		desiredCertConfigs = append(desiredCertConfigs, certConfig)
	}
	{
		certConfig := newNodeOperatorCertConfig(clusterConfig, certs.NodeOperatorCert, r.projectName)
		desiredCertConfigs = append(desiredCertConfigs, certConfig)
	}
	{
		certConfig := newPrometheusCertConfig(clusterConfig, certs.PrometheusCert, r.projectName)
		desiredCertConfigs = append(desiredCertConfigs, certConfig)
	}
	{
		certConfig := newServiceAccountCertConfig(clusterConfig, certs.ServiceAccountCert, r.projectName)
		desiredCertConfigs = append(desiredCertConfigs, certConfig)
	}
	{
		certConfig := newWorkerCertConfig(clusterConfig, certs.WorkerCert, r.projectName)
		desiredCertConfigs = append(desiredCertConfigs, certConfig)
	}

	return desiredCertConfigs, nil
}

func prepareClusterConfig(baseClusterConfig cluster.Config, clusterGuestConfig v1alpha1.ClusterGuestConfig) (cluster.Config, error) {
	var err error

	// Copy baseClusterConfig as basis and supplement it with information from
	// clusterGuestConfig.
	clusterConfig := baseClusterConfig

	clusterConfig.ClusterID = key.ClusterID(clusterGuestConfig)

	clusterConfig.Domain.API, err = newServerDomain(key.DNSZone(clusterGuestConfig), certs.APICert)
	if err != nil {
		return cluster.Config{}, microerror.Mask(err)
	}
	clusterConfig.Domain.Calico, err = newServerDomain(key.DNSZone(clusterGuestConfig), certs.CalicoCert)
	if err != nil {
		return cluster.Config{}, microerror.Mask(err)
	}
	clusterConfig.Domain.Etcd, err = newServerDomain(key.DNSZone(clusterGuestConfig), certs.EtcdCert)
	if err != nil {
		return cluster.Config{}, microerror.Mask(err)
	}
	clusterConfig.Domain.NodeOperator, err = newServerDomain(key.DNSZone(clusterGuestConfig), certs.NodeOperatorCert)
	if err != nil {
		return cluster.Config{}, microerror.Mask(err)
	}
	clusterConfig.Domain.Prometheus, err = newServerDomain(key.DNSZone(clusterGuestConfig), certs.PrometheusCert)
	if err != nil {
		return cluster.Config{}, microerror.Mask(err)
	}
	clusterConfig.Domain.ServiceAccount, err = newServerDomain(key.DNSZone(clusterGuestConfig), certs.ServiceAccountCert)
	if err != nil {
		return cluster.Config{}, microerror.Mask(err)
	}
	clusterConfig.Domain.Worker, err = newServerDomain(key.DNSZone(clusterGuestConfig), certs.WorkerCert)
	if err != nil {
		return cluster.Config{}, microerror.Mask(err)
	}

	clusterConfig.Organization = clusterGuestConfig.Owner

	versionBundle, err := versionbundle.GetBundleByName(key.VersionBundles(clusterGuestConfig), certOperatorID)
	if err != nil {
		return cluster.Config{}, microerror.Mask(err)
	}

	clusterConfig.VersionBundleVersion = versionBundle.Version

	return clusterConfig, nil
}

func newAPICertConfig(clusterConfig cluster.Config, cert certs.Cert, projectName string) *v1alpha1.CertConfig {
	certName := string(cert)
	return &v1alpha1.CertConfig{
		TypeMeta: apimetav1.TypeMeta{
			Kind:       certKind,
			APIVersion: certAPIVersion,
		},
		ObjectMeta: apimetav1.ObjectMeta{
			Name: key.CertConfigName(clusterConfig.ClusterID, cert),
			Labels: map[string]string{
				label.Cluster:         clusterConfig.ClusterID,
				label.LegacyClusterID: clusterConfig.ClusterID,
				label.LegacyComponent: certName,
				label.ManagedBy:       projectName,
				label.Organization:    clusterConfig.Organization,
			},
		},
		Spec: v1alpha1.CertConfigSpec{
			Cert: v1alpha1.CertConfigSpecCert{
				AllowBareDomains:    true,
				AltNames:            kubeAltNames,
				ClusterComponent:    certName,
				ClusterID:           clusterConfig.ClusterID,
				CommonName:          clusterConfig.Domain.API,
				DisableRegeneration: false,
				IPSANs:              []string{clusterConfig.IP.API.String()},
				Organizations:       []string{systemMastersOrganization},
				TTL:                 clusterConfig.CertTTL,
			},
			VersionBundle: v1alpha1.CertConfigSpecVersionBundle{
				Version: clusterConfig.VersionBundleVersion,
			},
		},
	}
}

func newCalicoCertConfig(clusterConfig cluster.Config, cert certs.Cert, projectName string) *v1alpha1.CertConfig {
	certName := string(cert)
	return &v1alpha1.CertConfig{
		TypeMeta: apimetav1.TypeMeta{
			Kind:       certKind,
			APIVersion: certAPIVersion,
		},
		ObjectMeta: apimetav1.ObjectMeta{
			Name: key.CertConfigName(clusterConfig.ClusterID, cert),
			Labels: map[string]string{
				label.Cluster:         clusterConfig.ClusterID,
				label.LegacyClusterID: clusterConfig.ClusterID,
				label.LegacyComponent: certName,
				label.ManagedBy:       projectName,
				label.Organization:    clusterConfig.Organization,
			},
		},
		Spec: v1alpha1.CertConfigSpec{
			Cert: v1alpha1.CertConfigSpecCert{
				AllowBareDomains:    false,
				ClusterComponent:    certName,
				ClusterID:           clusterConfig.ClusterID,
				CommonName:          clusterConfig.Domain.Calico,
				DisableRegeneration: false,
				TTL:                 clusterConfig.CertTTL,
			},
			VersionBundle: v1alpha1.CertConfigSpecVersionBundle{
				Version: clusterConfig.VersionBundleVersion,
			},
		},
	}
}

func newClusterOperatorAPICertConfig(clusterConfig cluster.Config, cert certs.Cert, projectName string) *v1alpha1.CertConfig {
	certName := string(cert)
	return &v1alpha1.CertConfig{
		TypeMeta: apimetav1.TypeMeta{
			Kind:       certKind,
			APIVersion: certAPIVersion,
		},
		ObjectMeta: apimetav1.ObjectMeta{
			Name: key.CertConfigName(clusterConfig.ClusterID, cert),
			Labels: map[string]string{
				label.Cluster:      clusterConfig.ClusterID,
				label.ManagedBy:    projectName,
				label.Organization: clusterConfig.Organization,
			},
		},
		Spec: v1alpha1.CertConfigSpec{
			Cert: v1alpha1.CertConfigSpecCert{
				AllowBareDomains:    false,
				ClusterComponent:    certName,
				ClusterID:           clusterConfig.ClusterID,
				CommonName:          clusterConfig.Domain.API,
				DisableRegeneration: false,
				TTL:                 clusterConfig.CertTTL,
			},
			VersionBundle: v1alpha1.CertConfigSpecVersionBundle{
				Version: clusterConfig.VersionBundleVersion,
			},
		},
	}
}

func newEtcdCertConfig(clusterConfig cluster.Config, cert certs.Cert, projectName string) *v1alpha1.CertConfig {
	certName := string(cert)
	return &v1alpha1.CertConfig{
		TypeMeta: apimetav1.TypeMeta{
			Kind:       certKind,
			APIVersion: certAPIVersion,
		},
		ObjectMeta: apimetav1.ObjectMeta{
			Name: key.CertConfigName(clusterConfig.ClusterID, cert),
			Labels: map[string]string{
				label.Cluster:         clusterConfig.ClusterID,
				label.LegacyClusterID: clusterConfig.ClusterID,
				label.LegacyComponent: certName,
				label.ManagedBy:       projectName,
				label.Organization:    clusterConfig.Organization,
			},
		},
		Spec: v1alpha1.CertConfigSpec{
			Cert: v1alpha1.CertConfigSpecCert{
				AllowBareDomains:    true,
				ClusterComponent:    certName,
				ClusterID:           clusterConfig.ClusterID,
				CommonName:          clusterConfig.Domain.Etcd,
				DisableRegeneration: false,
				IPSANs:              []string{loopbackIP},
				TTL:                 clusterConfig.CertTTL,
			},
			VersionBundle: v1alpha1.CertConfigSpecVersionBundle{
				Version: clusterConfig.VersionBundleVersion,
			},
		},
	}
}

func newNodeOperatorCertConfig(clusterConfig cluster.Config, cert certs.Cert, projectName string) *v1alpha1.CertConfig {
	certName := string(cert)
	return &v1alpha1.CertConfig{
		TypeMeta: apimetav1.TypeMeta{
			Kind:       certKind,
			APIVersion: certAPIVersion,
		},
		ObjectMeta: apimetav1.ObjectMeta{
			Name: key.CertConfigName(clusterConfig.ClusterID, cert),
			Labels: map[string]string{
				label.Cluster:         clusterConfig.ClusterID,
				label.LegacyClusterID: clusterConfig.ClusterID,
				label.LegacyComponent: certName,
				label.ManagedBy:       projectName,
				label.Organization:    clusterConfig.Organization,
			},
		},
		Spec: v1alpha1.CertConfigSpec{
			Cert: v1alpha1.CertConfigSpecCert{
				AllowBareDomains:    false,
				ClusterComponent:    certName,
				ClusterID:           clusterConfig.ClusterID,
				CommonName:          clusterConfig.Domain.NodeOperator,
				DisableRegeneration: false,
				TTL:                 clusterConfig.CertTTL,
			},
			VersionBundle: v1alpha1.CertConfigSpecVersionBundle{
				Version: clusterConfig.VersionBundleVersion,
			},
		},
	}
}

func newPrometheusCertConfig(clusterConfig cluster.Config, cert certs.Cert, projectName string) *v1alpha1.CertConfig {
	certName := string(cert)
	return &v1alpha1.CertConfig{
		TypeMeta: apimetav1.TypeMeta{
			Kind:       certKind,
			APIVersion: certAPIVersion,
		},
		ObjectMeta: apimetav1.ObjectMeta{
			Name: key.CertConfigName(clusterConfig.ClusterID, cert),
			Labels: map[string]string{
				label.Cluster:         clusterConfig.ClusterID,
				label.LegacyClusterID: clusterConfig.ClusterID,
				label.LegacyComponent: certName,
				label.ManagedBy:       projectName,
				label.Organization:    clusterConfig.Organization,
			},
		},
		Spec: v1alpha1.CertConfigSpec{
			Cert: v1alpha1.CertConfigSpecCert{
				AllowBareDomains:    false,
				ClusterComponent:    certName,
				ClusterID:           clusterConfig.ClusterID,
				CommonName:          clusterConfig.Domain.Prometheus,
				DisableRegeneration: false,
				TTL:                 clusterConfig.CertTTL,
			},
			VersionBundle: v1alpha1.CertConfigSpecVersionBundle{
				Version: clusterConfig.VersionBundleVersion,
			},
		},
	}
}

func newServiceAccountCertConfig(clusterConfig cluster.Config, cert certs.Cert, projectName string) *v1alpha1.CertConfig {
	certName := string(cert)
	return &v1alpha1.CertConfig{
		TypeMeta: apimetav1.TypeMeta{
			Kind:       certKind,
			APIVersion: certAPIVersion,
		},
		ObjectMeta: apimetav1.ObjectMeta{
			Name: key.CertConfigName(clusterConfig.ClusterID, cert),
			Labels: map[string]string{
				label.Cluster:         clusterConfig.ClusterID,
				label.LegacyClusterID: clusterConfig.ClusterID,
				label.LegacyComponent: certName,
				label.ManagedBy:       projectName,
				label.Organization:    clusterConfig.Organization,
			},
		},
		Spec: v1alpha1.CertConfigSpec{
			Cert: v1alpha1.CertConfigSpecCert{
				AllowBareDomains:    false,
				ClusterComponent:    certName,
				ClusterID:           clusterConfig.ClusterID,
				CommonName:          clusterConfig.Domain.ServiceAccount,
				DisableRegeneration: false,
				TTL:                 clusterConfig.CertTTL,
			},
			VersionBundle: v1alpha1.CertConfigSpecVersionBundle{
				Version: clusterConfig.VersionBundleVersion,
			},
		},
	}
}

func newWorkerCertConfig(clusterConfig cluster.Config, cert certs.Cert, projectName string) *v1alpha1.CertConfig {
	certName := string(cert)
	return &v1alpha1.CertConfig{
		TypeMeta: apimetav1.TypeMeta{
			Kind:       certKind,
			APIVersion: certAPIVersion,
		},
		ObjectMeta: apimetav1.ObjectMeta{
			Name: key.CertConfigName(clusterConfig.ClusterID, cert),
			Labels: map[string]string{
				label.Cluster:         clusterConfig.ClusterID,
				label.LegacyClusterID: clusterConfig.ClusterID,
				label.LegacyComponent: certName,
				label.ManagedBy:       projectName,
				label.Organization:    clusterConfig.Organization,
			},
		},
		Spec: v1alpha1.CertConfigSpec{
			Cert: v1alpha1.CertConfigSpecCert{
				AllowBareDomains:    true,
				AltNames:            kubeAltNames,
				ClusterComponent:    certName,
				ClusterID:           clusterConfig.ClusterID,
				CommonName:          clusterConfig.Domain.Worker,
				DisableRegeneration: false,
				TTL:                 clusterConfig.CertTTL,
			},
			VersionBundle: v1alpha1.CertConfigSpecVersionBundle{
				Version: clusterConfig.VersionBundleVersion,
			},
		},
	}
}

func newServerDomain(commonDomain string, cert certs.Cert) (string, error) {
	if !strings.Contains(commonDomain, ".") {
		return "", microerror.Maskf(invalidConfigError, "commonDomain must be a valid domain")
	}

	return string(cert) + "." + strings.TrimLeft(commonDomain, "\t ."), nil
}