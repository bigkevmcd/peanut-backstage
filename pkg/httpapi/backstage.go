package httpapi

const (
	apiVersion   = "backstage.io/v1alpha1"
	locationKind = "Location"
)

func newLocation(name, description string) *Location {
	return &Location{
		APIVersion: apiVersion,
		Kind:       locationKind,
		Metadata: metadata{
			Name:        name,
			Description: description,
		},
	}
}

// Location is a representation of a Backstage Location.
type Location struct {
	APIVersion string       `yaml:"apiVersion"`
	Kind       string       `yaml:"kind"`
	Metadata   metadata     `yaml:"metadata"`
	Spec       locationSpec `yaml:"spec,omitempty"`
}

type metadata struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

type locationSpec struct {
	Targets []string `yaml:"targets,omitempty"`
}
