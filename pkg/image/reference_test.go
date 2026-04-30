package image

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		wantRef *Reference
	}{
		{
			name:    "empty image name",
			input:   "",
			wantErr: true,
		},
		{
			name:  "simple image name",
			input: "nginx",
			wantRef: &Reference{
				Registry: "docker.io",
				Name:     "library/nginx",
				Tag:      "latest",
			},
		},
		{
			name:  "image with tag",
			input: "nginx:1.21",
			wantRef: &Reference{
				Registry: "docker.io",
				Name:     "library/nginx",
				Tag:      "1.21",
			},
		},
		{
			name:  "image with namespace",
			input: "prom/prometheus",
			wantRef: &Reference{
				Registry: "docker.io",
				Name:     "prom/prometheus",
				Tag:      "latest",
			},
		},
		{
			name:  "image with namespace and tag",
			input: "prom/prometheus:v2.45.0",
			wantRef: &Reference{
				Registry: "docker.io",
				Name:     "prom/prometheus",
				Tag:      "v2.45.0",
			},
		},
		{
			name:  "image with custom registry",
			input: "ghcr.io/pldq/prom/prometheus",
			wantRef: &Reference{
				Registry: "ghcr.io",
				Name:     "pldq/prom/prometheus",
				Tag:      "latest",
			},
		},
		{
			name:  "image with custom registry and tag",
			input: "ghcr.io/pldq/prom/prometheus:v2.45.0",
			wantRef: &Reference{
				Registry: "ghcr.io",
				Name:     "pldq/prom/prometheus",
				Tag:      "v2.45.0",
			},
		},
		{
			name:  "localhost registry",
			input: "localhost:5000/myimage",
			wantRef: &Reference{
				Registry: "localhost:5000",
				Name:     "myimage",
				Tag:      "latest",
			},
		},
		{
			name:  "localhost registry with tag",
			input: "localhost:5000/myimage:latest",
			wantRef: &Reference{
				Registry: "localhost:5000",
				Name:     "myimage",
				Tag:      "latest",
			},
		},
		{
			name:  "registry with port",
			input: "registry.example.com:5000/myimage:v1",
			wantRef: &Reference{
				Registry: "registry.example.com:5000",
				Name:     "myimage",
				Tag:      "v1",
			},
		},
		{
			name:  "image with port in path no tag",
			input: "localhost:5000/namespace/image",
			wantRef: &Reference{
				Registry: "localhost:5000",
				Name:     "namespace/image",
				Tag:      "latest",
			},
		},
		{
			name:  "image with port in path and tag",
			input: "localhost:5000/namespace/image:v1",
			wantRef: &Reference{
				Registry: "localhost:5000",
				Name:     "namespace/image",
				Tag:      "v1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref, err := Parse(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Parse(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("Parse(%q) unexpected error: %v", tt.input, err)
				return
			}
			if ref.Registry != tt.wantRef.Registry {
				t.Errorf("Registry: got %q, want %q", ref.Registry, tt.wantRef.Registry)
			}
			if ref.Name != tt.wantRef.Name {
				t.Errorf("Name: got %q, want %q", ref.Name, tt.wantRef.Name)
			}
			if ref.Tag != tt.wantRef.Tag {
				t.Errorf("Tag: got %q, want %q", ref.Tag, tt.wantRef.Tag)
			}
		})
	}
}

func TestReference_String(t *testing.T) {
	ref := &Reference{
		Registry: "docker.io",
		Name:     "library/nginx",
		Tag:      "1.21",
	}
	want := "docker.io/library/nginx:1.21"
	if got := ref.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestReference_BranchName(t *testing.T) {
	ref := &Reference{
		Name: "prom/prometheus",
	}
	want := "prom/prometheus"
	if got := ref.BranchName(); got != want {
		t.Errorf("BranchName() = %q, want %q", got, want)
	}
}

func TestReference_GHCRName(t *testing.T) {
	ref := &Reference{
		Name: "prom/prometheus",
	}
	want := "ghcr.io/myuser/prom/prometheus"
	if got := ref.GHCRName("myuser"); got != want {
		t.Errorf("GHCRName() = %q, want %q", got, want)
	}
}

func TestReference_GHCRReference(t *testing.T) {
	ref := &Reference{
		Name: "prom/prometheus",
		Tag:  "v2.45.0",
	}
	want := "ghcr.io/myuser/prom/prometheus:v2.45.0"
	if got := ref.GHCRReference("myuser"); got != want {
		t.Errorf("GHCRReference() = %q, want %q", got, want)
	}
}

func TestReference_OriginalReference(t *testing.T) {
	tests := []struct {
		name string
		ref  *Reference
		want string
	}{
		{
			name: "library image",
			ref:  &Reference{Name: "library/nginx", Tag: "latest"},
			want: "nginx:latest",
		},
		{
			name: "non-library image",
			ref:  &Reference{Name: "prom/prometheus", Tag: "v2.45.0"},
			want: "prom/prometheus:v2.45.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ref.OriginalReference(); got != tt.want {
				t.Errorf("OriginalReference() = %q, want %q", got, tt.want)
			}
		})
	}
}
