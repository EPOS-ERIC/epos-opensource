package internal

import (
	"epos-cli/common"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

// composeCommand creates a docker compose command configured with the given directory and environment name
func composeCommand(dir, name string, args ...string) *exec.Cmd {
	cmd := exec.Command("docker", append([]string{"compose"}, args...)...)
	cmd.Dir = dir
	if name != "" {
		cmd.Env = append(cmd.Env, "ENV_NAME="+name)
	}
	return cmd
}

// pullEnvImages pulls docker images for the environment with custom messages
func pullEnvImages(dir, name string) error {
	common.PrintStep("Pulling images for environment %s", name)
	if err := common.RunCommand(composeCommand(dir, "", "pull")); err != nil {
		return fmt.Errorf("pull images failed: %w", err)
	}
	common.PrintDone("Images pulled for environment %s", name)
	return nil
}

// deployStack deploys the stack in the specified directory
func deployStack(dir, name string) error {
	common.PrintStep("Deploying stack")
	if err := common.RunCommand(composeCommand(dir, name, "up", "-d")); err != nil {
		return fmt.Errorf("deploy stack failed: %w", err)
	}
	common.PrintDone("Deployed environment: %s", name)
	return nil
}

// downStack stops the stack running in the given directory
func downStack(dir string, removeVolumes bool) error {
	if removeVolumes {
		return common.RunCommand(composeCommand(dir, "", "down", "-v"))
	}
	return common.RunCommand(composeCommand(dir, "", "down"))
}

// removeEnvDir deletes the environment directory with logs
func removeEnvDir(dir string) error {
	common.PrintStep("Deleting environment directory: %s", dir)
	if err := DeleteEnvDir(dir); err != nil {
		return err
	}
	common.PrintDone("Deleted environment directory: %s", dir)
	return nil
}

// deployMetadataCache deploys an nginx docker container running a file server exposing a volume
func deployMetadataCache(dir, envName string) (int, error) {
	port, err := common.GetFreePort()
	if err != nil {
		return 0, fmt.Errorf("error getting a free port for the metadata-cache: %w", err)
	}
	cmd := exec.Command(
		"docker",
		"run",
		"-d",
		"--name",
		envName+"-metadata-cache",
		"-p",
		strconv.Itoa(port)+":80",
		"-v",
		dir+":/usr/share/nginx/html",
		"nginx",
	)

	err = common.RunCommand(cmd)
	if err != nil {
		return 0, fmt.Errorf("error deploying metadata-cache: %w", err)
	}

	return port, nil
}

// deployMetadataCache removes a deployment of a metadata cache container
func deleteMetadataCache(envName string) error {
	cmd := exec.Command(
		"docker",
		"rm",
		"-f",
		envName+"-metadata-cache",
	)

	if err := common.RunCommand(cmd); err != nil {
		return fmt.Errorf("error removing metadata-cache: %w", err)
	}

	return nil
}

// createTmpCopy creates a backup copy of the environment directory in a temporary location
func createTmpCopy(dir string) (string, error) {
	common.PrintStep("Creating backup copy of environment")

	tmpDir, err := os.MkdirTemp("", "env-backup-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %w", err)
	}

	if err := copyDir(dir, tmpDir); err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("failed to copy environment to backup: %w", err)
	}

	common.PrintDone("Backup created at: %s", tmpDir)
	return tmpDir, nil
}

// restoreTmpDir restores the environment from temporary backup to target directory
func restoreTmpDir(tmpDir, targetDir string) error {
	common.PrintStep("Restoring environment from backup")

	if err := os.MkdirAll(targetDir, 0700); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	if err := copyDir(tmpDir, targetDir); err != nil {
		return fmt.Errorf("failed to restore from backup: %w", err)
	}

	common.PrintDone("Environment restored from backup")
	return nil
}

// copyDir recursively copies a directory from src to dst
func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source directory %s: %w", src, err)
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to create destination directory %s: %w", dst, err)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read source directory %s: %w", src, err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return fmt.Errorf("failed to copy subdirectory %s to %s: %w", srcPath, dstPath, err)
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return fmt.Errorf("failed to copy file %s to %s: %w", srcPath, dstPath, err)
			}
		}
	}

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", src, err)
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source file %s: %w", src, err)
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dst, err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy data from %s to %s: %w", src, dst, err)
	}

	return nil
}

// removeTmpDir removes the temporary backup directory with logs
func removeTmpDir(tmpDir string) error {
	common.PrintStep("Cleaning up backup directory: %s", tmpDir)

	if err := os.RemoveAll(tmpDir); err != nil {
		return fmt.Errorf("failed to cleanup backup directory: %w", err)
	}

	common.PrintDone("Backup directory cleaned up: %s", tmpDir)
	return nil
}

