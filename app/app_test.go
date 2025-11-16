package app

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/magneticstain/ip-2-cloudresource/resource"
)

func TestGetSupportedPlatforms(t *testing.T) {
	got := GetSupportedPlatforms()
	want := []string{"aws", "gcp", "azure"}
	if len(got) != len(want) {
		t.Fatalf("expected %v got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected %v got %v", want, got)
		}
	}
}

func TestOutputResults_JSON(t *testing.T) {
	// capture stdout
	rPipe, wPipe, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = wPipe
	defer func() { os.Stdout = old }()

	// Build a resource and expect JSON when silent+json
	r := resource.Resource{
		RID:            "r-123",
		AccountID:      "acc-1",
		AccountAliases: []string{"alias1"},
	}

	// call function
	OutputResults(r, false, true, true)

	// close writer and read output
	_ = wPipe.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, rPipe)
	out := strings.TrimSpace(buf.String())
	if out == "" {
		t.Fatalf("expected json output, got empty")
	}
	if !strings.Contains(out, "r-123") {
		t.Fatalf("expected RID in output; got %s", out)
	}
}
