package common

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"golang.org/x/sync/errgroup"
)

var ErrImageMissing = errors.New("image not found locally")

type Images struct {
	RabbitmqImage           string `yaml:"rabbitmq_image"`
	DataportalImage         string `yaml:"dataportal_image"`
	GatewayImage            string `yaml:"gateway_image"`
	MetadataDatabaseImage   string `yaml:"metadata_database_image"`
	ResourcesServiceImage   string `yaml:"resources_service_image"`
	IngestorServiceImage    string `yaml:"ingestor_service_image"`
	ExternalAccessImage     string `yaml:"external_access_image"`
	ConverterServiceImage   string `yaml:"converter_service_image"`
	ConverterRoutineImage   string `yaml:"converter_routine_image"`
	BackofficeServiceImage  string `yaml:"backoffice_service_image"`
	BackofficeUIImage       string `yaml:"backoffice_ui_image"`
	EmailSenderServiceImage string `yaml:"email_sender_service_image"`
	SharingServiceImage     string `yaml:"sharing_service_image"`
}

func imageHasUpdate(ctx context.Context, imageRef string) (bool, *time.Time, error) {
	if imageRef == "" {
		return false, nil, fmt.Errorf("invalid image reference: %q", imageRef)
	}

	cmd := exec.CommandContext(ctx, "docker", "inspect", "--format={{index .RepoDigests 0}}", imageRef)
	output, err := cmd.Output()
	if err != nil {
		return false, nil, ErrImageMissing
	}

	digest := strings.TrimSpace(string(output))
	if digest == "" {
		return false, nil, fmt.Errorf("no digest found for image")
	}

	parts := strings.Split(digest, "@")
	if len(parts) != 2 {
		return false, nil, fmt.Errorf("invalid digest format")
	}

	localDigest := parts[1]

	ref, err := name.ParseReference(imageRef)
	if err != nil {
		return false, nil, fmt.Errorf("invalid image reference: %w", err)
	}

	remoteDescriptor, err := remote.Head(ref, remote.WithContext(ctx))
	if err != nil {
		return false, nil, fmt.Errorf("failed to fetch remote descriptor: %w", err)
	}

	remoteDigest := remoteDescriptor.Digest.String()

	hasUpdate := localDigest != remoteDigest
	if !hasUpdate {
		return false, nil, nil
	}

	img, err := remote.Image(ref, remote.WithContext(ctx))
	if err != nil {
		return true, nil, fmt.Errorf("failed to get remote image: %w", err)
	}

	cf, err := img.ConfigFile()
	if err != nil {
		return true, nil, fmt.Errorf("failed to get remote image config: %w", err)
	}

	return true, &cf.Created.Time, nil
}

func CheckEnvForUpdates(images map[string]string) ([]display.ImageUpdateInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	updates := []display.ImageUpdateInfo{}
	g, ctx := errgroup.WithContext(ctx)
	var mu sync.Mutex

	g.SetLimit(13)

	for varName, imageRef := range images {
		g.Go(func() error {
			hasUpdate, lastUpdate, err := imageHasUpdate(ctx, imageRef)
			if err != nil {
				return fmt.Errorf("failed to check image update for %s: %w", varName, err)
			}
			if hasUpdate {
				mu.Lock()
				updates = append(updates, display.ImageUpdateInfo{Name: varName, LastUpdate: *lastUpdate})
				mu.Unlock()
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return updates, err
	}

	return updates, nil
}
