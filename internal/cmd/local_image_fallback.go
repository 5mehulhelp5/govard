package cmd

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	dockerassets "govard/docker"
	"govard/internal/engine"
)

const (
	defaultGovardImageRepository = "ddtcorex/govard-"
)

var (
	embeddedDockerAssetsOnce sync.Once
	embeddedDockerAssetsDir  string
	embeddedDockerAssetsErr  error
	majorMinorVersionPattern = regexp.MustCompile(`^\d+\.\d+$`)
)

type localImageBuildArg struct {
	Name  string
	Value string
}

type localImageBuildSpec struct {
	ContextRel    string
	DockerfileRel string
	BuildArgs     []localImageBuildArg
}

// LocalBuildSpecForTest exposes resolved local build spec details for tests.
type LocalBuildSpecForTest struct {
	ContextRel    string
	DockerfileRel string
	BuildArgs     map[string]string
}

func fallbackBuildMissingGovardImagesFromCompose(composePath string, out io.Writer, errOut io.Writer) ([]string, error) {
	serviceImages, err := engine.ReadServiceImagesFromCompose(composePath)
	if err != nil {
		return nil, fmt.Errorf("read compose service images: %w", err)
	}

	uniqueImages := make([]string, 0, len(serviceImages))
	seen := make(map[string]struct{}, len(serviceImages))
	for _, image := range serviceImages {
		trimmed := strings.TrimSpace(image)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		uniqueImages = append(uniqueImages, trimmed)
	}
	sort.Strings(uniqueImages)

	buildableMissing := make([]string, 0, len(uniqueImages))
	nonBuildableMissing := make([]string, 0)
	for _, image := range uniqueImages {
		if imageExistsLocally(image) {
			continue
		}

		repoPrefix, service, tag, ok := parseGovardImageReference(image)
		if !ok {
			nonBuildableMissing = append(nonBuildableMissing, image)
			continue
		}
		if _, specErr := localBuildSpecForGovardService(service, tag, repoPrefix); specErr != nil {
			nonBuildableMissing = append(nonBuildableMissing, image)
			continue
		}
		buildableMissing = append(buildableMissing, image)
	}

	if len(buildableMissing) == 0 {
		if len(nonBuildableMissing) > 0 {
			return nil, fmt.Errorf(
				"missing images are not eligible for Govard local fallback build: %s",
				strings.Join(nonBuildableMissing, ", "),
			)
		}
		return nil, nil
	}

	dockerRoot, err := ensureDockerAssetsRoot(".")
	if err != nil {
		return nil, fmt.Errorf("resolve docker build contexts: %w", err)
	}

	built := make([]string, 0, len(buildableMissing))
	for _, image := range buildableMissing {
		if err := buildGovardImageLocally(image, dockerRoot, out, errOut); err != nil {
			return built, err
		}
		built = append(built, image)
	}

	if len(nonBuildableMissing) > 0 {
		return built, fmt.Errorf(
			"missing images are not eligible for Govard local fallback build: %s",
			strings.Join(nonBuildableMissing, ", "),
		)
	}

	return built, nil
}

func buildGovardImageLocally(image string, dockerRoot string, out io.Writer, errOut io.Writer) error {
	repoPrefix, service, tag, ok := parseGovardImageReference(image)
	if !ok {
		return fmt.Errorf("image %q is not a Govard-managed image", image)
	}

	spec, err := localBuildSpecForGovardService(service, tag, repoPrefix)
	if err != nil {
		return fmt.Errorf("local build spec for %s: %w", image, err)
	}

	contextPath := filepath.Join(dockerRoot, spec.ContextRel)
	if stat, statErr := os.Stat(contextPath); statErr != nil || !stat.IsDir() {
		if statErr != nil {
			return fmt.Errorf("build context %s: %w", contextPath, statErr)
		}
		return fmt.Errorf("build context %s is not a directory", contextPath)
	}

	args := []string{"build", "-t", image}
	if spec.DockerfileRel != "" {
		dockerfilePath := filepath.Join(dockerRoot, spec.DockerfileRel)
		if _, statErr := os.Stat(dockerfilePath); statErr != nil {
			return fmt.Errorf("dockerfile %s: %w", dockerfilePath, statErr)
		}
		args = append(args, "-f", dockerfilePath)
	}
	for _, buildArg := range spec.BuildArgs {
		args = append(args, "--build-arg", buildArg.Name+"="+buildArg.Value)
	}
	args = append(args, contextPath)

	command := exec.Command("docker", args...)
	command.Stdout = normalizeWriter(out)
	command.Stderr = normalizeWriter(errOut)
	if err := command.Run(); err != nil {
		return fmt.Errorf("docker %s: %w", strings.Join(args, " "), err)
	}
	return nil
}

