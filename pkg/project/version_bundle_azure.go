package project

import (
	"github.com/giantswarm/versionbundle"
)

var versionBundleAzure = versionbundle.Bundle{
	Changelogs: []versionbundle.Changelog{
		{
			Component:   "net-exporter",
			Description: "Updated to version 1.6.0.",
			Kind:        versionbundle.KindChanged,
			URLs: []string{
				"https://github.com/giantswarm/net-exporter/releases/tag/v1.6.0",
			},
		},
	},
	Components: []versionbundle.Component{
		{
			Name:    "kube-state-metrics",
			Version: "1.9.2",
		},
		{
			Name:    "nginx-ingress-controller",
			Version: "0.27.1",
		},
		{
			Name:    "node-exporter",
			Version: "0.18.1",
		},
		{
			Name:    "coredns",
			Version: "1.6.5",
		},
		{
			Name:    "metrics-server",
			Version: "0.3.3",
		},
		{
			Name:    "external-dns",
			Version: "0.5.11",
		},
		{
			Name:    "net-exporter",
			Version: "1.6.0",
		},
	},
	Name:     "cluster-operator",
	Provider: "azure",
	Version:  BundleVersion(),
}
