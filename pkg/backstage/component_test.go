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
					})),
				},
			},
			want: []Component{
				{
					APIVersion: BackstageAPIVersion,
					Kind:       KindComponent,
					Metadata: BackstageMetadata{
						Name: "mysql",
					},
					Spec: ComponentSpec{
						Type:      TypeService,
						Lifecycle: "staging",
						Owner:     "test-team",
					},
				},
			},
		},
		// {
		// 	name: "one application, two instances, no parents",
		// 	items: [][]corev1.Pod{
		// 		{
		// 			makePod(withLabels(map[string]string{
		// 				instanceLabel:  "mysql-abcxzy",
		// 				nameLabel:      "mysql",
		// 				componentLabel: "database",
		// 			})),
		// 			makePod(withLabels(map[string]string{
		// 				instanceLabel:  "mysql-deftuv",
		// 				nameLabel:      "mysql",
		// 				componentLabel: "database",
		// 			})),
		// 		},
		// 	},
		// 	want: []Application{
		// 		Application{
		// 			Name:       "mysql",
		// 			Instances:  []string{"mysql-abcxzy", "mysql-deftuv"},
		// 			Components: []string{"database"},
		// 		},
		// 	},
		// },
		// {
		// 	name: "two applications, one instance, with a parent",
		// 	items: [][]corev1.Pod{
		// 		{
		// 			makePod(withLabels(map[string]string{
		// 				instanceLabel:  "mysql-abcxzy",
		// 				nameLabel:      "mysql",
		// 				componentLabel: "database",
		// 				partOfLabel:    "wordpress",
		// 			})),
		// 			makePod(withLabels(map[string]string{
		// 				instanceLabel:  "php-deftuv",
		// 				nameLabel:      "php",
		// 				componentLabel: "web",
		// 				partOfLabel:    "wordpress",
		// 			})),
		// 		},
		// 	},
		// 	want: []Application{
		// 		Application{
		// 			Name:       "mysql",
		// 			Instances:  []string{"mysql-abcxzy"},
		// 			Components: []string{"database"},
		// 			Parents:    []Application{{Name: "wordpress"}},
		// 		},
		// 		{
		// 			Name:       "php",
		// 			Instances:  []string{"php-deftuv"},
		// 			Components: []string{"web"},
		// 			Parents:    []Application{{Name: "wordpress"}},
		// 		},
		// 		{
		// 			Name: "wordpress",
		// 		},
		// 	},
		// },
		// {
		// 	name: "three applications, one instance, with nested parents",
		// 	items: [][]corev1.Pod{
		// 		{
		// 			makePod(withLabels(map[string]string{
		// 				instanceLabel:  "mysql-abcxzy",
		// 				nameLabel:      "mysql",
		// 				componentLabel: "database",
		// 				partOfLabel:    "server",
		// 			})),
		// 			makePod(withLabels(map[string]string{
		// 				instanceLabel:  "php-deftuv",
		// 				nameLabel:      "php",
		// 				componentLabel: "web",
		// 				partOfLabel:    "server",
		// 			})),
		// 			makePod(withLabels(map[string]string{
		// 				instanceLabel:  "php-deftuv",
		// 				nameLabel:      "server",
		// 				componentLabel: "web",
		// 				partOfLabel:    "wordpress",
		// 			})),
		// 		},
		// 	},
		// 	want: []Application{
		// 		Application{
		// 			Name:       "mysql",
		// 			Instances:  []string{"mysql-abcxzy"},
		// 			Components: []string{"database"},
		// 			Parents:    []Application{{Name: "server"}},
		// 		},
		// 		{
		// 			Name:       "php",
		// 			Instances:  []string{"php-deftuv"},
		// 			Components: []string{"web"},
		// 			Parents:    []Application{{Name: "server"}},
		// 		},
		// 		{
		// 			Name:       "server",
		// 			Instances:  []string{"php-deftuv"},
		// 			Components: []string{"web"},
		// 			Parents:    []Application{{Name: "wordpress"}},
		// 		},
		// 		{
		// 			Name: "wordpress",
		// 		},
		// 	},
		// },

		// {
		// 	name: "simple application, with kustomization labels",
		// 	items: [][]corev1.Pod{
		// 		{
		// 			makePod(withLabels(map[string]string{
		// 				instanceLabel:          "mysql-abcxzy",
		// 				nameLabel:              "mysql",
		// 				componentLabel:         "database",
		// 				kustomizationNamespace: "testing",
		// 				kustomizationName:      "abcxzy",
		// 			})),
		// 		},
		// 	},
		// 	want: []Application{
		// 		Application{
		// 			Name:       "mysql",
		// 			Instances:  []string{"mysql-abcxzy"},
		// 			Components: []string{"database"},
		// 			Kustomizations: []types.NamespacedName{
		// 				{Name: "abcxzy", Namespace: "testing"},
		// 			},
		// 		},
		// 	},
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

func makeObjectMetaWithLabels(m map[string]string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Labels: m,
	}
}
