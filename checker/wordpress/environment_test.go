package wordpress

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectEnvironment_Ddev(t *testing.T) {
	dir := t.TempDir()
	ddevDir := filepath.Join(dir, ".ddev")
	os.Mkdir(ddevDir, 0755)
	os.WriteFile(filepath.Join(ddevDir, "config.yaml"), []byte("project_type: wordpress\n"), 0644)

	env := DetectEnvironment(dir)
	if env != EnvDdev {
		t.Errorf("expected EnvDdev, got %s", env)
	}
}

func TestDetectEnvironment_Lando(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".lando.yml"), []byte("recipe: wordpress\n"), 0644)

	env := DetectEnvironment(dir)
	if env != EnvLando {
		t.Errorf("expected EnvLando, got %s", env)
	}
}

func TestDetectEnvironment_WpEnv(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".wp-env.json"), []byte(`{"core": "WordPress/WordPress#master"}`), 0644)

	env := DetectEnvironment(dir)
	if env != EnvWpEnv {
		t.Errorf("expected EnvWpEnv, got %s", env)
	}
}

func TestDetectEnvironment_Bedrock(t *testing.T) {
	dir := t.TempDir()
	composerJSON := `{"require": {"roots/wordpress": "^6.0"}}`
	os.WriteFile(filepath.Join(dir, "composer.json"), []byte(composerJSON), 0644)

	env := DetectEnvironment(dir)
	if env != EnvBedrock {
		t.Errorf("expected EnvBedrock, got %s", env)
	}
}

func TestDetectEnvironment_DockerCompose(t *testing.T) {
	dir := t.TempDir()
	composeYml := "services:\n  wordpress:\n    image: wordpress:latest\n"
	os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte(composeYml), 0644)

	env := DetectEnvironment(dir)
	if env != EnvDockerCompose {
		t.Errorf("expected EnvDockerCompose, got %s", env)
	}
}

func TestDetectEnvironment_Native(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "wp-config.php"), []byte("<?php\n"), 0644)

	env := DetectEnvironment(dir)
	if env != EnvNative {
		t.Errorf("expected EnvNative, got %s", env)
	}
}

func TestDetectEnvironment_WalksUp(t *testing.T) {
	// Create ddev marker in parent, check from subdirectory
	parent := t.TempDir()
	child := filepath.Join(parent, "web", "wp")
	os.MkdirAll(child, 0755)

	ddevDir := filepath.Join(parent, ".ddev")
	os.Mkdir(ddevDir, 0755)
	os.WriteFile(filepath.Join(ddevDir, "config.yaml"), []byte("project_type: wordpress\n"), 0644)

	env := DetectEnvironment(child)
	if env != EnvDdev {
		t.Errorf("expected EnvDdev when walking up, got %s", env)
	}
}

func TestDetectEnvironment_DdevPriority(t *testing.T) {
	// When both ddev and composer.json exist, ddev should win
	dir := t.TempDir()
	ddevDir := filepath.Join(dir, ".ddev")
	os.Mkdir(ddevDir, 0755)
	os.WriteFile(filepath.Join(ddevDir, "config.yaml"), []byte("project_type: wordpress\n"), 0644)
	composerJSON := `{"require": {"roots/wordpress": "^6.0"}}`
	os.WriteFile(filepath.Join(dir, "composer.json"), []byte(composerJSON), 0644)

	env := DetectEnvironment(dir)
	if env != EnvDdev {
		t.Errorf("expected EnvDdev (higher priority), got %s", env)
	}
}

func TestDetectEnvironment_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	env := DetectEnvironment(dir)
	if env != EnvNative {
		t.Errorf("expected EnvNative for empty dir, got %s", env)
	}
}

func TestFindProjectDir_Ddev(t *testing.T) {
	parent := t.TempDir()
	child := filepath.Join(parent, "web", "wp")
	os.MkdirAll(child, 0755)

	ddevDir := filepath.Join(parent, ".ddev")
	os.Mkdir(ddevDir, 0755)
	os.WriteFile(filepath.Join(ddevDir, "config.yaml"), []byte("project_type: wordpress\n"), 0644)

	projectDir := FindProjectDir(child, EnvDdev)
	if projectDir != parent {
		t.Errorf("expected project dir %q, got %q", parent, projectDir)
	}
}

func TestBuildCommand_Ddev(t *testing.T) {
	spec := BuildCommand(EnvDdev, "/some/path", "/project/root")
	if spec.Command != "ddev" {
		t.Errorf("expected command 'ddev', got %q", spec.Command)
	}
	if len(spec.Args) < 1 || spec.Args[0] != "wp" {
		t.Errorf("expected first arg 'wp', got %v", spec.Args)
	}
	if spec.WorkDir != "/project/root" {
		t.Errorf("expected workdir '/project/root', got %q", spec.WorkDir)
	}
	if spec.NeedsPath {
		t.Error("ddev should not need --path")
	}
}

func TestBuildCommand_Lando(t *testing.T) {
	spec := BuildCommand(EnvLando, "/some/path", "/project/root")
	if spec.Command != "lando" {
		t.Errorf("expected command 'lando', got %q", spec.Command)
	}
	if spec.NeedsPath {
		t.Error("lando should not need --path")
	}
}

func TestBuildCommand_Native(t *testing.T) {
	spec := BuildCommand(EnvNative, "/var/www/html", "")
	if spec.Command != "wp" {
		t.Errorf("expected command 'wp', got %q", spec.Command)
	}
	if !spec.NeedsPath {
		t.Error("native should need --path")
	}
	if !spec.NeedsRunAs {
		t.Error("native should need run_as")
	}
}

func TestEnvironment_IsContainerBased(t *testing.T) {
	containerEnvs := []Environment{EnvDdev, EnvLando, EnvWpEnv, EnvDockerCompose}
	for _, e := range containerEnvs {
		if !e.IsContainerBased() {
			t.Errorf("expected %s to be container-based", e)
		}
	}

	hostEnvs := []Environment{EnvNative, EnvBedrock, EnvMAMP, EnvXAMPP, EnvValet, EnvLocalWP, EnvLaragon}
	for _, e := range hostEnvs {
		if e.IsContainerBased() {
			t.Errorf("expected %s to NOT be container-based", e)
		}
	}
}

func TestEnvironment_Label(t *testing.T) {
	if EnvDdev.Label() != "ddev" {
		t.Errorf("expected 'ddev', got %q", EnvDdev.Label())
	}
	if EnvNative.Label() != "Native (wp-cli)" {
		t.Errorf("expected 'Native (wp-cli)', got %q", EnvNative.Label())
	}
}

func TestParseSites_WithEnvironment(t *testing.T) {
	raw := []map[string]interface{}{
		{
			"name":        "My Blog",
			"path":        "/tmp/nonexistent",
			"environment": "ddev",
		},
	}

	sites := parseSites(raw)
	if len(sites) != 1 {
		t.Fatalf("expected 1 site, got %d", len(sites))
	}
	if sites[0].Environment != EnvDdev {
		t.Errorf("expected EnvDdev, got %s", sites[0].Environment)
	}
}
