package backstage

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/exp/maps"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	// APIVersion used in all Backstage resources.
	APIVersion = "backstage.io/v1alpha1"

	// KindComponent is the kind for Backstage components.
	KindComponent = "Component"
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
//
// Labels are based on https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/
func (p *ComponentParser) Add(list runtime.Object) error {
	return meta.EachListItem(list, func(obj runtime.Object) error {
		labels, err := p.Accessor.Labels(obj)
		if err != nil {
			return fmt.Errorf("failed to get labels from %v: %w", obj, err)
		}
		componentName := labels[nameLabel]
		if componentName == "" {
			return nil
		}
		// TODO: split the parsing of the component out
		c, ok := p.components[componentName]
		if !ok {
			c = discoveryComponent{
				name: componentName,
			}
		}
		// TODO: what should we do about multiple discovered values for this?
		// Backstage keys are not generally "multi-value"
		// Perhaps create a set to store them in discovery
		// but with a `.one() (string, err)

		c.createdBy = labels[createdByLabel]
		c.componentType = labels[componentLabel]
		c.system = labels[partOfLabel]
		c.lifecycle = labels[lifecycleLabel]
		// TODO: annotate the component with something to indicate which
		// resource it came from

		annotations, err := p.Accessor.Annotations(obj)
		if err != nil {
			return fmt.Errorf("failed to get annotations from %v: %w", obj, err)
		}

		if rawTags := strings.Split(annotations[tagsAnnotation], ","); len(rawTags) != 0 {
			tags := []string{}
			for _, v := range rawTags {
				if s := strings.TrimSpace(v); s != "" {
					tags = append(tags, s)
				}
			}
			c.tags = tags
		}
		c.description = annotations[descriptionAnnotation]

		// this merges labels and annotations from the K8s resource into the
		// annotations on the component if they are backstage annotations.
		// i.e. start with backstage.io
		// keys in labels override keys in annotations.
		c.annotations = backstageAnnotations(annotations)
		maps.Copy(c.annotations, backstageAnnotations(labels))

		links, err := parseLinkAnnotations(annotations)
		if err != nil {
			return fmt.Errorf("failed to parse links in annotations: %w", err)
		}
		c.links = links

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
			APIVersion: APIVersion,
			Kind:       KindComponent,
			Metadata: BackstageMetadata{
				Name:        v.name,
				Tags:        v.tags,
				Description: v.description,
				Annotations: v.annotations,
				Links:       v.links,
			},
			Spec: ComponentSpec{
				Owner:     v.createdBy,
				Type:      v.componentType,
				Lifecycle: v.lifecycle,
				System:    v.system,
			},
		})
	}

	return result
}

type discoveryComponent struct {
	name          string
	description   string
	createdBy     string
	lifecycle     string
	system        string
	tags          []string
	links         []Link
	componentType string
	annotations   map[string]string
}

func backstageAnnotations(src map[string]string) map[string]string {
	dst := map[string]string{}
	for k, v := range src {
		if parts := strings.SplitN(k, "/", 2); len(parts) == 2 {
			if parts[0] == "backstage.io" {
				dst[k] = v
			}
		}
	}

	return dst
}

func parseLinkAnnotations(annotations map[string]string) ([]Link, error) {
	type link struct {
		seq   int
		url   string
		title string
		icon  string
	}
	links := []link{}

	for k, v := range annotations {
		if strings.HasPrefix(k, urlAnnotationPrefix) {
			seq, err := strconv.Atoi(strings.TrimPrefix(k, urlAnnotationPrefix))
			if err != nil {
				return nil, fmt.Errorf("failed to parse link annotation %q: %w", k, err)
			}
			// TODO: check which fields are required for Links
			if parts := strings.SplitN(v, ",", 3); len(parts) == 3 {
				links = append(links, link{
					seq:   seq,
					url:   strings.TrimSpace(parts[0]),
					title: strings.TrimSpace(parts[1]),
					icon:  strings.TrimSpace(parts[2]),
				})
			}
		}
	}
	sort.Slice(links, func(i, j int) bool { return links[i].seq < links[j].seq })
	result := []Link{}
	for _, v := range links {
		result = append(result, Link{URL: v.url, Title: v.title, Icon: v.icon})
	}

	return result, nil
}
