package common

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/EPOS-ERIC/epos-opensource/display"
)

type ontology struct {
	path    string
	name    string
	ontType string
}

// PopulateOntologies populates an environment deployed in a dir with the base ontologies for the ingestor
func PopulateOntologies(baseURL string) error {
	httpClient := &http.Client{
		Timeout: 1 * time.Minute,
	}

	apiURL, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("error parsing api url '%s': %w", baseURL, err)
	}
	apiURL = apiURL.JoinPath("ontology")

	ontologies := []ontology{
		{
			path:    "https://raw.githubusercontent.com/epos-eu/EPOS-DCAT-AP/EPOS-DCAT-AP-shapes/epos-dcat-ap_shapes.ttl",
			name:    "EPOS-DCAT-AP-V1",
			ontType: "BASE",
		},
		{
			path:    "https://raw.githubusercontent.com/epos-eu/EPOS-DCAT-AP/EPOS-DCAT-AP-v3.0/docs/epos-dcat-ap_v3.0.0_shacl.ttl",
			name:    "EPOS-DCAT-AP-V3",
			ontType: "BASE",
		},
		{
			path:    "https://raw.githubusercontent.com/epos-eu/EPOS_Data_Model_Mapping/main/edm-schema-shapes.ttl",
			name:    "EDM-TO-DCAT-AP",
			ontType: "MAPPING",
		},
	}

	display.Step("Populating the environment with base ontologies")

	for i, ont := range ontologies {
		reqURL := *apiURL
		q := reqURL.Query()
		q.Set("path", ont.path)
		q.Set("name", ont.name)
		q.Set("type", ont.ontType)
		q.Set("securityCode", "changeme") // TODO: remove this from the ingestor
		reqURL.RawQuery = q.Encode()

		display.Step("  [%d/3] Loading %s ontology...", i+1, ont.name)

		req, err := http.NewRequest("POST", reqURL.String(), nil)
		if err != nil {
			return fmt.Errorf("error creating ontology request for %s: %w", ont.name, err)
		}
		req.Header.Set("Accept", "*/*")

		resp, err := httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("error making ontology request for %s: %w", ont.name, err)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read body: %w", err)
		}
		err = resp.Body.Close()
		if err != nil {
			return fmt.Errorf("failed to close body: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("ontology request for %s failed with status %d: %s", ont.name, resp.StatusCode, string(body))
		}
	}
	display.Done("All ontologies loaded successfully")

	return nil
}
