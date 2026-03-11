package config_test

import (
	"sort"
	"strings"
	"testing"

	"github.com/EPOS-ERIC/epos-opensource/pkg/k8s/config"
)

func TestConfigRender_EmailSenderAuth(t *testing.T) {
	cfg := config.GetDefaultConfig()
	cfg.Name = "test-email-auth"
	cfg.Components.EmailSenderService.Enabled = true
	cfg.Components.EmailSenderService.Auth = config.Auth{
		Enabled:   true,
		OnlyAdmin: true,
	}

	files, err := cfg.Render()
	if err != nil {
		t.Fatalf("Render() returned error: %v", err)
	}

	var gatewayManifest string
	for path, content := range files {
		if strings.HasSuffix(path, "templates/gateway.yaml") {
			gatewayManifest = content
			break
		}
	}

	if gatewayManifest == "" {
		keys := make([]string, 0, len(files))
		for path := range files {
			keys = append(keys, path)
		}
		sort.Strings(keys)
		t.Fatalf("gateway manifest not found in rendered files: %v", keys)
	}

	if !strings.Contains(gatewayManifest, `LOAD_EMAIL_SENDER_API: "true:true"`) {
		t.Fatalf("gateway manifest does not contain expected email sender auth flag")
	}
}
