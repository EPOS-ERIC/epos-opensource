package common

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/EPOS-ERIC/epos-opensource/db"
	"github.com/EPOS-ERIC/epos-opensource/display"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"golang.org/x/sync/errgroup"
)

var ErrImageMissing = errors.New("image not found locally")

const imageUpdateCacheTTL = 12 * time.Hour

type NamedImage struct {
	Name string
	Ref  string
}

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

func ImageExistsLocally(ctx context.Context, imageRef string) (bool, error) {
	if imageRef == "" {
		return false, fmt.Errorf("invalid image reference: %q", imageRef)
	}

	cmd := exec.CommandContext(ctx, "docker", "image", "inspect", imageRef)
	if err := cmd.Run(); err != nil {
		if _, ok := errors.AsType[*exec.ExitError](err); ok {
			return false, nil
		}

		return false, fmt.Errorf("failed to inspect local image %q: %w", imageRef, err)
	}

	return true, nil
}

func localImageDigest(ctx context.Context, imageRef string) (string, error) {
	if imageRef == "" {
		return "", fmt.Errorf("invalid image reference: %q", imageRef)
	}

	cmd := exec.CommandContext(ctx, "docker", "inspect", "--format={{index .RepoDigests 0}}", imageRef)
	output, err := cmd.Output()
	if err != nil {
		if _, ok := errors.AsType[*exec.ExitError](err); ok {
			return "", ErrImageMissing
		}

		return "", fmt.Errorf("failed to inspect local image %q: %w", imageRef, err)
	}

	digest := strings.TrimSpace(string(output))
	if digest == "" || digest == "<no value>" {
		return "", fmt.Errorf("no digest found for image %q", imageRef)
	}

	return digest, nil
}

func imageHasUpdate(ctx context.Context, imageRef string) (bool, *time.Time, error) {
	if imageRef == "" {
		return false, nil, fmt.Errorf("invalid image reference: %q", imageRef)
	}

	digest, err := localImageDigest(ctx, imageRef)
	if err != nil {
		return false, nil, err
	}

	parts := strings.Split(digest, "@")
	if len(parts) != 2 {
		return false, nil, fmt.Errorf("invalid digest format")
	}

	localDigest := parts[1]

	cached, err := db.GetImageUpdateCache(ctx, imageRef)
	if err == nil && cached != nil && cached.FetchedAt != nil && time.Since(*cached.FetchedAt) < imageUpdateCacheTTL {
		if localDigest == cached.RemoteDigest {
			return false, nil, nil
		}
	}

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
		_ = db.UpsertImageUpdateCache(ctx, imageRef, remoteDigest, nil, time.Now())
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

	_ = db.UpsertImageUpdateCache(ctx, imageRef, remoteDigest, &cf.Created.Time, time.Now())

	return true, &cf.Created.Time, nil
}

func checkImageForUpdate(ctx context.Context, image NamedImage) (*display.ImageUpdateInfo, error) {
	hasUpdate, lastUpdate, err := imageHasUpdate(ctx, image.Ref)
	if err != nil {
		return nil, err
	}
	if !hasUpdate {
		return nil, nil
	}

	return &display.ImageUpdateInfo{Name: image.Name, LastUpdate: *lastUpdate}, nil
}

func CheckImagesForUpdates(images []NamedImage) ([]display.ImageUpdateInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	updates := []display.ImageUpdateInfo{}
	if len(images) == 0 {
		return updates, nil
	}

	g, ctx := errgroup.WithContext(ctx)
	var mu sync.Mutex

	g.SetLimit(len(images))

	for _, image := range images {
		g.Go(func() error {
			update, err := checkImageForUpdate(ctx, image)
			if err != nil {
				display.Debug("skipping image update check for %s (%s): %v", image.Name, image.Ref, err)
				return nil
			}
			if update != nil {
				mu.Lock()
				updates = append(updates, *update)
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
