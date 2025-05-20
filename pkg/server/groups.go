package server

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

var standardAPIGroupList = []metav1.APIGroup{
	{
		Name: "v1",
		Versions: []metav1.GroupVersionForDiscovery{
			{GroupVersion: "v1", Version: "v1"},
		},
	},
	{
		Name: "admissionregistration.k8s.io",
		Versions: []metav1.GroupVersionForDiscovery{
			{GroupVersion: "admissionregistration.k8s.io/v1", Version: "v1"},
		},
		PreferredVersion: metav1.GroupVersionForDiscovery{
			GroupVersion: "admissionregistration.k8s.io/v1", Version: "v1",
		},
	},
	{
		Name: "apps",
		Versions: []metav1.GroupVersionForDiscovery{
			{GroupVersion: "apps/v1", Version: "v1"},
		},
		PreferredVersion: metav1.GroupVersionForDiscovery{
			GroupVersion: "apps/v1", Version: "v1",
		},
	},
	{
		Name: "authentication.k8s.io",
		Versions: []metav1.GroupVersionForDiscovery{
			{GroupVersion: "authentication.k8s.io/v1", Version: "v1"},
		},
		PreferredVersion: metav1.GroupVersionForDiscovery{
			GroupVersion: "authentication.k8s.io/v1", Version: "v1",
		},
	},
	{
		Name: "authorization.k8s.io",
		Versions: []metav1.GroupVersionForDiscovery{
			{GroupVersion: "authorization.k8s.io/v1", Version: "v1"},
		},
		PreferredVersion: metav1.GroupVersionForDiscovery{
			GroupVersion: "authorization.k8s.io/v1", Version: "v1",
		},
	},
	{
		Name: "autoscaling",
		Versions: []metav1.GroupVersionForDiscovery{
			{GroupVersion: "autoscaling/v1", Version: "v1"},
			{GroupVersion: "autoscaling/v2", Version: "v2"},
			{GroupVersion: "autoscaling/v2beta2", Version: "v2beta2"},
		},
		PreferredVersion: metav1.GroupVersionForDiscovery{
			GroupVersion: "autoscaling/v2", Version: "v2",
		},
	},
	{
		Name: "batch",
		Versions: []metav1.GroupVersionForDiscovery{
			{GroupVersion: "batch/v1", Version: "v1"},
			{GroupVersion: "batch/v1beta1", Version: "v1beta1"}, // deprecated
		},
		PreferredVersion: metav1.GroupVersionForDiscovery{
			GroupVersion: "batch/v1", Version: "v1",
		},
	},
	{
		Name: "certificates.k8s.io",
		Versions: []metav1.GroupVersionForDiscovery{
			{GroupVersion: "certificates.k8s.io/v1", Version: "v1"},
		},
		PreferredVersion: metav1.GroupVersionForDiscovery{
			GroupVersion: "certificates.k8s.io/v1", Version: "v1",
		},
	},
	{
		Name: "coordination.k8s.io",
		Versions: []metav1.GroupVersionForDiscovery{
			{GroupVersion: "coordination.k8s.io/v1", Version: "v1"},
		},
		PreferredVersion: metav1.GroupVersionForDiscovery{
			GroupVersion: "coordination.k8s.io/v1", Version: "v1",
		},
	},
	{
		Name: "discovery.k8s.io",
		Versions: []metav1.GroupVersionForDiscovery{
			{GroupVersion: "discovery.k8s.io/v1", Version: "v1"},
		},
		PreferredVersion: metav1.GroupVersionForDiscovery{
			GroupVersion: "discovery.k8s.io/v1", Version: "v1",
		},
	},
	{
		Name: "events.k8s.io",
		Versions: []metav1.GroupVersionForDiscovery{
			{GroupVersion: "events.k8s.io/v1", Version: "v1"},
		},
		PreferredVersion: metav1.GroupVersionForDiscovery{
			GroupVersion: "events.k8s.io/v1", Version: "v1",
		},
	},
	{
		Name: "flowcontrol.apiserver.k8s.io",
		Versions: []metav1.GroupVersionForDiscovery{
			{GroupVersion: "flowcontrol.apiserver.k8s.io/v1", Version: "v1"},
		},
		PreferredVersion: metav1.GroupVersionForDiscovery{
			GroupVersion: "flowcontrol.apiserver.k8s.io/v1", Version: "v1",
		},
	},
	{
		Name: "networking.k8s.io",
		Versions: []metav1.GroupVersionForDiscovery{
			{GroupVersion: "networking.k8s.io/v1", Version: "v1"},
		},
		PreferredVersion: metav1.GroupVersionForDiscovery{
			GroupVersion: "networking.k8s.io/v1", Version: "v1",
		},
	},
	{
		Name: "node.k8s.io",
		Versions: []metav1.GroupVersionForDiscovery{
			{GroupVersion: "node.k8s.io/v1", Version: "v1"},
		},
		PreferredVersion: metav1.GroupVersionForDiscovery{
			GroupVersion: "node.k8s.io/v1", Version: "v1",
		},
	},
	{
		Name: "policy",
		Versions: []metav1.GroupVersionForDiscovery{
			{GroupVersion: "policy/v1", Version: "v1"},
		},
		PreferredVersion: metav1.GroupVersionForDiscovery{
			GroupVersion: "policy/v1", Version: "v1",
		},
	},
	{
		Name: "rbac.authorization.k8s.io",
		Versions: []metav1.GroupVersionForDiscovery{
			{GroupVersion: "rbac.authorization.k8s.io/v1", Version: "v1"},
		},
		PreferredVersion: metav1.GroupVersionForDiscovery{
			GroupVersion: "rbac.authorization.k8s.io/v1", Version: "v1",
		},
	},
	{
		Name: "scheduling.k8s.io",
		Versions: []metav1.GroupVersionForDiscovery{
			{GroupVersion: "scheduling.k8s.io/v1", Version: "v1"},
		},
		PreferredVersion: metav1.GroupVersionForDiscovery{
			GroupVersion: "scheduling.k8s.io/v1", Version: "v1",
		},
	},
	{
		Name: "storage.k8s.io",
		Versions: []metav1.GroupVersionForDiscovery{
			{GroupVersion: "storage.k8s.io/v1", Version: "v1"},
		},
		PreferredVersion: metav1.GroupVersionForDiscovery{
			GroupVersion: "storage.k8s.io/v1", Version: "v1",
		},
	},
}
