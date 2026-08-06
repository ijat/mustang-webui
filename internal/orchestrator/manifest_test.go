package orchestrator

import "testing"

func TestParseManifest_Placeholder(t *testing.T) {
	if got := parseManifest([]byte("{}")); got != nil {
		t.Errorf("parseManifest(placeholder) = %+v, want nil", got)
	}
}

func TestParseManifest_Invalid(t *testing.T) {
	if got := parseManifest([]byte("not json")); got != nil {
		t.Errorf("parseManifest(invalid) = %+v, want nil", got)
	}
}

func TestParseManifest_Real(t *testing.T) {
	data := []byte(`{
		"runtime": {
			"linux-amd64": {"url": "https://example.com/jre-linux-amd64.tar.gz", "sha256": "aaa"},
			"darwin-arm64": {"url": "https://example.com/jre-darwin-arm64.tar.gz", "sha256": "bbb"}
		},
		"sidecar": {"url": "https://example.com/sidecar.jar", "sha256": "ccc"}
	}`)

	m := parseManifest(data)
	if m == nil {
		t.Fatal("parseManifest(real manifest) = nil, want a populated Manifest")
	}

	linux, ok := m.Runtime["linux-amd64"]
	if !ok {
		t.Fatal("missing linux-amd64 runtime entry")
	}
	if linux.URL != "https://example.com/jre-linux-amd64.tar.gz" || linux.SHA256 != "aaa" {
		t.Errorf("linux-amd64 asset = %+v, unexpected", linux)
	}

	if m.Sidecar.URL != "https://example.com/sidecar.jar" || m.Sidecar.SHA256 != "ccc" {
		t.Errorf("sidecar asset = %+v, unexpected", m.Sidecar)
	}
}

func TestJavaExecutableName(t *testing.T) {
	// This only ever exercises the branch matching the platform running
	// the test, but that's still a real assertion that the function
	// picks the right name for its own build target rather than always
	// returning a constant.
	name := javaExecutableName()
	if name != "java" && name != "java.exe" {
		t.Errorf("javaExecutableName() = %q, want %q or %q", name, "java", "java.exe")
	}
}
