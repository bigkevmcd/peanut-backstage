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
)

func TestGetRouteLocation(t *testing.T) {
	ts := newTestServer(t)
	req := makeClientRequest(t, ts, "/backstage/test-service/catalog-info.yaml")
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
	})
}

func newTestServer(t *testing.T) *httptest.Server {
	router := NewRouter(zapr.NewLogger(zap.NewNop()))
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
