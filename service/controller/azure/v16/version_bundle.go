package v16

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "cluster-operator",
				Description: "Added support for creating a cluster configmap for use by managed apps.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "nginx-ingress-controller",
				Description: "Added separate podsecuritypolicy.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "nginx-ingress-controller",
				Description: "Switched security context to non-root user.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "coredns",
				Description: "Added separate podsecuritypolicy.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "coredns",
				Description: "Switched security context to non-root user.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "node-exporter",
				Description: "Added separate podsecuritypolicy.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "node-exporter",
				Description: "Switched security context to non-root user.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "metrics-server",
				Description: "Added separate podsecuritypolicy.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "metrics-server",
				Description: "Switched security context to non-root user.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "kube-state-metrics",
				Description: "Added separate podsecuritypolicy.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "kube-state-metrics",
				Description: "Switched security context to non-root user.",
				Kind:        versionbundle.KindChanged,
			},
			{
				Component:   "external-dns",
				Description: "Added separate podsecuritypolicy.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "external-dns",
				Description: "Switched security context to non-root user.",
				Kind:        versionbundle.KindChanged,
			},
		},
		Components: []versionbundle.Component{
			{
				Name:    "nginx-ingress-controller",
				Version: "0.24.1",
			},
			{
				Name:    "external-dns",
				Version: "0.5.2",
			},
			{
				Name:    "kube-state-metrics",
				Version: "1.5.0",
			},
			{
				Name:    "node-exporter",
				Version: "0.15.1",
			},
			{
				Name:    "coredns",
				Version: "1.5.0",
			},
			{
				Name:    "metrics-server",
				Version: "0.3.1",
			},
		},
		Name:     "cluster-operator",
		Provider: "azure",
		Version:  "0.16.0",
	}
}