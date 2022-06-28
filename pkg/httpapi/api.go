package httpapi

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/julienschmidt/httprouter"
	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/bigkevmcd/peanut-backstage/pkg/backstage"
)

// BackstageRouter is an HTTP API for generating Backstage data from appropriately
// annotated resources.
type BackstageRouter struct {
	*httprouter.Router
	logger logr.Logger
	client client.Client
}

// NewRouter creates and returns a new Backstage router ready for use.
func NewRouter(l logr.Logger, c client.Client) *BackstageRouter {
	api := &BackstageRouter{
		Router: httprouter.New(),
		logger: l,
		client: c,
	}
	api.HandlerFunc(http.MethodGet, "/backstage/catalog-info.yaml", api.handleCatalogInfo)
	api.HandlerFunc(http.MethodGet, "/backstage/component/:name/info.yaml", api.handleComponent)
	return api
}

func (a *BackstageRouter) handleComponent(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	name := params.ByName("name")
	a.logger.Info("querying component", "component", name, "path", r.URL.String())

	var deploymentList appsv1.DeploymentList
	if err := a.client.List(r.Context(), &deploymentList); err != nil {
		http.Error(w, "failed to load deployments", http.StatusInternalServerError)
		return
	}

	parser := backstage.NewComponentParser()
	if err := parser.Add(&deploymentList); err != nil {
		http.Error(w, "failed to parse deployments", http.StatusInternalServerError)
		return
	}

	for _, v := range parser.Components() {
		if v.Metadata.Name == name {
			marshalResponse(w, v)
			return
		}
	}
	http.NotFound(w, r)
}

func (a *BackstageRouter) handleCatalogInfo(w http.ResponseWriter, r *http.Request) {
	a.logger.Info("querying catalog-info.yaml")
	var deploymentList appsv1.DeploymentList
	if err := a.client.List(r.Context(), &deploymentList); err != nil {
		http.Error(w, "failed to load deployments", http.StatusInternalServerError)
		return
	}

	parser := backstage.NewComponentParser()
	if err := parser.Add(&deploymentList); err != nil {
		http.Error(w, "failed to parse deployments", http.StatusInternalServerError)
		return
	}

	targets := []string{}
	for _, v := range parser.Components() {
		targets = append(targets, fmt.Sprintf("./component/%s/info.yaml", v.Metadata.Name))
	}
	// TODO: How to configure name, description for catalog-info?
	marshalResponse(w, backstage.NewLocation("test-service", "just a test", targets...))
}

func marshalResponse(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/yaml")
	if err := yaml.NewEncoder(w).Encode(v); err != nil {
		log.Printf("failed to encode response: %s", err)
	}
}
