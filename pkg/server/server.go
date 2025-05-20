package server

import (
	"context"
	"encoding/json"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"net/http"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/registry/rest"
)

type APIServerStub struct {
	mux *http.ServeMux
}

func NewAPIServerStub() *APIServerStub {
	s := &APIServerStub{}
	mux := http.NewServeMux()
	mux.Handle("/apis", http.HandlerFunc(s.handleAPIs))
	mux.Handle("/apis/", http.HandlerFunc(s.handleAPIs))
	mux.Handle("/api/", http.HandlerFunc(s.handleAPI))
	mux.Handle("/api", http.HandlerFunc(s.handleAPI))

	mux.Handle("/api/v1/namespaces/", http.HandlerFunc(s.handleNamespaced))

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

	// This should handle both `/apis` and `/apis/discovery.k8s.io/v1` like requests.

	path := strings.TrimPrefix(r.URL.Path, "/apis")
	path = strings.Trim(path, "/")

	if path == "" {
		// Handle GET /apis
		resp := &metav1.APIGroupList{
			TypeMeta: metav1.TypeMeta{
				Kind:       "APIGroupList",
				APIVersion: "v1",
			},
			Groups: standardAPIGroupList,
		}
		writeJSON(w, resp)
		return
	}

	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}

	group := parts[0]
	version := parts[1]

	resp := apiResourceListForGroupVersion(group, version)
	resp.Kind = "APIResourceList"
	resp.APIVersion = "v1"

	writeJSON(w, resp)
}

func apiResourceListForGroupVersion(group string, version string) *metav1.APIResourceList {
	switch group {
	case "apps":
		return &metav1.APIResourceList{
			GroupVersion: "apps/v1",
			APIResources: []metav1.APIResource{
				{Name: "deployments", Namespaced: true, Kind: "Deployment", Verbs: []string{"get", "list", "watch"}},
				{Name: "replicasets", Namespaced: true, Kind: "ReplicaSet", Verbs: []string{"get", "list", "watch"}},
				{Name: "daemonsets", Namespaced: true, Kind: "DaemonSet", Verbs: []string{"get", "list", "watch"}},
				{Name: "statefulsets", Namespaced: true, Kind: "StatefulSet", Verbs: []string{"get", "list", "watch"}},
			},
		}
	default:
		// If not found, return an empty APIResourceList.
		return &metav1.APIResourceList{
			GroupVersion: group + "/" + version,
		}
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
				// add more as needed
			},
		}
		writeJSON(w, resp)
	default:
		http.NotFound(w, r)
	}
}

// handleNamespaced handles `GET /api/v1/namespaces/{namespace}/{resource}` requests.
func (s *APIServerStub) handleNamespaced(w http.ResponseWriter, r *http.Request) {
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

	// TODO: fetch the actual resource from the state

	// return the resource
	tableConvertor := rest.NewDefaultTableConvertor(schema.GroupResource{
		Resource: resource,
	})

	obj := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind: "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			CreationTimestamp: metav1.Time{Time: time.Now()},
			Name:              "example-pod",
			Namespace:         namespace,
		},
	}

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
