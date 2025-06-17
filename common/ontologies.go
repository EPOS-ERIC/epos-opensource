package common

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

type ontology struct {
	path    string
	name    string
	ontType string
}

// populateOntologies populates an environment deployed in a `dir` with the base ontologies for the ingestor
func PopulateOntologies(dir string) error {
	// curl -X POST --header 'accept: */*' 'http://localhost:33000/api/v1/ontology?path=https://raw.githubusercontent.com/epos-eu/EPOS-DCAT-AP/EPOS-DCAT-AP-shapes/epos-dcat-ap_shapes.ttl&securityCode=changeme&name=EPOS-DCAT-AP-V1&type=BASE'
	// curl -X POST --header 'accept: */*' 'http://localhost:33000/api/v1/ontology?path=https://raw.githubusercontent.com/epos-eu/EPOS-DCAT-AP/EPOS-DCAT-AP-v3.0/docs/epos-dcat-ap_v3.0.0_shacl.ttl&securityCode=changeme&name=EPOS-DCAT-AP-V3&type=BASE'
	// curl -X POST --header 'accept: */*' 'http://localhost:33000/api/v1/ontology?path=https://raw.githubusercontent.com/epos-eu/EPOS_Data_Model_Mapping/main/edm-schema-shapes.ttl&securityCode=changeme&name=EDM-TO-DCAT-AP&type=MAPPING'

	httpClient := &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
		Timeout: 30 * time.Second,
	}

	baseURL, err := GetApiURL(dir)
	if err != nil {
		return fmt.Errorf("error getting base api URL for environment in dir %s: %w", dir, err)
	}
	baseURL = baseURL.JoinPath("ontology")

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

	PrintStep("Populating the environment with base ontologies")

	// HACK: implemented a retry attempt because sometimes (god knows why) the first couple of posts give errors. This seems to fix it
	const maxAttempts = 3
	const retryDelay = 500 * time.Millisecond

	for i, ont := range ontologies {
		reqURL := *baseURL
		q := reqURL.Query()
		q.Set("path", ont.path)
		q.Set("securityCode", "changeme")
		q.Set("type", ont.ontType)
		q.Set("name", ont.name)
		reqURL.RawQuery = q.Encode()

		PrintStep("  [%d/3] Loading %s ontology...", i+1, ont.name)

		var lastErr error
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			req, err := http.NewRequest("POST", reqURL.String(), nil)
			if err != nil {
				return fmt.Errorf("error creating ontology request for %s: %w", ont.name, err)
			}
			req.Header.Set("Accept", "*/*")
			req.Close = true

			resp, err := httpClient.Do(req)
			if err != nil {
				lastErr = fmt.Errorf("attempt %d: error making ontology request for %s: %w", attempt, ont.name, err)
			} else {
				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					lastErr = nil
					break
				}
				lastErr = fmt.Errorf("attempt %d: ontology request for %s failed with status %d: %s",
					attempt, ont.name, resp.StatusCode, string(body))
			}

			// if not last attempt, wait before retrying
			if attempt < maxAttempts {
				time.Sleep(retryDelay)
			}
		}

		if lastErr != nil {
			// all attempts failed
			return lastErr
		}
	}
	PrintDone("All ontologies loaded successfully")

	return nil
}