func localBuildSpecForGovardService(service string, tag string, repoPrefix string) (localImageBuildSpec, error) {
	service = strings.TrimSpace(service)
	tag = strings.TrimSpace(tag)
	if tag == "" {
		tag = "latest"
	}
	if strings.TrimSpace(repoPrefix) == "" {
		repoPrefix = defaultGovardImageRepository
	}

	switch service {
	case "apache":
		return localImageBuildSpec{
			ContextRel: "apache",
			BuildArgs: []localImageBuildArg{
				{Name: "APACHE_VERSION", Value: resolveApacheBuildVersion(tag)},
			},
		}, nil
	case "nginx":
		return localImageBuildSpec{
			ContextRel: "nginx",
			BuildArgs: []localImageBuildArg{
				{Name: "NGINX_VERSION", Value: resolveNginxBuildVersion(tag)},
			},
		}, nil
	case "php":
		return localImageBuildSpec{
			ContextRel: "php",
			BuildArgs: []localImageBuildArg{
				{Name: "PHP_VERSION", Value: tag},
			},
		}, nil
	case "php-magento2":
		return localImageBuildSpec{
			ContextRel:    "php",
			DockerfileRel: filepath.Join("php", "magento2", "Dockerfile"),
			BuildArgs: []localImageBuildArg{
				{Name: "PHP_VERSION", Value: tag},
				{Name: "GOVARD_IMAGE_REPOSITORY", Value: repoPrefix},
			},
		}, nil
	case "mariadb":
		return localImageBuildSpec{
			ContextRel: "mariadb",
			BuildArgs: []localImageBuildArg{
				{Name: "MARIADB_VERSION", Value: tag},
			},
		}, nil
	case "mysql":
		return localImageBuildSpec{
			ContextRel: "mysql",
			BuildArgs: []localImageBuildArg{
				{Name: "MYSQL_VERSION", Value: tag},
			},
		}, nil
	case "redis":
		return localImageBuildSpec{
			ContextRel: "redis",
			BuildArgs: []localImageBuildArg{
				{Name: "REDIS_VERSION", Value: tag},
			},
		}, nil
	case "valkey":
		return localImageBuildSpec{
			ContextRel: "valkey",
			BuildArgs: []localImageBuildArg{
				{Name: "VALKEY_VERSION", Value: tag},
			},
		}, nil
	case "rabbitmq":
		return localImageBuildSpec{
			ContextRel: "rabbitmq",
			BuildArgs: []localImageBuildArg{
				{Name: "RABBITMQ_VERSION", Value: tag},
			},
		}, nil
	case "opensearch":
		return localImageBuildSpec{
			ContextRel: "opensearch",
			BuildArgs: []localImageBuildArg{
				{Name: "OPENSEARCH_VERSION", Value: tag},
			},
		}, nil
	case "elasticsearch":
		elasticsearchImage := "docker.elastic.co/elasticsearch/elasticsearch"
		if tag == "2.4.6" {
			elasticsearchImage = "elasticsearch"
		}
		return localImageBuildSpec{
			ContextRel: "elasticsearch",
			BuildArgs: []localImageBuildArg{
				{Name: "ELASTICSEARCH_VERSION", Value: tag},
				{Name: "ELASTICSEARCH_IMAGE", Value: elasticsearchImage},
			},
		}, nil
	case "varnish":
		varnishVersion, varnishImageTag := resolveVarnishBuildVersions(tag)
		return localImageBuildSpec{
			ContextRel: "varnish",
			BuildArgs: []localImageBuildArg{
				{Name: "VARNISH_VERSION", Value: varnishVersion},
				{Name: "VARNISH_IMAGE_TAG", Value: varnishImageTag},
			},
		}, nil
	case "dnsmasq":
		return localImageBuildSpec{
			ContextRel: "dnsmasq",
		}, nil
	default:
		return localImageBuildSpec{}, fmt.Errorf("unsupported Govard image service %q", service)
	}
}

func parseGovardImageReference(image string) (repoPrefix string, service string, tag string, ok bool) {
	repository, version := splitImageRepositoryAndTag(image)
	if repository == "" {
		return "", "", "", false
	}

	markerIndex := strings.LastIndex(repository, "govard-")
	if markerIndex < 0 {
		return "", "", "", false
	}

	repoPrefix = repository[:markerIndex+len("govard-")]
	service = strings.TrimSpace(repository[markerIndex+len("govard-"):])
	if service == "" {
		return "", "", "", false
	}
	return repoPrefix, service, version, true
}

func splitImageRepositoryAndTag(image string) (repository string, tag string) {
	trimmed := strings.TrimSpace(image)
	if trimmed == "" {
		return "", ""
	}

	if atIndex := strings.Index(trimmed, "@"); atIndex >= 0 {
		trimmed = trimmed[:atIndex]
	}

	lastSlash := strings.LastIndex(trimmed, "/")
	lastColon := strings.LastIndex(trimmed, ":")
	if lastColon > lastSlash {
		return trimmed[:lastColon], strings.TrimSpace(trimmed[lastColon+1:])
	}
	return trimmed, "latest"
}

