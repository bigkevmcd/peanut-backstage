package backstage

// BackstageMetadata is a struct that contains Backstage-specific metadata.
type BackstageMetadata struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Annotations map[string]string `yaml:"annotations"`
	Tags        []string          `yaml:"tags"`
	Links       []Link            `yaml:"links"`
}

// Link is a link for users to access some facet of data for a component.
type Link struct {
	URL   string `yaml:"uml"`
	Title string `yaml:"title,omitempty"`
	Icon  string `yaml:"icon,omitempty"`
}
