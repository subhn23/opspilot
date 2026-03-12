package deploy

import (
	"context"
	"fmt"
	"io"
	"log"
	"opspilot/internal/audit"
	"time"

	"github.com/google/uuid"
	"github.com/moby/moby/client"
	"gorm.io/gorm"
)

// ImageService abstracts container image operations
type ImageService interface {
	ListImages(ctx context.Context, registryAddr string) ([]string, error)
	PullImage(ctx context.Context, image string) error
	PushImage(ctx context.Context, imageName, targetRegistry string) error
}

type RealImageService struct {
	Docker *client.Client
}

func (r *RealImageService) ListImages(ctx context.Context, registryAddr string) ([]string, error) {
	resp, err := r.Docker.ImageList(ctx, client.ImageListOptions{})
	if err != nil {
		return nil, err
	}

	var names []string
	for _, img := range resp.Items {
		for _, tag := range img.RepoTags {
			if fmt.Sprintf("%s/", registryAddr) == tag[:len(registryAddr)+1] {
				names = append(names, tag[len(registryAddr)+1:])
			}
		}
	}
	return names, nil
}

func (r *RealImageService) PullImage(ctx context.Context, imageStr string) error {
	resp, err := r.Docker.ImagePull(ctx, imageStr, client.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer resp.Close()
	// Drain reader
	_, _ = io.Copy(io.Discard, resp)
	return nil
}

func (r *RealImageService) PushImage(ctx context.Context, imageName, targetRegistry string) error {
	targetImage := fmt.Sprintf("%s/%s", targetRegistry, imageName)

	// Tag for target registry
	_, err := r.Docker.ImageTag(ctx, client.ImageTagOptions{
		Source: imageName,
		Target: targetImage,
	})
	if err != nil {
		return err
	}

	resp, err := r.Docker.ImagePush(ctx, targetImage, client.ImagePushOptions{})
	if err != nil {
		return err
	}
	defer resp.Close()
	// Drain reader
	_, _ = io.Copy(io.Discard, resp)
	return nil
}
type RegistrySync struct {
	DB           *gorm.DB
	ImageService ImageService
	Host1Addr    string
	Host2Addr    string
	Interval     time.Duration
}

func NewRegistrySync(db *gorm.DB, host1, host2 string, service ImageService) *RegistrySync {
	return &RegistrySync{
		DB:           db,
		Host1Addr:    host1,
		Host2Addr:    host2,
		ImageService: service,
		Interval:     60 * time.Second,
	}
}

// SyncNodes ensures all images in Host 1 registry are present in Host 2 registry
func (s *RegistrySync) SyncNodes(ctx context.Context) error {
	log.Printf("RegistrySync: Starting synchronization from %s to %s", s.Host1Addr, s.Host2Addr)

	images, err := s.ImageService.ListImages(ctx, s.Host1Addr)
	if err != nil {
		return fmt.Errorf("failed to list images from host1: %w", err)
	}

	for _, img := range images {
		log.Printf("RegistrySync: Syncing image %s", img)
		
		if err := s.ImageService.PullImage(ctx, s.Host1Addr+"/"+img); err != nil {
			log.Printf("RegistrySync: Failed to pull %s: %v", img, err)
			continue
		}

		if err := s.ImageService.PushImage(ctx, img, s.Host2Addr); err != nil {
			log.Printf("RegistrySync: Failed to push %s to host2: %v", img, err)
			continue
		}
	}

	audit.LogAction(s.DB, uuid.Nil, "REGISTRY_SYNC", s.Host1Addr, s.Host2Addr, fmt.Sprintf("Synced %d images", len(images)))
	log.Println("RegistrySync: Synchronization complete")
	
	return nil
}

// StartWorker runs the synchronization logic in a background loop
func (s *RegistrySync) StartWorker(ctx context.Context) {
	log.Println("RegistrySync: Background worker started")
	ticker := time.NewTicker(s.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("RegistrySync: Background worker stopping")
			return
		case <-ticker.C:
			if err := s.SyncNodes(ctx); err != nil {
				log.Printf("RegistrySync: Worker iteration failed: %v", err)
			}
		}
	}
}
