// Package httpapi provides an HTTP API for serving Backstage catalog resources
// dynamically generated from Kubernetes deployments. It exposes endpoints for
// retrieving catalog locations and individual component definitions.
package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/bigkevmcd/peanut-backstage/pkg/backstage"
)

// BackstageRouter is an HTTP API for generating Backstage data from appropriately
// annotated resources.
type BackstageRouter struct {
	*http.ServeMux
	logger logr.Logger
	client client.Client
}

const (
	// Default timeout for Kubernetes API calls
	kubernetesTimeout = 10 * time.Second
)

// NewRouter creates and returns a new Backstage router ready for use.
func NewRouter(l logr.Logger, c client.Client) *BackstageRouter {
	api := &BackstageRouter{
		ServeMux: http.NewServeMux(),
		logger:   l,
		client:   c,
	}
	api.HandleFunc("/backstage/catalog-info.yaml", api.loggingMiddleware(api.handleCatalogInfo))
	api.HandleFunc("/backstage/component/", api.loggingMiddleware(api.handleComponent))
	return api
}

// loggingMiddleware wraps HTTP handlers with request/response logging
func (a *BackstageRouter) loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		a.logger.V(1).Info("incoming request",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
		)
		
		next(w, r)
		
		a.logger.V(1).Info("request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}
}

// getDeployments fetches deployments from Kubernetes with timeout
func (a *BackstageRouter) getDeployments(ctx context.Context) (*appsv1.DeploymentList, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, kubernetesTimeout)
	defer cancel()

	var deploymentList appsv1.DeploymentList
	if err := a.client.List(timeoutCtx, &deploymentList); err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}
	
	a.logger.V(1).Info("fetched deployments", "count", len(deploymentList.Items))
	return &deploymentList, nil
}

// parseComponents parses components from a deployment list
func (a *BackstageRouter) parseComponents(deploymentList *appsv1.DeploymentList) ([]backstage.Component, error) {
	parser := backstage.NewComponentParser()
	if err := parser.Add(deploymentList); err != nil {
		return nil, fmt.Errorf("failed to parse deployments: %w", err)
	}
	
	components := parser.Components()
	a.logger.V(1).Info("parsed components", "count", len(components))
	return components, nil
}

func (a *BackstageRouter) handleComponent(w http.ResponseWriter, r *http.Request) {
	// Extract component name from URL path
	// Path format: /backstage/component/{name}/info.yaml
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/backstage/component/"), "/")
	if len(pathParts) < 2 {
		a.logger.Info("invalid component path", "path", r.URL.Path)
		http.Error(w, "invalid component path", http.StatusBadRequest)
		return
	}
	
	name := pathParts[0]
	
	// Validate component name to prevent path traversal
	if name == "" || name == "." || name == ".." || strings.Contains(name, "/") {
		a.logger.Info("invalid component name requested", "name", name)
		http.Error(w, "invalid component name", http.StatusBadRequest)
		return
	}
	
	a.logger.Info("querying component", "component", name, "path", r.URL.String())

	deploymentList, err := a.getDeployments(r.Context())
	if err != nil {
		a.logger.Error(err, "failed to load deployments")
		http.Error(w, "failed to load deployments", http.StatusInternalServerError)
		return
	}

	components, err := a.parseComponents(deploymentList)
	if err != nil {
		a.logger.Error(err, "failed to parse deployments")
		http.Error(w, "failed to parse deployments", http.StatusInternalServerError)
		return
	}

	for _, v := range components {
		if v.Metadata.Name == name {
			a.logger.Info("component found", "component", name)
			a.marshalResponse(w, v)
			return
		}
	}
	
	a.logger.Info("component not found", "component", name)
	http.NotFound(w, r)
}

func (a *BackstageRouter) handleCatalogInfo(w http.ResponseWriter, r *http.Request) {
	a.logger.Info("querying catalog-info.yaml")
	
	deploymentList, err := a.getDeployments(r.Context())
	if err != nil {
		a.logger.Error(err, "failed to load deployments")
		http.Error(w, "failed to load deployments", http.StatusInternalServerError)
		return
	}

	components, err := a.parseComponents(deploymentList)
	if err != nil {
		a.logger.Error(err, "failed to parse deployments")
		http.Error(w, "failed to parse deployments", http.StatusInternalServerError)
		return
	}

	targets := []string{}
	for _, v := range components {
		targets = append(targets, fmt.Sprintf("./component/%s/info.yaml", v.Metadata.Name))
	}
	
	a.logger.Info("catalog generated", "component_count", len(targets))
	// TODO: How to configure name, description for catalog-info?
	a.marshalResponse(w, backstage.NewLocation("test-service", "just a test", targets...))
}

// marshalResponse marshals a value to YAML and writes it to the response
func (a *BackstageRouter) marshalResponse(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/yaml")
	if err := yaml.NewEncoder(w).Encode(v); err != nil {
		a.logger.Error(err, "failed to encode response")
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
