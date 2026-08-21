package openapi

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"go.yaml.in/yaml/v2"
)

type document struct {
	OpenAPI    string              `yaml:"openapi"`
	Paths      map[string]pathItem `yaml:"paths"`
	Components componentSection    `yaml:"components"`
}

type pathItem map[string]any

type componentSection struct {
	Parameters map[string]any `yaml:"parameters"`
	Responses  map[string]any `yaml:"responses"`
	Schemas    map[string]any `yaml:"schemas"`
}

func loadOpenAPI(t *testing.T) document {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test file location")
	}

	root := filepath.Dir(filename)

	// Walk upward until docs/openapi.yaml is found.
	for {
		path := filepath.Join(root, "docs", "openapi.yaml")

		if _, err := os.Stat(path); err == nil {
			contents, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read OpenAPI document %q: %v", path, err)
			}

			var spec document

			if err := yaml.Unmarshal(contents, &spec); err != nil {
				t.Fatalf("parse OpenAPI document %q: %v", path, err)
			}

			return spec
		}

		parent := filepath.Dir(root)
		if parent == root {
			t.Fatal("could not find repository root containing docs/openapi.yaml")
		}

		root = parent
	}
}

func TestOpenAPIFoundation(t *testing.T) {
	spec := loadOpenAPI(t)

	if spec.OpenAPI != "3.1.0" {
		t.Fatalf(
			"expected OpenAPI 3.1.0, got %q",
			spec.OpenAPI,
		)
	}

	if len(spec.Paths) == 0 {
		t.Fatal("expected OpenAPI paths to be defined")
	}
}

func TestOpenAPIRouteCoverage(t *testing.T) {
	spec := loadOpenAPI(t)

	expected := map[string][]string{
		"/health":                  {"get"},
		"/ready":                   {"get"},
		"/version":                 {"get"},
		"/status":                  {"get"},
		"/metrics":                 {"get"},
		"/api/v1/communities":      {"get", "post"},
		"/api/v1/communities/{id}": {"get", "put", "delete"},
	}

	for route, methods := range expected {
		item, ok := spec.Paths[route]
		if !ok {
			t.Fatalf("OpenAPI missing route %q", route)
		}

		for _, method := range methods {
			if _, ok := item[method]; !ok {
				t.Fatalf(
					"OpenAPI missing %s %s",
					method,
					route,
				)
			}
		}
	}
}

func TestOpenAPIRequiredSchemas(t *testing.T) {
	spec := loadOpenAPI(t)

	expected := []string{
		"Community",
		"CommunityCreateRequest",
		"CommunityUpdateRequest",
		"CommunityListResponse",
		"ErrorResponse",
		"ErrorBody",
		"HealthResponse",
		"HealthCheck",
		"VersionResponse",
		"StatusResponse",
	}

	for _, name := range expected {
		if _, ok := spec.Components.Schemas[name]; !ok {
			t.Fatalf(
				"OpenAPI missing schema %q",
				name,
			)
		}
	}
}

func TestOpenAPIRequiredErrorResponses(t *testing.T) {
	spec := loadOpenAPI(t)

	expected := []string{
		"InvalidCommunity",
		"InvalidID",
		"CommunityNotFound",
		"CommunityAlreadyExists",
		"InternalServerError",
		"GatewayTimeout",
		"ServiceUnavailable",
	}

	for _, name := range expected {
		if _, ok := spec.Components.Responses[name]; !ok {
			t.Fatalf(
				"OpenAPI missing response component %q",
				name,
			)
		}
	}
}

func TestOpenAPICommunityIDParameter(t *testing.T) {
	spec := loadOpenAPI(t)

	if _, ok := spec.Components.Parameters["CommunityID"]; !ok {
		t.Fatal("OpenAPI missing CommunityID parameter")
	}
}
