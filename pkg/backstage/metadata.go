package backstage

// BackstageMetadata is a struct that contains Backstage-specific metadata.
type BackstageMetadata struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
	Tags        []string          `yaml:"tags,omitempty"`
	Links       []Link            `yaml:"links,omitempty"`
}

// Link is a link for users to access some facet of data for a component.
type Link struct {
	URL   string `yaml:"url"`
	Title string `yaml:"title,omitempty"`
	Icon  string `yaml:"icon,omitempty"`
}
