package backstage

// BackstageMetadata is a struct that contains Backstage-specific metadata.
// This includes the entity name, description, custom annotations, tags, and links.
type BackstageMetadata struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
	Tags        []string          `yaml:"tags,omitempty"`
	Links       []Link            `yaml:"links,omitempty"`
}

// Link is a link for users to access some facet of data for a component.
// Links provide quick access to related resources like dashboards, documentation, or tools.
type Link struct {
	URL   string `yaml:"url"`
	Title string `yaml:"title,omitempty"`
	Icon  string `yaml:"icon,omitempty"`
}
