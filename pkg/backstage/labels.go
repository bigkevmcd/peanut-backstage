package backstage

const (
	// AppLabel is the Kubernetes recommended label to indicate that a component
	// is part of an application.
	AppLabel = "app.kubernetes.io/part-of"

	partOfLabel    = AppLabel
	instanceLabel  = "app.kubernetes.io/instance"
	nameLabel      = "app.kubernetes.io/name"
	componentLabel = "app.kubernetes.io/component"
	createdByLabel = "app.kubernetes.io/created-by"
)

// Unofficial labels.
const (
	// LifeCycle label provides the Backstage entity lifecycle stage.
	LifecycleLabel = "backstage.gitops.pro/lifecycle"
)

// Unofficial annotations.
const (
	tagsAnnotation        = "backstage.gitops.pro/tags"
	descriptionAnnotation = "backstage.gitops.pro/description"

	urlAnnotationPrefix = "backstage.gitops.pro/link-"
)
