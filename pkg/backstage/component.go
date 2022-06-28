package backstage

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	// BackstageAPIVersion Used in all Backstage resources.
	BackstageAPIVersion = "backstage.io/v1alpha1"

	// KindComponent is the kind for Backstage components.
	KindComponent = "Component"

	// TypeService is the type of component as a string, e.g. website.
	TypeService = "service"
)

// Component is a representation of a Backstage Location.
type Component struct {
	APIVersion string            `yaml:"apiVersion"`
	Kind       string            `yaml:"kind"`
	Metadata   BackstageMetadata `yaml:"metadata"`
	Spec       ComponentSpec     `yaml:"spec,omitempty"`
}

// ComponentSpec
type ComponentSpec struct {
	Type      string `yaml:"type"`
	Lifecycle string `yaml:"lifecycle"`
	Owner     string `yaml:"owner"`
	System    string `yaml:"system"`
}

// ComponentParser parses the labels and annotations on runtime Objects and
// extracts components from the labels and annotations.
type ComponentParser struct {
	Accessor   meta.MetadataAccessor
	components map[string]discoveryComponent
}

// NewComponentParser creates and returns a new ComponentParser ready for use.
func NewComponentParser() *ComponentParser {
	return &ComponentParser{
		Accessor:   meta.NewAccessor(),
		components: make(map[string]discoveryComponent),
	}
}

// Add a list of objects to the parser.
//
// The list should be a List type, e.g. PodList, DeploymentList etc.
func (p *ComponentParser) Add(list runtime.Object) error {
	return meta.EachListItem(list, func(obj runtime.Object) error {
		l, err := p.Accessor.Labels(obj)
		if err != nil {
			return fmt.Errorf("failed to get labels from %v: %w", obj, err)
		}
		componentName := l[nameLabel]
		if componentName == "" {
			return nil
		}
		c, ok := p.components[componentName]
		if !ok {
			c = discoveryComponent{
				name: componentName,
			}
		}
		c.createdBy = l[createdByLabel]
		p.components[componentName] = c
		return nil
	})
}

// Components returns the Components that were discovered during the parsing
// process.
func (p *ComponentParser) Components() []Component {
	result := []Component{}
	for _, v := range p.components {
		result = append(result, Component{
			APIVersion: BackstageAPIVersion,
			Kind:       KindComponent,
			Metadata: BackstageMetadata{
				Name: v.name,
			},

			Spec: ComponentSpec{
				Owner: v.createdBy,
			},
		})
	}
	return result
}

type discoveryComponent struct {
	name      string
	createdBy string
}

// apiVersion: backstage.io/v1alpha1
// kind: Component
// metadata:
//   name: artist-lookup
//   description: Artist Lookup
//   annotations:
//     backstage.io/kubernetes-id: artist-lookup
//   tags:
//     - java
//     - data
//   links:
//     - url: https://example.com/user
//       title: Examples Users
//       icon: user
//     - url: https://example.com/group
//       title: Example Group
//       icon: group
//     - url: https://example.com/cloud
//       title: Link with Cloud Icon
//       icon: cloud
//     - url: https://example.com/dashboard
//       title: Dashboard
//       icon: dashboard
//     - url: https://example.com/help
//       title: Support
//       icon: help
//     - url: https://example.com/web
//       title: Website
//       icon: web
//     - url: https://example.com/alert
//       title: Alerts
//       icon: alert
// spec:
//   type: service
//   lifecycle: experimental
//   owner: guests
//   system: examples
