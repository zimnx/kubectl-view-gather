package server

import (
	"encoding/json"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"net/http"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/registry/rest"
)

type APIServerStub struct {
	mux *http.ServeMux
}

func NewAPIServerStub() *APIServerStub {
	s := &APIServerStub{}
	mux := http.NewServeMux()
	mux.Handle("/apis", http.HandlerFunc(s.handleAPIs))
	mux.Handle("/api/", http.HandlerFunc(s.handleAPI))
	mux.Handle("/api", http.HandlerFunc(s.handleAPI))

	mux.Handle()
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

	resp := &v1.APIGroupList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "APIGroupList",
			APIVersion: "v1",
		},
		Groups: standardAPIGroupList,
	}

	writeJSON(w, resp)
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
		// Handle GET /api/v1
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

	resp := map[string]string{
		"namespace": namespace,
		"resource":  resource,
	}
	writeJSON(w, resp)

	// TODO: fetch the actual resource from the state

	// return the resource
	rest.NewDefaultTableConvertor(schema.GroupResource())
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
