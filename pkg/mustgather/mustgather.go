package mustgather

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/zimnx/kubectl-view-gather/pkg/scheme"
	"github.com/zimnx/kubectl-view-gather/pkg/slices"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

const (
	clusterScopedResourcesPath = "/cluster-scoped"
	namespacesPath             = "/namespaces"
)

type MustGatherArchive struct {
	Path string
}

func NewMustGatherArchive(path string) *MustGatherArchive {
	return &MustGatherArchive{
		Path: path,
	}
}

func (mg *MustGatherArchive) GetAPIGroups() ([]metav1.APIGroup, error) {
	var apiGroups []metav1.APIGroup

	visitFunc := func(resource *APIGroupResource) error {
		versions := map[string]struct{}{}

		err := resource.VisitResources(func(unstr *unstructured.Unstructured, objMetadata *metav1.PartialObjectMetadata) {
			versions[objMetadata.TypeMeta.GroupVersionKind().Version] = struct{}{}
		})
		if err != nil {
			return fmt.Errorf("can't visit resources: %w", err)
		}

		var gvs []metav1.GroupVersionForDiscovery
		for v := range versions {
			gvs = append(gvs, metav1.GroupVersionForDiscovery{
				GroupVersion: resource.APIGroup + "/" + v,
				Version:      v,
			})
		}

		// FIXME: hack - should choose prefered, not shortest
		sort.Slice(gvs, func(i, j int) bool {
			return len(gvs[i].Version) < len(gvs[j].Version)
		})

		apiGroups = append(apiGroups, metav1.APIGroup{
			Name:             resource.APIGroup,
			Versions:         gvs,
			PreferredVersion: gvs[0],
		})

		return nil
	}

	err := mg.visitClusterWideAPIGroupResources(visitFunc)
	if err != nil {
		return nil, fmt.Errorf("can't visit cluster scoped resources: %w", err)
	}

	err = mg.visitNamespacedAPIGroupResources(visitFunc)
	if err != nil {
		return nil, fmt.Errorf("can't visit namespaced resources: %w", err)
	}

	// Filter and make apiGroups unique by group name
	uniqueGroups := make(map[string]metav1.APIGroup)
	for _, g := range apiGroups {
		key := strings.Join(slices.ConvertSlice(g.Versions, func(from metav1.GroupVersionForDiscovery) string {
			return from.GroupVersion
		}), ",")
		uniqueGroups[key] = g
	}
	apiGroups = make([]metav1.APIGroup, 0, len(uniqueGroups))
	for _, g := range uniqueGroups {
		apiGroups = append(apiGroups, g)
	}

	return apiGroups, nil
}

func (mg *MustGatherArchive) GetAPIResources(gv metav1.GroupVersion) ([]metav1.APIResource, error) {
	var apiResources []metav1.APIResource

	visitFunc := func(namespaced bool) func(resource *APIGroupResource) error {
		return func(resource *APIGroupResource) error {
			if resource.APIGroup != gv.Group {
				return nil
			}

			apiResourcesKindMap := map[string]metav1.APIResource{}

			err := resource.VisitResources(func(unstr *unstructured.Unstructured, objMetadata *metav1.PartialObjectMetadata) {
				objGVK := objMetadata.TypeMeta.GroupVersionKind()
				if objGVK.Group != gv.Group || objGVK.Version != gv.Version {
					return
				}

				if _, ok := apiResourcesKindMap[objGVK.Kind]; ok {
					return
				}

				apiResourcesKindMap[objGVK.Kind] = metav1.APIResource{
					Name:         fmt.Sprintf("%ss", strings.ToLower(objGVK.Kind)),
					SingularName: strings.ToLower(objGVK.Kind),
					Namespaced:   namespaced,
					Group:        objGVK.Group,
					Version:      objGVK.Version,
					Kind:         objGVK.Kind,
					Verbs:        []string{"get", "list"},
				}
			})
			if err != nil {
				return fmt.Errorf("can't visit resources: %w", err)
			}

			for _, v := range apiResourcesKindMap {
				apiResources = append(apiResources, v)
			}

			return nil
		}
	}

	err := mg.visitClusterWideAPIGroupResources(visitFunc(false))
	if err != nil {
		return nil, fmt.Errorf("can't visit cluster scoped resources: %w", err)
	}

	err = mg.visitNamespacedAPIGroupResources(visitFunc(true))
	if err != nil {
		return nil, fmt.Errorf("can't visit namespaced resources: %w", err)
	}

	unique := make(map[string]metav1.APIResource)
	for _, r := range apiResources {
		key := r.Group + "|" + r.Version + "|" + r.Kind
		unique[key] = r
	}
	apiResources = make([]metav1.APIResource, 0, len(unique))
	for _, v := range unique {
		apiResources = append(apiResources, v)
	}

	return apiResources, nil
}

type APIGroupResource struct {
	APIGroup   string
	Resource   string
	Namespaced bool

	path string
}

func (gr *APIGroupResource) VisitResources(visitFunc func(unstr *unstructured.Unstructured, objMetadata *metav1.PartialObjectMetadata)) error {
	err := filepath.WalkDir(gr.path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == gr.path {
			return nil
		}

		if d.IsDir() {
			return filepath.SkipDir
		}

		objectRaw, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("can't read file %q: %w", path, err)
		}

		unstr := &unstructured.Unstructured{}
		_, _, err = scheme.Codecs.UniversalDeserializer().Decode(objectRaw, nil, unstr)
		if err != nil {
			return fmt.Errorf("can't deserialize path %q: %w", path, err)
		}

		visitFunc(unstr, &metav1.PartialObjectMetadata{
			TypeMeta: metav1.TypeMeta{
				Kind:       unstr.GetKind(),
				APIVersion: unstr.GetAPIVersion(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:        unstr.GetName(),
				Namespace:   unstr.GetNamespace(),
				Labels:      unstr.GetLabels(),
				Annotations: unstr.GetAnnotations(),
			},
		})

		return nil
	})
	if err != nil {
		return fmt.Errorf("can't visit path %q: %w", gr.path, err)
	}

	return nil
}

