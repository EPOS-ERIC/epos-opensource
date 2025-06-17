package common

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/joho/godotenv"
)

var configPath string

func init() {
	// Initialize the platform-specific config path
	initConfigPath()

	// Create the directory if it doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		err = os.MkdirAll(configPath, 0700)
		if err != nil {
			PrintError("failed to create config directory %s: %v", configPath, err)
			os.Exit(1)
		}
	}
}

// Helper function to create a file with given content
func CreateFileWithContent(filePath, content string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer file.Close()
	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("failed to write content to file %s: %w", filePath, err)
	}
	return nil
}

// GetContentFromPathOrDefault reads content from filePath if provided, otherwise returns defaultContent
func GetContentFromPathOrDefault(filePath, defaultContent string) (string, error) {
	if filePath == "" {
		return defaultContent, nil
	}
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	return string(content), nil
}

func DeleteEnvDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("directory %s does not exist", path)
	}
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to remove directory %s: %w", path, err)
	}
	return nil
}

// BuildEnvPath constructs the environment directory path
func BuildEnvPath(customPath, name, prefix string) string {
	var basePath string
	if customPath != "" {
		basePath = customPath
	} else {
		basePath = path.Join(configPath, prefix)
	}
	return path.Join(basePath, name)
}

// GetEnvDir validates that the full directory path exists and returns it
func GetEnvDir(customPath, name, prefix string) (string, error) {
	envPath := BuildEnvPath(customPath, name, prefix)
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		return "", fmt.Errorf("directory %s does not exist: %w", envPath, err)
	} else if err != nil {
		return "", fmt.Errorf("failed to check directory %s: %w", envPath, err)
	}
	return envPath, nil
}

// removeEnvDir deletes the environment directory with logs
func RemoveEnvDir(dir string) error {
	PrintStep("Deleting environment directory: %s", dir)
	if err := DeleteEnvDir(dir); err != nil {
		return err
	}
	PrintDone("Deleted environment directory: %s", dir)
	return nil
}

func BuildEnvURLs(dir string) (portalURL, gatewayURL string, err error) {
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

	localIP, err := GetLocalIP()
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

func GetApiURL(dir string) (*url.URL, error) {
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

	localIP, err := GetLocalIP()
	if err != nil {
		return nil, fmt.Errorf("error getting local ip: %w", err)
	}
	postPath, err := url.JoinPath(fmt.Sprintf("http://%s:%s", localIP, env["GATEWAY_PORT"]), env["DEPLOY_PATH"], env["API_PATH"])
	if err != nil {
		return nil, fmt.Errorf("error building post url: %w", err)
	}
	posturl, err := url.Parse(postPath)
	if err != nil {
		return nil, fmt.Errorf("error building post url: %w", err)
	}
	return posturl, nil
}
