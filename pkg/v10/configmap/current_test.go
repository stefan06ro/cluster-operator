package configmap

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/micrologger/microloggertest"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/giantswarm/cluster-operator/pkg/label"
)

func Test_ConfigMap_GetCurrentState(t *testing.T) {
	testCases := []struct {
		name               string
		config             ClusterConfig
		presentConfigMaps  []*corev1.ConfigMap
		expectedConfigMaps []*corev1.ConfigMap
	}{
		{
			name: "case 0: no results",
			config: ClusterConfig{
				APIDomain:  "5xchu.aws.giantswarm.io",
				ClusterID:  "5xchu",
				Namespaces: []string{},
			},
			presentConfigMaps:  []*corev1.ConfigMap{},
			expectedConfigMaps: []*corev1.ConfigMap{},
		},
		{
			name: "case 1: single result",
			config: ClusterConfig{
				APIDomain:  "5xchu.aws.giantswarm.io",
				ClusterID:  "5xchu",
				Namespaces: []string{},
			},
			presentConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							label.ServiceType: label.ServiceTypeManaged,
							label.ManagedBy:   "cluster-operator",
						},
						Name:      "test-configmap",
						Namespace: metav1.NamespaceSystem,
					},
					Data: map[string]string{
						"test": "test",
					},
				},
			},
			expectedConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							label.ServiceType: label.ServiceTypeManaged,
							label.ManagedBy:   "cluster-operator",
						},
						Name:      "test-configmap",
						Namespace: metav1.NamespaceSystem,
					},
					Data: map[string]string{
						"test": "test",
					},
				},
			},
		},
		{
			name: "case 2: multiple results",
			config: ClusterConfig{
				APIDomain:  "5xchu.aws.giantswarm.io",
				ClusterID:  "5xchu",
				Namespaces: []string{},
			},
			presentConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							label.ServiceType: label.ServiceTypeManaged,
							label.ManagedBy:   "cluster-operator",
						},
						Name:      "test-configmap",
						Namespace: metav1.NamespaceSystem,
					},
					Data: map[string]string{
						"test": "test",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							label.ServiceType: label.ServiceTypeManaged,
							label.ManagedBy:   "cluster-operator",
						},
						Name:      "another-configmap",
						Namespace: metav1.NamespaceSystem,
					},
					Data: map[string]string{
						"test": "test",
					},
				},
			},
			expectedConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							label.ServiceType: label.ServiceTypeManaged,
							label.ManagedBy:   "cluster-operator",
						},
						Name:      "test-configmap",
						Namespace: metav1.NamespaceSystem,
					},
					Data: map[string]string{
						"test": "test",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							label.ServiceType: label.ServiceTypeManaged,
							label.ManagedBy:   "cluster-operator",
						},
						Name:      "another-configmap",
						Namespace: metav1.NamespaceSystem,
					},
					Data: map[string]string{
						"test": "test",
					},
				},
			},
		},
		{
			name: "case 3: multiple namespaces, single result",
			config: ClusterConfig{
				APIDomain: "5xchu.aws.giantswarm.io",
				ClusterID: "5xchu",
				Namespaces: []string{
					"giantswarm",
				},
			},
			presentConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							label.ServiceType: label.ServiceTypeManaged,
							label.ManagedBy:   "cluster-operator",
						},
						Name:      "giantswarm-configmap",
						Namespace: "giantswarm",
					},
					Data: map[string]string{
						"test": "test",
					},
				},
			},
			expectedConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							label.ServiceType: label.ServiceTypeManaged,
							label.ManagedBy:   "cluster-operator",
						},
						Name:      "giantswarm-configmap",
						Namespace: "giantswarm",
					},
					Data: map[string]string{
						"test": "test",
					},
				},
			},
		},
		{
			name: "case 4: multiple namespaces, multiple results",
			config: ClusterConfig{
				APIDomain: "5xchu.aws.giantswarm.io",
				ClusterID: "5xchu",
				Namespaces: []string{
					"giantswarm",
				},
			},
			presentConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							label.ServiceType: label.ServiceTypeManaged,
							label.ManagedBy:   "cluster-operator",
						},
						Name:      "giantswarm-configmap",
						Namespace: "giantswarm",
					},
					Data: map[string]string{
						"test": "test",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							label.ServiceType: label.ServiceTypeManaged,
							label.ManagedBy:   "cluster-operator",
						},
						Name:      "test-configmap",
						Namespace: metav1.NamespaceSystem,
					},
					Data: map[string]string{
						"test": "test",
					},
				},
			},
			expectedConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							label.ServiceType: label.ServiceTypeManaged,
							label.ManagedBy:   "cluster-operator",
						},
						Name:      "giantswarm-configmap",
						Namespace: "giantswarm",
					},
					Data: map[string]string{
						"test": "test",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							label.ServiceType: label.ServiceTypeManaged,
							label.ManagedBy:   "cluster-operator",
						},
						Name:      "test-configmap",
						Namespace: metav1.NamespaceSystem,
					},
					Data: map[string]string{
						"test": "test",
					},
				},
			},
		},
		{
			name: "case 5: multiple namespaces, multiple results, omitting ones without required labels",
			config: ClusterConfig{
				APIDomain: "5xchu.aws.giantswarm.io",
				ClusterID: "5xchu",
				Namespaces: []string{
					"giantswarm",
				},
			},
			presentConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							label.ServiceType: label.ServiceTypeManaged,
							label.ManagedBy:   "cluster-operator",
						},
						Name:      "giantswarm-configmap",
						Namespace: "giantswarm",
					},
					Data: map[string]string{
						"test": "test",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							label.ServiceType: label.ServiceTypeManaged,
							label.ManagedBy:   "cluster-operator",
						},
						Name:      "test-configmap",
						Namespace: metav1.NamespaceSystem,
					},
					Data: map[string]string{
						"test": "test",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							label.ServiceType: label.ServiceTypeManaged,
						},
						Name:      "test-configmap2",
						Namespace: metav1.NamespaceSystem,
					},
					Data: map[string]string{
						"test": "test2",
					},
				},
			},
			expectedConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							label.ServiceType: label.ServiceTypeManaged,
							label.ManagedBy:   "cluster-operator",
						},
						Name:      "giantswarm-configmap",
						Namespace: "giantswarm",
					},
					Data: map[string]string{
						"test": "test",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							label.ServiceType: label.ServiceTypeManaged,
							label.ManagedBy:   "cluster-operator",
						},
						Name:      "test-configmap",
						Namespace: metav1.NamespaceSystem,
					},
					Data: map[string]string{
						"test": "test",
					},
				},
			},
		},
		{
			name: "case 6: as case 5, but different label: multiple namespaces, multiple results, omitting ones without required labels",
			config: ClusterConfig{
				APIDomain: "5xchu.aws.giantswarm.io",
				ClusterID: "5xchu",
				Namespaces: []string{
					"giantswarm",
				},
			},
			presentConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							label.ServiceType: label.ServiceTypeManaged,
							label.ManagedBy:   "cluster-operator",
						},
						Name:      "giantswarm-configmap",
						Namespace: "giantswarm",
					},
					Data: map[string]string{
						"test": "test",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							label.ServiceType: label.ServiceTypeManaged,
							label.ManagedBy:   "cluster-operator",
						},
						Name:      "test-configmap",
						Namespace: metav1.NamespaceSystem,
					},
					Data: map[string]string{
						"test": "test",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							label.ManagedBy: "cluster-operator",
						},
						Name:      "test-configmap2",
						Namespace: metav1.NamespaceSystem,
					},
					Data: map[string]string{
						"test": "test2",
					},
				},
			},
			expectedConfigMaps: []*corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							label.ServiceType: label.ServiceTypeManaged,
							label.ManagedBy:   "cluster-operator",
						},
						Name:      "giantswarm-configmap",
						Namespace: "giantswarm",
					},
					Data: map[string]string{
						"test": "test",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							label.ServiceType: label.ServiceTypeManaged,
							label.ManagedBy:   "cluster-operator",
						},
						Name:      "test-configmap",
						Namespace: metav1.NamespaceSystem,
					},
					Data: map[string]string{
						"test": "test",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			objs := make([]runtime.Object, 0, len(tc.presentConfigMaps))
			for _, cc := range tc.presentConfigMaps {
				objs = append(objs, cc)
			}

			fakeTenantK8sClient := fake.NewSimpleClientset(objs...)
			tenantService := &tenantMock{
				fakeTenantK8sClient: fakeTenantK8sClient,
			}

			c := Config{
				Logger: microloggertest.New(),
				Tenant: tenantService,

				ProjectName: "cluster-operator",
			}
			newService, err := New(c)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}

			configMaps, err := newService.GetCurrentState(context.TODO(), tc.config)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}

			if len(configMaps) != len(tc.expectedConfigMaps) {
				t.Fatalf("expected %d configsmaps got %d", len(tc.expectedConfigMaps), len(configMaps))
			}

			for _, cm := range configMaps {
				found := false
				for _, ec := range tc.expectedConfigMaps {
					if reflect.DeepEqual(cm, ec) {
						found = true
						break
					}
				}

				if !found {
					t.Fatalf("unexpected configmap %#v among returned values", *cm)
				}
			}

		})
	}
}