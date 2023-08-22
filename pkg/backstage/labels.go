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

// Unofficial annotations.
const (
	// LifecycleAnnotation is used to populate spec.lifecycle for Components.
	LifecycleAnnotation = "backstage.io/kubernetes-lifecycle"
	// DescriptionAnnotation is used to populate the medata.description for
	// Components.
	DescriptionAnnotation = "backstage.io/kubernetes-description"
	ownerAnnotation       = "backstage.io/kubernetes-owner"
	systemAnnotation      = "backstage.io/kubernetes-system"
	tagsAnnotation        = "backstage.io/kubernetes-tags"

	urlAnnotationPrefix = "backstage.gitops.pro/link-"
)

var parsedAnnotations = []string{
	LifecycleAnnotation,
	DescriptionAnnotation,
	ownerAnnotation,
	systemAnnotation,
	tagsAnnotation,
}
