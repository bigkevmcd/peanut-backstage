package backstage

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestParseComponents(t *testing.T) {
	discoverTests := []struct {
		name  string
		items [][]corev1.Pod
		want  []Component
	}{
		{
			name: "pods with no labels",
			items: [][]corev1.Pod{
				{
					makePod(),
				},
			},
			want: []Component{},
		},
		{
			name: "single component",
			items: [][]corev1.Pod{
				{
					makePod(withLabels(map[string]string{
						instanceLabel:  "mysql-staging",
						nameLabel:      "mysql",
						componentLabel: "database",
						createdByLabel: "test-team",
					}),
						withAnnotations(map[string]string{
							tagsAnnotation:                "java,data",
							descriptionAnnotation:         "This is a test",
							"testing.com/annotation":      "test-annotation",
							"backstage.io/kubernetes-id":  "testing",
							"backstage.gitops.pro/link-0": "https://example.com/user,Example Users,user",
							"backstage.gitops.pro/link-1": "https://example.com/group,Example Groups,group",
						}),
					),
				},
			},
			want: []Component{
				{
					APIVersion: BackstageAPIVersion,
					Kind:       KindComponent,
					Metadata: BackstageMetadata{
						Name:        "mysql",
						Description: "This is a test",
						Annotations: map[string]string{
							"backstage.io/kubernetes-id": "testing",
						},
						Tags: []string{"data", "java"},
						Links: []Link{
							{
								URL:   "https://example.com/user",
								Title: "Example Users",
								Icon:  "user",
							},
							{
								URL:   "https://example.com/group",
								Title: "Example Groups",
								Icon:  "group",
							},
						},
					},
					Spec: ComponentSpec{
						Type:      "database",
						Lifecycle: "staging",
						Owner:     "test-team",
					},
				},
			},
		},
		// {
		// 	name: "multiple components",
		// },
		// {
		// 	name: "invalid instance label e.g. staging",
		// },
		// {
		// 	name: "invalid tags",
		// },
	}
	strSort := func(x, y string) bool {
		return strings.Compare(x, y) < 0
	}

	for _, tt := range discoverTests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewComponentParser()
			for _, v := range tt.items {
				pods := &corev1.PodList{
					Items: v,
				}

				err := p.Add(pods)
				if err != nil {
					t.Fatal(err)
				}
			}
			components := p.Components()
			if diff := cmp.Diff(tt.want, components, cmpopts.SortSlices(strSort)); diff != "" {
				t.Fatalf("failed discovery:\n%s", diff)
			}
		})
	}
}

func makePod(opts ...func(runtime.Object)) corev1.Pod {
	p := corev1.Pod{}
	for _, o := range opts {
		o(&p)
	}
	return p
}

func withLabels(m map[string]string) func(runtime.Object) {
	var accessor = meta.NewAccessor()
	return func(obj runtime.Object) {
		accessor.SetLabels(obj, m)
	}
}

func withAnnotations(m map[string]string) func(runtime.Object) {
	var accessor = meta.NewAccessor()
	return func(obj runtime.Object) {
		accessor.SetAnnotations(obj, m)
	}
}

func makeObjectMetaWithLabels(m map[string]string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Labels: m,
	}
}