func getApiURL(dir string) (*url.URL, error) {
	env, err := godotenv.Read(filepath.Join(dir, ".env"))
	if err != nil {
		return nil, fmt.Errorf("failed to read .env file at %s: %w", filepath.Join(dir, ".env"), err)
	}
	// check that all the vars we need are set
	if _, ok := env["GATEWAY_PORT"]; !ok {
		return nil, fmt.Errorf("environment variable GATEWAY_PORT is not set")
	}
	if _, ok := env["DEPLOY_PATH"]; !ok {
		return nil, fmt.Errorf("environment variable DEPLOY_PATH is not set")
	}
	if _, ok := env["API_PATH"]; !ok {
		return nil, fmt.Errorf("environment variable API_PATH is not set")
	}
	postPath, err := url.JoinPath("http://localhost:"+env["GATEWAY_PORT"], env["DEPLOY_PATH"], env["API_PATH"])
	if err != nil {
		return nil, fmt.Errorf("error building post url: %w", err)
	}
	posturl, err := url.Parse(postPath)
	if err != nil {
		return nil, fmt.Errorf("error building post url: %w", err)
	}
	return posturl, nil
}

type ontology struct {
	path    string
	name    string
	ontType string
}

// populateOntologies populates an environment deployed in a `dir` with the base ontologies for the ingestor
func populateOntologies(dir string) error {
	// curl -X POST --header 'accept: */*' 'http://localhost:33000/api/v1/ontology?path=https://raw.githubusercontent.com/epos-eu/EPOS-DCAT-AP/EPOS-DCAT-AP-shapes/epos-dcat-ap_shapes.ttl&securityCode=changeme&name=EPOS-DCAT-AP-V1&type=BASE'
	// curl -X POST --header 'accept: */*' 'http://localhost:33000/api/v1/ontology?path=https://raw.githubusercontent.com/epos-eu/EPOS-DCAT-AP/EPOS-DCAT-AP-v3.0/docs/epos-dcat-ap_v3.0.0_shacl.ttl&securityCode=changeme&name=EPOS-DCAT-AP-V3&type=BASE'
	// curl -X POST --header 'accept: */*' 'http://localhost:33000/api/v1/ontology?path=https://raw.githubusercontent.com/epos-eu/EPOS_Data_Model_Mapping/main/edm-schema-shapes.ttl&securityCode=changeme&name=EDM-TO-DCAT-AP&type=MAPPING'

	baseURL, err := getApiURL(dir)
	if err != nil {
		return fmt.Errorf("error getting base api URL for environment in dir %s: %w", dir, err)
	}
	baseURL = baseURL.JoinPath("/ontology")

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

	common.PrintStep("Populating the environment with base ontologies")

	for i, ont := range ontologies {
		reqURL := *baseURL
		q := reqURL.Query()
		q.Set("path", ont.path)
		q.Set("securityCode", "changeme") // TODO: remove this once it's removed from the ingestor
		q.Set("type", ont.ontType)
		q.Set("name", ont.name)
		reqURL.RawQuery = q.Encode()

		common.PrintStep("  [%d/3] Loading %s ontology...", i+1, ont.name)

		req, err := http.NewRequest("POST", reqURL.String(), nil)
		if err != nil {
			return fmt.Errorf("error creating ontology request for %s: %w", ont.name, err)
		}
		req.Header.Set("accept", "*/*")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("error making ontology request for %s: %w", ont.name, err)
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != 200 {
			return fmt.Errorf("ontology request for %s failed with status %d: %s", ont.name, resp.StatusCode, string(body))
		}
	}

	common.PrintDone("All ontologies loaded successfully")

	return nil
}

func buildEnvURLs(dir string) (portalURL, gatewayURL string, err error) {
	env, err := godotenv.Read(filepath.Join(dir, ".env"))
	if err != nil {
		return "", "", fmt.Errorf("failed to read .env file at %s: %w", filepath.Join(dir, ".env"), err)
	}
	if _, ok := env["DATAPORTAL_PORT"]; !ok {
		return "", "", fmt.Errorf("environment variable DATAPORTAL_PORT is not set")
	}
	if _, ok := env["GATEWAY_PORT"]; !ok {
		return "", "", fmt.Errorf("environment variable GATEWAY_PORT is not set")
	}
	if _, ok := env["DEPLOY_PATH"]; !ok {
		return "", "", fmt.Errorf("environment variable DEPLOY_PATH is not set")
	}
	if _, ok := env["API_PATH"]; !ok {
		return "", "", fmt.Errorf("environment variable API_PATH is not set")
	}

	localIP, err := common.GetLocalIP()
	if err != nil {
		return "", "", fmt.Errorf("error getting local IP address: %w", err)
	}

	portalURL = "http://" + localIP + ":" + env["DATAPORTAL_PORT"]
	gatewayURL, err = url.JoinPath("http://"+localIP+":"+env["GATEWAY_PORT"], env["DEPLOY_PATH"], env["API_PATH"], "ui")
	if err != nil {
		return "", "", fmt.Errorf("error building path for gateway url: %w", err)
	}
	return portalURL, gatewayURL, nil
}
