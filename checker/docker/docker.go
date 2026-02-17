package docker

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/mahype/update-watcher/checker"
	"github.com/mahype/update-watcher/config"
	"github.com/mahype/update-watcher/internal/executil"
)

// isImageID returns true if the image reference is a bare hex ID (not a name:tag).
var hexPattern = regexp.MustCompile(`^[0-9a-f]{12,64}$`)

func init() {
	checker.Register("docker", NewFromConfig)
}

// DockerChecker checks for Docker container image updates.
type DockerChecker struct {
	containers string
	exclude    []string
}

// NewFromConfig creates a DockerChecker from a watcher configuration.
func NewFromConfig(cfg config.WatcherConfig) (checker.Checker, error) {
	return &DockerChecker{
		containers: cfg.GetString("containers", "all"),
		exclude:    cfg.GetStringSlice("exclude", nil),
	}, nil
}

func (d *DockerChecker) Name() string { return "docker" }

// containerInfo holds parsed container data.
type containerInfo struct {
	Name    string
	Image   string
	ImageID string
}

func (d *DockerChecker) Check() (*checker.CheckResult, error) {
	result := &checker.CheckResult{
		CheckerName: d.Name(),
		CheckedAt:   time.Now(),
	}

	// List running containers using docker CLI
	containers, err := d.listContainers()
	if err != nil {
		return result, fmt.Errorf("failed to list containers: %w", err)
	}

	slog.Info("checking docker containers for updates", "count", len(containers))

	// Check each container for updates concurrently
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 5) // limit to 5 concurrent checks

	for _, c := range containers {
		wg.Add(1)
		go func(ci containerInfo) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			update, err := d.checkContainer(ci)
			if err != nil {
				slog.Warn("failed to check container", "name", ci.Name, "error", err)
				return
			}
			if update != nil {
				mu.Lock()
				result.Updates = append(result.Updates, *update)
				mu.Unlock()
			}
		}(c)
	}
	wg.Wait()

	if len(result.Updates) == 0 {
		result.Summary = "all containers are up to date"
	} else {
		result.Summary = fmt.Sprintf("%d containers", len(result.Updates))
	}

	return result, nil
}

func (d *DockerChecker) listContainers() ([]containerInfo, error) {
	// Use docker ps with custom format for reliable parsing
	res, err := executil.Run("docker", "ps", "--format", "{{.Names}}\t{{.Image}}\t{{.ID}}")
	if err != nil {
		return nil, fmt.Errorf("docker ps failed: %w", err)
	}

	var containers []containerInfo
	for _, line := range strings.Split(strings.TrimSpace(res.Stdout), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) < 3 {
			continue
		}

		name := parts[0]

		// Apply filters
		if d.containers != "all" {
			allowed := strings.Split(d.containers, ",")
			if !containsString(allowed, name) {
				continue
			}
		}
		if containsString(d.exclude, name) {
			continue
		}

		image := parts[1]

		// Skip containers running with a bare image ID (no name:tag to check against registry)
		if hexPattern.MatchString(image) {
			slog.Debug("skipping container with image ID reference", "name", name, "image", image)
			continue
		}

		// Skip locally-built images (no registry slash and no standard library prefix)
		if !strings.Contains(image, "/") && !strings.Contains(image, ".") {
			slog.Debug("skipping local image", "name", name, "image", image)
			continue
		}

		containers = append(containers, containerInfo{
			Name:    name,
			Image:   image,
			ImageID: parts[2],
		})
	}

	return containers, nil
}

func (d *DockerChecker) checkContainer(ci containerInfo) (*checker.Update, error) {
	slog.Debug("checking image", "container", ci.Name, "image", ci.Image)

	// Get local image's RepoDigest (the registry digest from when it was pulled)
	localRes, err := executil.Run("docker", "image", "inspect", "--format",
		"{{index .RepoDigests 0}}", ci.Image)
	if err != nil {
		return nil, fmt.Errorf("local image inspect failed: %w", err)
	}
	// RepoDigests format: "image@sha256:abc..." — extract digest after @
	localRepoDigest := strings.TrimSpace(localRes.Stdout)
	if idx := strings.Index(localRepoDigest, "@"); idx >= 0 {
		localRepoDigest = localRepoDigest[idx+1:]
	}
	if localRepoDigest == "" {
		return nil, fmt.Errorf("no RepoDigest found for %s (locally built image?)", ci.Image)
	}

	// Get remote digest WITHOUT pulling — read-only registry query
	remoteRes, err := executil.RunWithTimeout(30*time.Second,
		"docker", "buildx", "imagetools", "inspect", ci.Image)
	if err != nil {
		return nil, fmt.Errorf("remote inspect failed for %s (image may be locally built or private): %w", ci.Image, err)
	}

	// Parse "Digest: sha256:abc..." from output
	remoteDigest := ""
	for _, line := range strings.Split(remoteRes.Stdout, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Digest:") {
			remoteDigest = strings.TrimSpace(strings.TrimPrefix(line, "Digest:"))
			break
		}
	}
	if remoteDigest == "" {
		return nil, fmt.Errorf("could not parse remote digest for %s", ci.Image)
	}

	slog.Debug("digest comparison", "container", ci.Name,
		"local", shortDigest(localRepoDigest), "remote", shortDigest(remoteDigest))

	// Compare digests
	if localRepoDigest == remoteDigest {
		slog.Debug("container is up to date", "name", ci.Name)
		return nil, nil
	}

	return &checker.Update{
		Name:           ci.Name,
		CurrentVersion: shortDigest(localRepoDigest),
		NewVersion:     shortDigest(remoteDigest),
		Type:           checker.UpdateTypeImage,
		Priority:       checker.PriorityNormal,
		Source:         ci.Image,
	}, nil
}

func shortDigest(digest string) string {
	// sha256:abc123... -> abc123...
	if idx := strings.Index(digest, ":"); idx >= 0 {
		digest = digest[idx+1:]
	}
	if len(digest) > 12 {
		return digest[:12]
	}
	return digest
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if strings.TrimSpace(item) == s {
			return true
		}
	}
	return false
}
