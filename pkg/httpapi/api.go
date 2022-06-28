package httpapi

import (
	"log"
	"net/http"

	"gopkg.in/yaml.v3"

	"github.com/go-logr/logr"
	"github.com/julienschmidt/httprouter"
)

// BackstageRouter is an HTTP API for generating Backstage data from appropriately
// annotated resources.
type BackstageRouter struct {
	*httprouter.Router
	logger logr.Logger
}

// NewRouter creates and returns a new Backstage router ready for use.
func NewRouter(l logr.Logger) *BackstageRouter {
	api := &BackstageRouter{
		Router: httprouter.New(),
		logger: l,
	}
	api.HandlerFunc(http.MethodGet, "/backstage/:name/catalog-info.yaml", api.handleCatalogInfo)
	return api
}

func (a *BackstageRouter) handleCatalogInfo(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	name := params.ByName("name")

	marshalResponse(w, newLocation(name, "just a test"))
}

func marshalResponse(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/yaml")
	if err := yaml.NewEncoder(w).Encode(v); err != nil {
		log.Printf("failed to encode response: %s", err)
	}
}
