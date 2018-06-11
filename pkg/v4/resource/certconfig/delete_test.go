package certconfig

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned/fake"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/core/v1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientgofake "k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"

	"github.com/giantswarm/cluster-operator/pkg/v4/key"
)

func Test_ApplyDeleteChange_Deletes_deleteChange(t *testing.T) {
	logger, err := micrologger.New(micrologger.Config{})
	if err != nil {
		t.Fatalf("micrologger.New() failed: %#v", err)
	}

	clusterGuestConfig := v1alpha1.ClusterGuestConfig{
		ID: "cluster-1",
	}

	deleteChange := []*v1alpha1.CertConfig{
		newCertConfig("cluster-1", certs.APICert),
		newCertConfig("cluster-1", certs.EtcdCert),
		newCertConfig("cluster-1", certs.PrometheusCert),
		newCertConfig("cluster-1", certs.WorkerCert),
	}

	verificationTable := map[string]bool{
		key.CertConfigName(key.ClusterID(clusterGuestConfig), certs.APICert):        false,
		key.CertConfigName(key.ClusterID(clusterGuestConfig), certs.EtcdCert):       false,
		key.CertConfigName(key.ClusterID(clusterGuestConfig), certs.PrometheusCert): false,
		key.CertConfigName(key.ClusterID(clusterGuestConfig), certs.WorkerCert):     false,
	}

	client := fake.NewSimpleClientset()
	client.ReactionChain = append([]k8stesting.Reactor{
		verifyCertConfigDeletedReactor(t, verificationTable),
	}, client.ReactionChain...)

	r, err := New(Config{
		BaseClusterConfig: newClusterConfig(),
		G8sClient:         client,
		K8sClient:         clientgofake.NewSimpleClientset(),
		Logger:            logger,
		ProjectName:       "cluster-operator",
		Provider:          "kvm",
		ToClusterGuestConfigFunc: func(v interface{}) (v1alpha1.ClusterGuestConfig, error) {
			return v.(v1alpha1.ClusterGuestConfig), nil
		},
		ToClusterObjectMetaFunc: func(v interface{}) (apismetav1.ObjectMeta, error) {
			return apismetav1.ObjectMeta{
				Namespace: v1.NamespaceDefault,
			}, nil
		},
	})

	if err != nil {
		t.Fatalf("Resource construction failed: %#v", err)
	}

	err = r.ApplyDeleteChange(context.TODO(), clusterGuestConfig, deleteChange)
	if err != nil {
		t.Fatalf("ApplyDeleteChange(...) == %#v, want nil", err)
	}

	for k, v := range verificationTable {
		// Was CoreV1alpha1().CertConfigs(...).Delete(...) called for given
		// CertConfig?
		if !v {
			t.Fatalf("ApplyDeleteChange(...) didn't create CertConfig(%s)", k)
		}
	}
}

func Test_ApplyDeleteChange_Does_Not_Make_API_Call_With_Empty_deleteChange(t *testing.T) {
	logger, err := micrologger.New(micrologger.Config{})
	if err != nil {
		t.Fatalf("micrologger.New() failed: %#v", err)
	}

	clusterGuestConfig := v1alpha1.ClusterGuestConfig{
		ID: "cluster-1",
	}

	deleteChange := []*v1alpha1.CertConfig{}

	client := fake.NewSimpleClientset()
	client.ReactionChain = append([]k8stesting.Reactor{
		alwaysReturnErrorReactor(unknownAPIError),
	}, client.ReactionChain...)

	r, err := New(Config{
		BaseClusterConfig: newClusterConfig(),
		G8sClient:         client,
		K8sClient:         clientgofake.NewSimpleClientset(),
		Logger:            logger,
		ProjectName:       "cluster-operator",
		Provider:          "kvm",
		ToClusterGuestConfigFunc: func(v interface{}) (v1alpha1.ClusterGuestConfig, error) {
			return v.(v1alpha1.ClusterGuestConfig), nil
		},
		ToClusterObjectMetaFunc: func(v interface{}) (apismetav1.ObjectMeta, error) {
			return apismetav1.ObjectMeta{
				Namespace: v1.NamespaceDefault,
			}, nil
		},
	})

	if err != nil {
		t.Fatalf("Resource construction failed: %#v", err)
	}

	err = r.ApplyDeleteChange(context.TODO(), clusterGuestConfig, deleteChange)
	if err != nil {
		t.Fatalf("ApplyDeleteChange(...) == %#v, want nil", err)
	}
}