func resolveApacheBuildVersion(tag string) string {
	switch strings.TrimSpace(tag) {
	case "", "latest", "2.4":
		return "2.4.66"
	default:
		return tag
	}
}

func resolveNginxBuildVersion(tag string) string {
	tag = strings.TrimSpace(tag)
	switch tag {
	case "", "latest", "1.28":
		return "1.28.0"
	default:
		if majorMinorVersionPattern.MatchString(tag) {
			return tag + ".0"
		}
		return tag
	}
}

func resolveVarnishBuildVersions(tag string) (version string, imageTag string) {
	tag = strings.TrimSpace(tag)
	switch tag {
	case "", "latest":
		return "7.6", "7.6"
	case "6.0":
		return "6.0", "6.0"
	default:
		return tag, tag
	}
}

func imageExistsLocally(image string) bool {
	command := exec.Command("docker", "image", "inspect", image)
	return command.Run() == nil
}

func ensureDockerAssetsRoot(startDir string) (string, error) {
	if root, err := findDockerAssetsDir(startDir); err == nil {
		return root, nil
	}
	return ensureEmbeddedDockerAssetsDir()
}

func findDockerAssetsDir(startDir string) (string, error) {
	if override := strings.TrimSpace(os.Getenv("GOVARD_DOCKER_DIR")); override != "" {
		if abs, err := filepath.Abs(override); err == nil && isDockerAssetsDir(abs) {
			return abs, nil
		}
	}

	absStart, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}

	current := absStart
	for {
		candidate := filepath.Join(current, "docker")
		if isDockerAssetsDir(candidate) {
			return candidate, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	for _, candidate := range []string{
		"/usr/local/share/govard/docker",
		"/usr/share/govard/docker",
	} {
		if isDockerAssetsDir(candidate) {
			return candidate, nil
		}
	}

	if executablePath, err := os.Executable(); err == nil {
		executableDir := filepath.Dir(executablePath)
		for _, candidate := range []string{
			filepath.Join(executableDir, "docker"),
			filepath.Join(executableDir, "..", "docker"),
		} {
			clean := filepath.Clean(candidate)
			if isDockerAssetsDir(clean) {
				return clean, nil
			}
		}
	}

	return "", fmt.Errorf("docker build assets directory not found")
}

func ensureEmbeddedDockerAssetsDir() (string, error) {
	embeddedDockerAssetsOnce.Do(func() {
		tempDir, err := os.MkdirTemp("", "govard-docker-assets-*")
		if err != nil {
			embeddedDockerAssetsErr = fmt.Errorf("create temp docker assets dir: %w", err)
			return
		}
		if err := materializeDockerAssetsFS(dockerassets.FS, tempDir); err != nil {
			embeddedDockerAssetsErr = fmt.Errorf("materialize embedded docker assets: %w", err)
			return
		}
		embeddedDockerAssetsDir = tempDir
	})
	if embeddedDockerAssetsErr != nil {
		return "", embeddedDockerAssetsErr
	}
	return embeddedDockerAssetsDir, nil
}

func materializeDockerAssetsFS(source fs.FS, destination string) error {
	return fs.WalkDir(source, ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == "." {
			return nil
		}

		targetPath := filepath.Join(destination, path)
		if entry.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}

		content, err := fs.ReadFile(source, path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return err
		}
		return os.WriteFile(targetPath, content, 0o644)
	})
}

func isDockerAssetsDir(root string) bool {
	if strings.TrimSpace(root) == "" {
		return false
	}
	requiredPaths := []string{
		filepath.Join(root, "docker-bake.hcl"),
		filepath.Join(root, "php", "Dockerfile"),
		filepath.Join(root, "dnsmasq", "Dockerfile"),
	}
	for _, requiredPath := range requiredPaths {
		if _, err := os.Stat(requiredPath); err != nil {
			return false
		}
	}
	return true
}

func normalizeWriter(writer io.Writer) io.Writer {
	if writer == nil {
		return io.Discard
	}
	return writer
}

// ParseGovardImageReferenceForTest exposes Govard image parsing for tests.
func ParseGovardImageReferenceForTest(image string) (string, string, string, bool) {
	return parseGovardImageReference(image)
}

// ResolveLocalBuildSpecForTest exposes local build spec resolution for tests.
func ResolveLocalBuildSpecForTest(service string, tag string, repositoryPrefix string) (LocalBuildSpecForTest, error) {
	spec, err := localBuildSpecForGovardService(service, tag, repositoryPrefix)
	if err != nil {
		return LocalBuildSpecForTest{}, err
	}
	buildArgs := make(map[string]string, len(spec.BuildArgs))
	for _, buildArg := range spec.BuildArgs {
		buildArgs[buildArg.Name] = buildArg.Value
	}
	return LocalBuildSpecForTest{
		ContextRel:    spec.ContextRel,
		DockerfileRel: spec.DockerfileRel,
		BuildArgs:     buildArgs,
	}, nil
}