func (mg *MustGatherArchive) visitClusterWideAPIGroupResources(visitFunc func(resources *APIGroupResource) error) error {
	clusterScopedAPIResourcesPath := path.Join(mg.Path, clusterScopedResourcesPath)
	return mg.visitAPIGroupResources(clusterScopedAPIResourcesPath, visitFunc)
}

func (mg *MustGatherArchive) visitNamespacedAPIGroupResources(visitFunc func(resources *APIGroupResource) error) error {
	namespacesDirPath := path.Join(mg.Path, namespacesPath)

	err := filepath.WalkDir(namespacesDirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == namespacesDirPath {
			return nil
		}

		if !d.IsDir() {
			return fmt.Errorf("unexpected file on namespaces directory: %s", path)
		}

		err = mg.visitAPIGroupResources(path, visitFunc)
		if err != nil {
			return fmt.Errorf("can't visit path %q: %w", path, err)
		}

		return filepath.SkipDir
	})

	if err != nil {
		return fmt.Errorf("can't visit path %q: %w", namespacesDirPath, err)
	}

	return nil
}

func (mg *MustGatherArchive) visitAPIGroupResources(rootPath string, visitFunc func(resources *APIGroupResource) error) error {
	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == rootPath {
			return nil
		}

		if !d.IsDir() {
			return fmt.Errorf("unexpected file on cluster scoped directory: %s", path)
		}

		dirName := filepath.Base(path)
		dirNamePart := strings.SplitN(dirName, ".", 2)

		resource := dirNamePart[0]
		apiGroup := ""
		if len(dirNamePart) > 1 {
			apiGroup = dirNamePart[1]
		}

		visitErr := visitFunc(&APIGroupResource{
			APIGroup: apiGroup,
			Resource: resource,
			path:     path,
		})
		if visitErr != nil {
			return fmt.Errorf("can't visit path %q: %w", path, err)
		}

		return filepath.SkipDir
	})
	if err != nil {
		return fmt.Errorf("can't visit path %q: %w", rootPath, err)
	}

	return nil
}

func (mg *MustGatherArchive) ListNamespacedObjects(gvr metav1.GroupVersionResource, namespace string) (runtime.Object, error) {
	var objects []unstructured.Unstructured
	err := mg.visitNamespacedAPIGroupResources(func(apiResources *APIGroupResource) error {
		if gvr.Group != apiResources.APIGroup || gvr.Resource != apiResources.Resource {
			return nil
		}

		err := apiResources.VisitResources(func(unstr *unstructured.Unstructured, objMetadata *metav1.PartialObjectMetadata) {
			if namespace != corev1.NamespaceAll && objMetadata.Namespace != namespace {
				return
			}
			objects = append(objects, *unstr)
		})
		if err != nil {
			return fmt.Errorf("can't visit resources: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("can't visit namespaced resources: %w", err)
	}

	return &unstructured.UnstructuredList{
		Object: nil,
		Items:  objects,
	}, nil
}

func (mg *MustGatherArchive) GetNamespacedObject(gvr metav1.GroupVersionResource, nn types.NamespacedName) (runtime.Object, error) {
	var obj *unstructured.Unstructured
	err := mg.visitNamespacedAPIGroupResources(func(apiResources *APIGroupResource) error {
		if gvr.Group != apiResources.APIGroup || gvr.Resource != apiResources.Resource {
			return nil
		}

		err := apiResources.VisitResources(func(unstr *unstructured.Unstructured, objMetadata *metav1.PartialObjectMetadata) {
			if objMetadata.Namespace != nn.Namespace || objMetadata.Name != nn.Name {
				return
			}
			obj = unstr
		})
		if err != nil {
			return fmt.Errorf("can't visit resources: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("can't visit namespaced resources: %w", err)
	}

	return obj, nil
}
func (mg *MustGatherArchive) GetClusterObject(gvr metav1.GroupVersionResource, name string) (runtime.Object, error) {
	var obj *unstructured.Unstructured
	err := mg.visitClusterWideAPIGroupResources(func(apiResources *APIGroupResource) error {
		if gvr.Group != apiResources.APIGroup || gvr.Resource != apiResources.Resource {
			return nil
		}

		err := apiResources.VisitResources(func(unstr *unstructured.Unstructured, objMetadata *metav1.PartialObjectMetadata) {
			if objMetadata.Name != name {
				return
			}
			obj = unstr
		})
		if err != nil {
			return fmt.Errorf("can't visit resources: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("can't visit cluster wide resources: %w", err)
	}

	return obj, nil
}

func (mg *MustGatherArchive) ListClusterObjects(gvr metav1.GroupVersionResource) (runtime.Object, error) {
	var objects []unstructured.Unstructured
	err := mg.visitClusterWideAPIGroupResources(func(apiResources *APIGroupResource) error {
		if gvr.Group != apiResources.APIGroup || gvr.Resource != apiResources.Resource {
			return nil
		}

		err := apiResources.VisitResources(func(unstr *unstructured.Unstructured, objMetadata *metav1.PartialObjectMetadata) {
			objects = append(objects, *unstr)
		})
		if err != nil {
			return fmt.Errorf("can't visit resources: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("can't visit namespaced resources: %w", err)
	}

	return &unstructured.UnstructuredList{
		Object: nil,
		Items:  objects,
	}, nil
}
