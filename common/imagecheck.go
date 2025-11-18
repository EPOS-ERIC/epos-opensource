package common

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/epos-eu/epos-opensource/display"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/joho/godotenv"
	"golang.org/x/sync/errgroup"
)

var ErrImageMissing = errors.New("image not found locally")

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

func CheckEnvForUpdates(envFile string) ([]display.ImageUpdateInfo, error) {
	if envFile == "" {
		return nil, fmt.Errorf("invalid env file: %q", envFile)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	env, err := godotenv.Read(envFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read .env file at %s: %w", envFile, err)
	}

	updates := []display.ImageUpdateInfo{}
	g, ctx := errgroup.WithContext(ctx)
	var mu sync.Mutex

	g.SetLimit(13)

	for v, k := range env {
		if !strings.HasSuffix(v, "_IMAGE") {
			continue
		}

		varName, imageRef := v, k
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