func Test_ApplyDeleteChange_Handles_K8S_API_Error(t *testing.T) {
	logger, err := micrologger.New(micrologger.Config{})
	if err != nil {
		t.Fatalf("micrologger.New() failed: %#v", err)
	}

	clusterGuestConfig := v1alpha1.ClusterGuestConfig{
		ID: "cluster-1",
	}

	deleteChange := []*v1alpha1.CertConfig{
		newCertConfig("cluster-1", certs.APICert),
	}

	client := fake.NewSimpleClientset()
	client.ReactionChain = append([]k8stesting.Reactor{
		alwaysReturnErrorReactor(unknownAPIError),
	}, client.ReactionChain...)

	r, err := New(Config{
		BaseClusterConfig: newClusterConfig(),
		G8sClient:         client,
		K8sClient:         clientgofake.NewSimpleClientset(),
		Logger:            logger,
		ProjectName:       "cluster-operator",
		Provider:          "kvm",
		ToClusterGuestConfigFunc: func(v interface{}) (v1alpha1.ClusterGuestConfig, error) {
			return v.(v1alpha1.ClusterGuestConfig), nil
		},
		ToClusterObjectMetaFunc: func(v interface{}) (apismetav1.ObjectMeta, error) {
			return apismetav1.ObjectMeta{
				Namespace: v1.NamespaceDefault,
			}, nil
		},
	})

	if err != nil {
		t.Fatalf("Resource construction failed: %#v", err)
	}

	err = r.ApplyDeleteChange(context.TODO(), clusterGuestConfig, deleteChange)
	if microerror.Cause(err) != unknownAPIError {
		t.Fatalf("ApplyDeleteChange(...) == %#v, want %#v", err, unknownAPIError)
	}
}

func Test_newDeleteChangeForDeletePatch_Deletes_Existing_CertConfigs(t *testing.T) {
	testCases := []struct {
		name                string
		clusterGuestConfig  v1alpha1.ClusterGuestConfig
		currentState        interface{}
		desiredState        interface{}
		expectedCertConfigs []*v1alpha1.CertConfig
		errorMatcher        func(error) bool
	}{
		{
			name: "case 0: No certconfigs exist, single certconfig desired",
			clusterGuestConfig: v1alpha1.ClusterGuestConfig{
				ID: "cluster-1",
			},
			currentState: nil,
			desiredState: []*v1alpha1.CertConfig{
				newCertConfig("cluster-1", certs.APICert),
			},
			expectedCertConfigs: []*v1alpha1.CertConfig{},
			errorMatcher:        nil,
		},
		{
			name: "case 1: One certconfig exists and it's the same as desired one",
			clusterGuestConfig: v1alpha1.ClusterGuestConfig{
				ID: "cluster-1",
			},
			currentState: []*v1alpha1.CertConfig{
				newCertConfig("cluster-1", certs.APICert),
			},
			desiredState: []*v1alpha1.CertConfig{
				newCertConfig("cluster-1", certs.APICert),
			},
			expectedCertConfigs: []*v1alpha1.CertConfig{
				newCertConfig("cluster-1", certs.APICert),
			},
			errorMatcher: nil,
		},
		{
			name: "case 2: Some of desired certconfigs exist but not all",
			clusterGuestConfig: v1alpha1.ClusterGuestConfig{
				ID: "cluster-1",
			},
			currentState: []*v1alpha1.CertConfig{
				newCertConfig("cluster-1", certs.APICert),
				newCertConfig("cluster-1", certs.CalicoCert),
				newCertConfig("cluster-1", certs.EtcdCert),
			},
			desiredState: []*v1alpha1.CertConfig{
				newCertConfig("cluster-1", certs.APICert),
				newCertConfig("cluster-1", certs.CalicoCert),
				newCertConfig("cluster-1", certs.EtcdCert),
				newCertConfig("cluster-1", certs.FlanneldEtcdClientCert),
				newCertConfig("cluster-1", certs.NodeOperatorCert),
				newCertConfig("cluster-1", certs.PrometheusCert),
				newCertConfig("cluster-1", certs.ServiceAccountCert),
				newCertConfig("cluster-1", certs.WorkerCert),
			},
			expectedCertConfigs: []*v1alpha1.CertConfig{
				newCertConfig("cluster-1", certs.APICert),
				newCertConfig("cluster-1", certs.CalicoCert),
				newCertConfig("cluster-1", certs.EtcdCert),
			},
			errorMatcher: nil,
		},
		{
			name: "case 3: currentState is wrong type",
			clusterGuestConfig: v1alpha1.ClusterGuestConfig{
				ID: "cluster-1",
			},
			currentState: []string{
				"foo",
				"bar",
				"baz",
			},
			desiredState: []*v1alpha1.CertConfig{
				newCertConfig("cluster-1", certs.APICert),
				newCertConfig("cluster-1", certs.CalicoCert),
				newCertConfig("cluster-1", certs.EtcdCert),
			},
			expectedCertConfigs: []*v1alpha1.CertConfig{},
			errorMatcher:        IsWrongType,
		},
	}

	logger, err := micrologger.New(micrologger.Config{})
	if err != nil {
		t.Fatalf("micrologger.New() failed: %#v", err)
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			r, err := New(Config{
				BaseClusterConfig: newClusterConfig(),
				G8sClient:         fake.NewSimpleClientset(),
				K8sClient:         clientgofake.NewSimpleClientset(),
				Logger:            logger,
				ProjectName:       "cluster-operator",
				Provider:          "kvm",
				ToClusterGuestConfigFunc: func(v interface{}) (v1alpha1.ClusterGuestConfig, error) {
					return v.(v1alpha1.ClusterGuestConfig), nil
				},
				ToClusterObjectMetaFunc: func(v interface{}) (apismetav1.ObjectMeta, error) {
					return apismetav1.ObjectMeta{
						Namespace: v1.NamespaceDefault,
					}, nil
				},
			})

			if err != nil {
				t.Fatalf("Resource construction failed: %#v", err)
			}

			certConfigs, err := r.newDeleteChangeForDeletePatch(context.TODO(), tt.clusterGuestConfig, tt.currentState, tt.desiredState)

			switch {
			case err == nil && tt.errorMatcher == nil: // correct; carry on
			case err != nil && tt.errorMatcher != nil:
				if !tt.errorMatcher(err) {
					t.Fatalf("error == %#v, want matching", err)
				}
			case err != nil && tt.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tt.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			}

			// Verify that certconfigs that are expected to be deleted, are the
			// only ones in the returned list of certconfigs that are to be
			// updated.  Order doesn't matter here.
			for _, c := range tt.expectedCertConfigs {
				found := false
				for i := 0; i < len(certConfigs); i++ {
					if reflect.DeepEqual(certConfigs[i], c) {
						// When matching certconfig is found, remove from list
						// returned by newDeleteChangeForDeletePatch(). When
						// all expected certconfigs are iterated, returned list
						// must be empty.
						certConfigs = append(certConfigs[:i], certConfigs[i+1:]...)
						found = true
						break
					}
				}

				if !found {
					t.Fatalf("%#v not found in certConfigs returned by newDeleteChangeForDeletePatch", c)
				}
			}

			// Verify that there aren't any unexpected extra certconfigs going
			// to be deleted.
			if len(certConfigs) > 0 {
				for _, c := range certConfigs {
					t.Errorf("unwanted certconfig present: %#v", c)
				}
			}
		})
	}
}

