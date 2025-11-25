package backstage

const (
	// KindLocation is the kind for Backstage locations.
	KindLocation = "Location"
)

// NewLocation creates and returns a prepopulated Location.
// A Location points to other catalog entity definitions through its targets.
func NewLocation(name, description string, targets ...string) *Location {
	return &Location{
		APIVersion: APIVersion,
		Kind:       KindLocation,
		Metadata: BackstageMetadata{
			Name:        name,
			Description: description,
		},
		Spec: LocationSpec{
			Targets: targets,
		},
	}
}

// Location is a representation of a Backstage Location.
// Locations are used to organize catalog entities by pointing to where
// component definitions can be found.
type Location struct {
	APIVersion string            `yaml:"apiVersion"`
	Kind       string            `yaml:"kind"`
	Metadata   BackstageMetadata `yaml:"metadata"`
	Spec       LocationSpec      `yaml:"spec,omitempty"`
}

// LocationSpec is the spec for Location resources.
// It contains an array of target URLs pointing to catalog entity definitions.
type LocationSpec struct {
	Targets []string `yaml:"targets,omitempty"`
}
