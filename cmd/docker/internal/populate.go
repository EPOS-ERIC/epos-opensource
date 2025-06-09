package internal

import (
	"epos-cli/common"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

func Populate(path, name, ttlDir string) error {
	common.PrintStep("Populating environment: %s", name)
	dir, err := GetEnvDir(path, name)
	if err != nil {
		return fmt.Errorf("failed to get environment directory: %w", err)
	}

	common.PrintDone("Environment found in dir: %s", dir)
	common.PrintStep("Deploying metadata-cache")

	err = deployMetadataCache(ttlDir, name)
	if err != nil {
		return fmt.Errorf("failed to deploy metadata-cache: %w", err)
	}

	env, err := godotenv.Read(filepath.Join(path, ".env"))
	if err != nil {
		return fmt.Errorf("error reading .env file: %w", err)
	}
	// check that all the vars we need are set
	if _, ok := env["GATEWAY_PORT"]; !ok {
		return fmt.Errorf("environment variable GATEWAY_PORT is not set")
	}
	if _, ok := env["DEPLOY_PATH"]; !ok {
		return fmt.Errorf("environment variable DEPLOY_PATH is not set")
	}
	if _, ok := env["API_PATH"]; !ok {
		return fmt.Errorf("environment variable API_PATH is not set")
	}
	posturl, err := url.Parse("http://localhost:" + env["GATEWAY_PORT"] + env["DEPLOY_PATH"] + env["API_PATH"] + "/populate")
	if err != nil {
		return fmt.Errorf("error building post url: %w", err)
	}

	freePort, err := common.GetFreePort()
	if err != nil {
		return fmt.Errorf("error getting free port: %w", err)
	}

	common.PrintDone("Deployed metadata-cache with mounted dir: %s", ttlDir)
	common.PrintStep("Starting the ingestion of the '*.ttl' files")

	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			common.PrintError("error while walkdir TODO: %v", err)
			return nil
		}

		if strings.HasSuffix(d.Name(), ".ttl") {
			common.PrintStep("ingesting file: %s", d.Name())
		}

		q := posturl.Query()
		q.Set("path", "http://"+os.Getenv("LOCAL_IP")+":"+strconv.Itoa(freePort)+"/"+d.Name())
		q.Set("securityCode", "changeme") // TODO: remove this in the ingestor
		q.Set("type", "single")
		q.Set("model", "EPOS-DCAT-AP-V1")
		q.Set("mapping", "EDM-TO-DCAT-AP")
		posturl.RawQuery = q.Encode()
		r, err := http.NewRequest("POST", posturl.String(), nil)
		if err != nil {
			common.PrintError("error building request for file '%s': %v", d.Name(), err)
			return err
		}
		r.Header.Add("accept", "*/*")

		res, err := http.DefaultClient.Do(r)
		if err != nil {
			common.PrintError("error ingesting file '%s' in database: %v", d.Name(), err)
			return err
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			common.PrintError("error reading response body of request: %v", err)
			return err
		}
		if res.StatusCode != 200 {
			common.PrintError("error ingesting file '%s' in database: received status code %d. Body of response: %s", d.Name(), res.StatusCode, string(body))
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
