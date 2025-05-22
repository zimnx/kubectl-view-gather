package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apiserver/pkg/registry/rest"
)

// APIGroupsMetaStore is an interface that defines methods to retrieve API groups and resources.
type APIGroupsMetaStore interface {
	GetAPIGroups() ([]metav1.APIGroup, error)
	GetAPIResources(gv metav1.GroupVersion) ([]metav1.APIResource, error)
}

// ObjectStore is an interface that defines methods to retrieve objects by namespace and name.
type ObjectStore interface {
	ListNamespacedObjects(gvr metav1.GroupVersionResource, namespace string) (runtime.Object, error)
	ListClusterObjects(gvr metav1.GroupVersionResource) (runtime.Object, error)
	GetNamespacedObject(gvr metav1.GroupVersionResource, nn types.NamespacedName) (runtime.Object, error)
	GetClusterObject(gvr metav1.GroupVersionResource, name string) (runtime.Object, error)
}

type APIServerStub struct {
	mux         *http.ServeMux
	metaStore   APIGroupsMetaStore
	objectStore ObjectStore
}

func NewAPIServerStub(metaStore APIGroupsMetaStore, objectStore ObjectStore) *APIServerStub {
	s := &APIServerStub{
		metaStore:   metaStore,
		objectStore: objectStore,
	}

	mux := http.NewServeMux()
	mux.Handle("/apis", http.HandlerFunc(s.handleAPIs))
	mux.Handle("/apis/", http.HandlerFunc(s.handleAPIs))
	mux.Handle("/api/", http.HandlerFunc(s.handleAPI))
	mux.Handle("/api", http.HandlerFunc(s.handleAPI))
	mux.Handle("/api/v1/{resource}/", http.HandlerFunc(s.handleV1ClusterWideList))
	mux.Handle("/api/v1/namespaces/{namespace}/{resource}", http.HandlerFunc(s.handleNamespacedV1List))
	mux.Handle("/api/v1/namespaces/{namespace}/{resourceName}/{objectName}", http.HandlerFunc(s.handleNamespacedGetV1))
	mux.Handle("/apis/{apiGroup}/{apiVersion}/{resource}", http.HandlerFunc(s.handleClusterWideList))
	mux.Handle("/apis/{apiGroup}/{apiVersion}/{resource}/{objectName}", http.HandlerFunc(s.handleClusterWideGet))
	mux.Handle("/apis/{apiGroup}/{apiVersion}/namespaces/{namespace}/{resourceName}", http.HandlerFunc(s.handleNamespacedListing))
	mux.Handle("/apis/{apiGroup}/{apiVersion}/namespaces/{namespace}/{resourceName}/{objectName}", http.HandlerFunc(s.handleNamespacedGet))

	s.mux = mux
	return s
}

func (s *APIServerStub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *APIServerStub) handleAPIs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/apis")
	path = strings.Trim(path, "/")

	if path == "" {
		// Handle GET /apis
		apiGroups, err := s.metaStore.GetAPIGroups()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp := &metav1.APIGroupList{
			TypeMeta: metav1.TypeMeta{
				Kind:       "APIGroupList",
				APIVersion: "v1",
			},
			Groups: apiGroups,
		}
		writeJSON(w, resp)
		return
	}

	var group, version string

	parts := strings.Split(path, "/")
	if len(parts) == 1 {
		version = parts[0]
	} else if len(parts) == 2 {
		group = parts[0]
		version = parts[1]
	} else {
		http.NotFound(w, r)
		return
	}

	resp := s.getAPIResourceList(group, version)
	writeJSON(w, resp)
}

func (s *APIServerStub) getAPIResourceList(group string, version string) *metav1.APIResourceList {
	resources, err := s.metaStore.GetAPIResources(metav1.GroupVersion{Group: group, Version: version})
	if err != nil {
		// If the group/version is not found, return an empty APIResourceList.
		// TODO: handle error better way?
		return &metav1.APIResourceList{
			TypeMeta: metav1.TypeMeta{
				Kind:       "APIResourceList",
				APIVersion: "v1",
			},
			GroupVersion: group + "/" + version,
		}
	}

	// If the group/version is found, return the APIResourceList.
	return &metav1.APIResourceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "APIResourceList",
			APIVersion: "v1",
		},
		GroupVersion: group + "/" + version,
		APIResources: resources,
	}
}

