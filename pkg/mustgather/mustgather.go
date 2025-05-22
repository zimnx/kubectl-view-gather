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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
	clusterScopedAPIResourcesPath := path.Join(mg.Path, clusterScopedResourcesPath)

	var apiGroups []metav1.APIGroup

	visitFunc := func(resource *APIGroupResource) error {
		versions := map[string]struct{}{}

		err := resource.VisitResources(func(objMetadata *metav1.PartialObjectMetadata) {
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

	err := mg.visitAPIGroupResources(clusterScopedAPIResourcesPath, visitFunc)
	if err != nil {
		return nil, fmt.Errorf("can't visit cluster scoped resources: %w", err)
	}

	err = mg.visitNamespacedAPIGroupResources(visitFunc)
	if err != nil {
		return nil, fmt.Errorf("can't visit namespaced resources: %w", err)
	}

	return apiGroups, nil

}

func (mg *MustGatherArchive) GetAPIResources(gv metav1.GroupVersion) ([]metav1.APIResource, error) {
	clusterScopedAPIResourcesPath := path.Join(mg.Path, clusterScopedResourcesPath)

	var apiResources []metav1.APIResource

	visitFunc := func(namespaced bool) func(resource *APIGroupResource) error {
		return func(resource *APIGroupResource) error {
			if resource.APIGroup != gv.Group {
				return nil
			}

			err := resource.VisitResources(func(objMetadata *metav1.PartialObjectMetadata) {
				objGVK := objMetadata.TypeMeta.GroupVersionKind()
				if objGVK.Group != gv.Group || objGVK.Version != gv.Version {
					return
				}

				apiResources = append(apiResources, metav1.APIResource{
					Name:         fmt.Sprintf("%ss", strings.ToLower(objGVK.Kind)),
					SingularName: strings.ToLower(objGVK.Kind),
					Namespaced:   namespaced,
					Group:        objGVK.Group,
					Version:      objGVK.Version,
					Kind:         objGVK.Kind,
					Verbs:        []string{"get", "list"},
				})

			})
			if err != nil {
				return fmt.Errorf("can't visit resources: %w", err)
			}

			return nil
		}
	}

	err := mg.visitAPIGroupResources(clusterScopedAPIResourcesPath, visitFunc(false))
	if err != nil {
		return nil, fmt.Errorf("can't visit cluster scoped resources: %w", err)
	}

	err = mg.visitNamespacedAPIGroupResources(visitFunc(true))
	if err != nil {
		return nil, fmt.Errorf("can't visit namespaced resources: %w", err)
	}

	return apiResources, nil
}

type APIGroupResource struct {
	APIGroup   string
	Resource   string
	Namespaced bool

	path string
}

func (gr *APIGroupResource) VisitResources(visitFunc func(objMetadata *metav1.PartialObjectMetadata)) error {
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

		visitFunc(&metav1.PartialObjectMetadata{
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
