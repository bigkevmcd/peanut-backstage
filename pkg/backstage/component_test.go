package backstage

import (
	"strings"
	"testing"

	"github.com/bigkevmcd/peanut-backstage/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	appsv1 "k8s.io/api/apps/v1"
)

func TestParseComponents(t *testing.T) {
	discoverTests := []struct {
		name  string
		items [][]appsv1.Deployment
		want  []Component
	}{
		{
			name: "deployment with no labels",
			items: [][]appsv1.Deployment{
				{
					test.NewDeployment("test", "test-ns"),
				},
			},
			want: []Component{},
		},
		{
			name: "deployment representing single component",
			items: [][]appsv1.Deployment{
				{
					test.NewDeployment("test", "test-ns",
						test.WithLabels(map[string]string{
							instanceLabel:                "mysql-staging",
							nameLabel:                    "mysql",
							componentLabel:               "database",
							createdByLabel:               "test-team",
							partOfLabel:                  "user-db",
							"backstage.io/kubernetes-id": "testing",
						}),
						test.WithAnnotations(map[string]string{
							tagsAnnotation:                           "java,data",
							DescriptionAnnotation:                    "This is a test",
							LifecycleAnnotation:                      "staging",
							"testing.com/annotation":                 "test-annotation",
							"backstage.gitops.pro/link-0":            "https://example.com/user,Example Users,user",
							"backstage.gitops.pro/link-1":            "https://example.com/group,Example Groups,group",
							"backstage.io/kubernetes-label-selector": "app=my-app,component=front-end",
						}),
					),
				},
			},
			want: []Component{
				{
					APIVersion: APIVersion,
					Kind:       KindComponent,
					Metadata: BackstageMetadata{
						Name:        "mysql",
						Description: "This is a test",
						Annotations: map[string]string{
							"backstage.io/kubernetes-label-selector": "app=my-app,component=front-end",
							"backstage.io/kubernetes-id":             "testing",
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
						System:    "user-db",
					},
				},
			},
		},
		{
			name: "multiple deployments, multiple components",
			items: [][]appsv1.Deployment{
				{
					test.NewDeployment("test-1", "test-ns",
						test.WithLabels(map[string]string{
							instanceLabel:                "mysql-production",
							nameLabel:                    "mysql",
							componentLabel:               "database",
							createdByLabel:               "test-team",
							partOfLabel:                  "user-db",
							"backstage.io/kubernetes-id": "testing-production",
						}),
						test.WithAnnotations(map[string]string{
							DescriptionAnnotation: "This is a test",
						}),
					),
					test.NewDeployment("test-2", "test-ns",
						test.WithLabels(map[string]string{
							instanceLabel:                "nginx-production",
							nameLabel:                    "nginx",
							componentLabel:               "webserver",
							createdByLabel:               "test-team",
							partOfLabel:                  "user-db",
							"backstage.io/kubernetes-id": "testing-staging",
						}),
						test.WithAnnotations(map[string]string{
							DescriptionAnnotation: "This is a test",
						}),
					),
				},
			},
			want: []Component{
				{
					APIVersion: "backstage.io/v1alpha1",
					Kind:       "Component",
					Metadata: BackstageMetadata{
						Name:        "mysql",
						Description: "This is a test",
						Annotations: map[string]string{
							"backstage.io/kubernetes-id": "testing-production",
						},
						Tags:  []string{},
						Links: []Link{},
					},
					Spec: ComponentSpec{Type: "database", Owner: "test-team", System: "user-db"},
				},
				{
					APIVersion: "backstage.io/v1alpha1",
					Kind:       "Component",
					Metadata: BackstageMetadata{
						Name:        "nginx",
						Description: "This is a test",
						Annotations: map[string]string{
							"backstage.io/kubernetes-id": "testing-staging",
						},
						Tags:  []string{},
						Links: []Link{},
					},
					Spec: ComponentSpec{Type: "webserver", Owner: "test-team", System: "user-db"},
				},
			},
		},
		// {
		// 	name: "invalid instance label e.g. staging",
		// },
		// {
		// 	name: "invalid tags",
		// },
		// {
		// 	name: "invalid system",
		// },
		// {
		// 	name: "invalid links",
		// },
	}
	strSort := func(x, y string) bool {
		return strings.Compare(x, y) < 0
	}

	for _, tt := range discoverTests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewComponentParser()
			for _, v := range tt.items {
				pods := &appsv1.DeploymentList{
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