func (s *APIServerStub) handleAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api")
	path = strings.Trim(path, "/")

	switch path {
	case "":
		// Handle GET /api
		resp := &metav1.APIVersions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "APIVersions",
				APIVersion: "v1",
			},
			Versions: []string{"v1"},
		}
		writeJSON(w, resp)
	case "v1":
		// Handle GET /api/v1 (looks like a special case as the rest is handled under /apis/{resource}/{version}).
		resp := &metav1.APIResourceList{
			GroupVersion: "v1",
			APIResources: []metav1.APIResource{
				{Name: "pods", Namespaced: true, Kind: "Pod", Verbs: []string{"get", "list", "watch"}},
				{Name: "services", Namespaced: true, Kind: "Service", Verbs: []string{"get", "list", "watch"}},
				{Name: "namespaces", Namespaced: false, Kind: "Namespace", Verbs: []string{"get", "list", "watch"}},
			},
		}
		writeJSON(w, resp)
	default:
		http.NotFound(w, r)
	}
}

// handleNamespacedV1List handles `GET /api/v1/namespaces/{namespace}/{resource}` requests.
func (s *APIServerStub) handleNamespacedV1List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/namespaces/")
	path = strings.Trim(path, "/")

	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}

	namespace := parts[0]
	resource := parts[1]

	gvr := metav1.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: resource,
	}

	objList, err := s.objectStore.ListNamespacedObjects(gvr, namespace)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// return the resource
	tableConvertor := rest.NewDefaultTableConvertor(schema.GroupResource{
		Resource: resource,
	})

	tbl, err := tableConvertor.ConvertToTable(context.Background(), objList, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tbl.Kind = "Table"
	tbl.APIVersion = "meta.k8s.io/v1"

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tbl); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleV1ClusterWideList handles `/api/v1/{resourceName}}/"` requests.
func (s *APIServerStub) handleV1ClusterWideList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/")
	path = strings.Trim(path, "/")

	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}

	version := parts[0]
	resourceName := parts[1]

	resource := "" + "/" + version + "/" + resourceName
	gvr := metav1.GroupVersionResource{
		Group:    "",
		Version:  version,
		Resource: resourceName,
	}

	obj, err := s.objectStore.ListClusterObjects(gvr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert into table
	tableConvertor := rest.NewDefaultTableConvertor(schema.GroupResource{
		Resource: resource,
	})
	tbl, err := tableConvertor.ConvertToTable(context.Background(), obj, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tbl.Kind = "Table"
	tbl.APIVersion = "meta.k8s.io/v1"
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tbl); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleNamespacedGetV1 handles `/api/v1/namespaces/{namespace}/{resourceName}/{objectName}` requests.
func (s *APIServerStub) handleNamespacedGetV1(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/")
	path = strings.Trim(path, "/")

	parts := strings.SplitN(path, "/", 5)
	if len(parts) != 5 {
		http.NotFound(w, r)
		return
	}

	version := parts[0]
	namespace := parts[2]
	resourceName := parts[3]
	objectName := parts[4]

	resource := "" + "/" + version + "/" + resourceName
	gvr := metav1.GroupVersionResource{
		Group:    "",
		Version:  version,
		Resource: resourceName,
	}

	obj, err := s.objectStore.GetNamespacedObject(gvr, types.NamespacedName{Namespace: namespace, Name: objectName})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert into table
	tableConvertor := rest.NewDefaultTableConvertor(schema.GroupResource{
		Resource: resource,
	})
	tbl, err := tableConvertor.ConvertToTable(context.Background(), obj, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tbl.Kind = "Table"
	tbl.APIVersion = "meta.k8s.io/v1"
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tbl); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleClusterWideList handles `/apis/{apiGroup}/{apiVersion}/{resource}` requests.
func (s *APIServerStub) handleClusterWideList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/apis/")
	path = strings.Trim(path, "/")

	parts := strings.SplitN(path, "/", 3)
	if len(parts) != 3 {
		http.NotFound(w, r)
		return
	}

	group := parts[0]
	version := parts[1]
	resourceName := parts[2]

	resource := group + "/" + version + "/" + resourceName
	gvr := metav1.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resourceName,
	}

	objList, err := s.objectStore.ListClusterObjects(gvr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// return the resource
	tableConvertor := rest.NewDefaultTableConvertor(schema.GroupResource{
		Resource: resource,
	})

	tbl, err := tableConvertor.ConvertToTable(context.Background(), objList, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tbl.Kind = "Table"
	tbl.APIVersion = "meta.k8s.io/v1"

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tbl); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleNamespacedListing handles `/apis/{apiGroup}/{apiVersion}/namespaces/{namespace}/{resourceName}` requests.
func (s *APIServerStub) handleNamespacedListing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/apis/")
	path = strings.Trim(path, "/")

	parts := strings.SplitN(path, "/", 5)
	if len(parts) != 5 {
		http.NotFound(w, r)
		return
	}

	group := parts[0]
	version := parts[1]
	namespace := parts[3]
	resourceName := parts[4]

	resource := group + "/" + version + "/" + resourceName
	gvr := metav1.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resourceName,
	}

	// List objects in the specified namespace
	objList, err := s.objectStore.ListNamespacedObjects(gvr, namespace)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// return the resource
	tableConvertor := rest.NewDefaultTableConvertor(schema.GroupResource{
		Resource: resource,
	})

	tbl, err := tableConvertor.ConvertToTable(context.Background(), objList, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tbl.Kind = "Table"
	tbl.APIVersion = "meta.k8s.io/v1"

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tbl); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleClusterWideGet handles `/apis/{apiGroup}/{apiVersion}/{resource}/{objectName}` requests.
func (s *APIServerStub) handleClusterWideGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/apis/")
	path = strings.Trim(path, "/")

	parts := strings.SplitN(path, "/", 4)
	if len(parts) != 4 {
		http.NotFound(w, r)
		return
	}

	group := parts[0]
	version := parts[1]
	resourceName := parts[3]
	objectName := parts[4]

	resource := group + "/" + version + "/" + resourceName
	gvr := metav1.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resourceName,
	}

	obj, err := s.objectStore.GetClusterObject(gvr, objectName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert into table
	tableConvertor := rest.NewDefaultTableConvertor(schema.GroupResource{
		Resource: resource,
	})
	tbl, err := tableConvertor.ConvertToTable(context.Background(), obj, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tbl.Kind = "Table"
	tbl.APIVersion = "meta.k8s.io/v1"
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tbl); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleNamespacedGet handles `/apis/{apiGroup}/{apiVersion}/namespaces/{namespace}/{resourceName}/{objectName}` requests.
func (s *APIServerStub) handleNamespacedGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/apis/")
	path = strings.Trim(path, "/")

	parts := strings.SplitN(path, "/", 6)
	if len(parts) != 6 {
		http.NotFound(w, r)
		return
	}

	group := parts[0]
	version := parts[1]
	namespace := parts[3]
	resourceName := parts[4]
	objectName := parts[5]

	resource := group + "/" + version + "/" + resourceName
	gvr := metav1.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resourceName,
	}

	obj, err := s.objectStore.GetNamespacedObject(gvr, types.NamespacedName{Namespace: namespace, Name: objectName})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert into table
	tableConvertor := rest.NewDefaultTableConvertor(schema.GroupResource{
		Resource: resource,
	})
	tbl, err := tableConvertor.ConvertToTable(context.Background(), obj, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tbl.Kind = "Table"
	tbl.APIVersion = "meta.k8s.io/v1"
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tbl); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
