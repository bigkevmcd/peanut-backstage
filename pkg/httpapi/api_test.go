package httpapi

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/zapr"
	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/bigkevmcd/peanut-backstage/pkg/backstage"
	"github.com/bigkevmcd/peanut-backstage/test"
)

const (
	instanceLabel  = "app.kubernetes.io/instance"
	nameLabel      = "app.kubernetes.io/name"
	componentLabel = "app.kubernetes.io/component"
	createdByLabel = "app.kubernetes.io/created-by"
)

func TestGetRootLocation(t *testing.T) {
	dep := test.NewDeployment("test", "test-ns",
		test.WithLabels(map[string]string{
			instanceLabel:  "mysql-staging",
			nameLabel:      "mysql",
			componentLabel: "database",
			createdByLabel: "test-team",
		}),
	)

	ts := newTestServer(t, newFakeClient(t, &dep))
	req := makeClientRequest(t, ts, "/backstage/catalog-info.yaml")
	res, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}

	assertYAMLResponse(t, res, map[string]interface{}{
		"apiVersion": "backstage.io/v1alpha1",
		"kind":       "Location",
		"metadata": map[string]interface{}{
			"name":        "test-service",
			"description": "just a test",
		},
		"spec": map[string]interface{}{
			"targets": []any{
				"./component/mysql/info.yaml",
			},
		},
	})
}

func TestGetComponent(t *testing.T) {
	dep := test.NewDeployment("test", "test-ns",
		test.WithLabels(map[string]string{
			instanceLabel:            "mysql-staging",
			nameLabel:                "mysql",
			componentLabel:           "database",
			createdByLabel:           "test-team",
			backstage.LifecycleLabel: "production",
		}),
	)

	ts := newTestServer(t, newFakeClient(t, &dep))
	req := makeClientRequest(t, ts, "/backstage/component/mysql/info.yaml")
	res, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}

	assertYAMLResponse(t, res, map[string]interface{}{
		"apiVersion": "backstage.io/v1alpha1",
		"kind":       "Component",
		"metadata": map[string]interface{}{
			"name": "mysql",
		},
		"spec": map[string]interface{}{
			"lifecycle": "production",
			"owner":     "test-team",
			"type":      "database",
			"system":    "",
		},
	})
}

func newTestServer(t *testing.T, c client.Client) *httptest.Server {
	router := NewRouter(zapr.NewLogger(zap.NewNop()), c)
	ts := httptest.NewTLSServer(router)
	t.Cleanup(ts.Close)
	return ts
}

func makeClientRequest(t *testing.T, ts *httptest.Server, path string, opts ...func(*http.Request)) *http.Request {
	t.Helper()
	r, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", ts.URL, path), nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

func assertYAMLResponse(t *testing.T, res *http.Response, want map[string]interface{}) {
	t.Helper()
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("didn't get a successful response: %v (%q)", res.StatusCode, bytes.TrimSpace(b))
	}

	if h := res.Header.Get("Content-Type"); h != "application/yaml" {
		t.Fatalf("wanted 'application/yaml' got %s", h)
	}
	got := map[string]interface{}{}
	err = yaml.Unmarshal(b, &got)
	if err != nil {
		t.Fatalf("failed to parse %s: %s", b, err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("YAML response failed:\n%s", diff)
	}
}

func newFakeClient(t *testing.T, objs ...runtime.Object) client.Client {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := appsv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(objs...).
		Build()
}