func Test_newDeleteChangeForUpdatePatch_Deletes_Existing_CertConfigs_That_Are_Not_Desired(t *testing.T) {
	testCases := []struct {
		name                string
		clusterGuestConfig  v1alpha1.ClusterGuestConfig
		currentState        interface{}
		desiredState        interface{}
		expectedCertConfigs []*v1alpha1.CertConfig
		errorMatcher        func(error) bool
	}{
		{
			name: "case 0: No certconfigs exist, single certconfig desired",
			clusterGuestConfig: v1alpha1.ClusterGuestConfig{
				ID: "cluster-1",
			},
			currentState: nil,
			desiredState: []*v1alpha1.CertConfig{
				newCertConfig("cluster-1", certs.APICert),
			},
			expectedCertConfigs: []*v1alpha1.CertConfig{},
			errorMatcher:        nil,
		},
		{
			name: "case 1: One certconfig exists and it's the same as desired one",
			clusterGuestConfig: v1alpha1.ClusterGuestConfig{
				ID: "cluster-1",
			},
			currentState: []*v1alpha1.CertConfig{
				newCertConfig("cluster-1", certs.APICert),
			},
			desiredState: []*v1alpha1.CertConfig{
				newCertConfig("cluster-1", certs.APICert),
			},
			expectedCertConfigs: []*v1alpha1.CertConfig{},
			errorMatcher:        nil,
		},
		{
			name: "case 2: Some of desired certconfigs exist but not all, there are also some leftovers from earlier implementation",
			clusterGuestConfig: v1alpha1.ClusterGuestConfig{
				ID: "cluster-1",
			},
			currentState: []*v1alpha1.CertConfig{
				newCertConfig("cluster-1", certs.APICert),
				newCertConfig("cluster-1", certs.CalicoCert),
				newCertConfig("cluster-1", certs.EtcdCert),
				newCertConfig("cluster-1", "legacy-cert-1"),
				newCertConfig("cluster-1", "legacy-cert-2"),
				newCertConfig("cluster-1", "not needed anymore"),
				newCertConfig("cluster-1", certs.FlanneldEtcdClientCert),
			},
			desiredState: []*v1alpha1.CertConfig{
				newCertConfig("cluster-1", certs.APICert),
				newCertConfig("cluster-1", certs.CalicoCert),
				newCertConfig("cluster-1", certs.EtcdCert),
				newCertConfig("cluster-1", certs.FlanneldEtcdClientCert),
				newCertConfig("cluster-1", certs.NodeOperatorCert),
				newCertConfig("cluster-1", certs.PrometheusCert),
				newCertConfig("cluster-1", certs.ServiceAccountCert),
				newCertConfig("cluster-1", certs.WorkerCert),
			},
			expectedCertConfigs: []*v1alpha1.CertConfig{
				newCertConfig("cluster-1", "legacy-cert-1"),
				newCertConfig("cluster-1", "legacy-cert-2"),
				newCertConfig("cluster-1", "not needed anymore"),
			},
			errorMatcher: nil,
		},
		{
			name: "case 3: currentState is wrong type",
			clusterGuestConfig: v1alpha1.ClusterGuestConfig{
				ID: "cluster-1",
			},
			currentState: []string{
				"foo",
				"bar",
				"baz",
			},
			desiredState: []*v1alpha1.CertConfig{
				newCertConfig("cluster-1", certs.APICert),
				newCertConfig("cluster-1", certs.CalicoCert),
				newCertConfig("cluster-1", certs.EtcdCert),
			},
			expectedCertConfigs: []*v1alpha1.CertConfig{},
			errorMatcher:        IsWrongType,
		},
		{
			name: "case 4: desiredState is wrong type",
			clusterGuestConfig: v1alpha1.ClusterGuestConfig{
				ID: "cluster-1",
			},
			currentState: []*v1alpha1.CertConfig{
				newCertConfig("cluster-1", certs.APICert),
				newCertConfig("cluster-1", certs.CalicoCert),
				newCertConfig("cluster-1", certs.EtcdCert),
			},
			desiredState: []string{
				"foo",
				"bar",
				"baz",
			},
			expectedCertConfigs: []*v1alpha1.CertConfig{},
			errorMatcher:        IsWrongType,
		},
	}

	logger, err := micrologger.New(micrologger.Config{})
	if err != nil {
		t.Fatalf("micrologger.New() failed: %#v", err)
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			r, err := New(Config{
				BaseClusterConfig: newClusterConfig(),
				G8sClient:         fake.NewSimpleClientset(),
				K8sClient:         clientgofake.NewSimpleClientset(),
				Logger:            logger,
				ProjectName:       "cluster-operator",
				Provider:          "kvm",
				ToClusterGuestConfigFunc: func(v interface{}) (v1alpha1.ClusterGuestConfig, error) {
					return v.(v1alpha1.ClusterGuestConfig), nil
				},
				ToClusterObjectMetaFunc: func(v interface{}) (apismetav1.ObjectMeta, error) {
					return apismetav1.ObjectMeta{
						Namespace: v1.NamespaceDefault,
					}, nil
				},
			})

			if err != nil {
				t.Fatalf("Resource construction failed: %#v", err)
			}

			certConfigs, err := r.newDeleteChangeForUpdatePatch(context.TODO(), tt.clusterGuestConfig, tt.currentState, tt.desiredState)

			switch {
			case err == nil && tt.errorMatcher == nil: // correct; carry on
			case err != nil && tt.errorMatcher != nil:
				if !tt.errorMatcher(err) {
					t.Fatalf("error == %#v, want matching", err)
				}
			case err != nil && tt.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tt.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			}

			// Verify that certconfigs that are expected to be deleted, are the
			// only ones in the returned list of certconfigs that are to be
			// updated.  Order doesn't matter here.
			for _, c := range tt.expectedCertConfigs {
				found := false
				for i := 0; i < len(certConfigs); i++ {
					if reflect.DeepEqual(certConfigs[i], c) {
						// When matching certconfig is found, remove from list
						// returned by newDeleteChangeForUpdatePatch(). When
						// all expected certconfigs are iterated, returned list
						// must be empty.
						certConfigs = append(certConfigs[:i], certConfigs[i+1:]...)
						found = true
						break
					}
				}

				if !found {
					t.Fatalf("%#v not found in certConfigs returned by newDeleteChangeForUpdatePatch", c)
				}
			}

			// Verify that there aren't any unexpected extra certconfigs going
			// to be deleted.
			if len(certConfigs) > 0 {
				for _, c := range certConfigs {
					t.Errorf("unwanted certconfig present: %#v", c)
				}
			}
		})
	}
}
