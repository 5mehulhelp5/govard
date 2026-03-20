package tests

import (
	"testing"

	"govard/internal/cmd"
)

func TestParseGovardImageReference(t *testing.T) {
	repositoryPrefix, service, tag, ok := cmd.ParseGovardImageReferenceForTest("ddtcorex/govard-php:8.3")
	if !ok {
		t.Fatal("expected govard image reference parse to succeed")
	}
	if repositoryPrefix != "ddtcorex/govard-" {
		t.Fatalf("unexpected repository prefix %q", repositoryPrefix)
	}
	if service != "php" {
		t.Fatalf("unexpected service %q", service)
	}
	if tag != "8.3" {
		t.Fatalf("unexpected tag %q", tag)
	}
}

func TestParseGovardImageReferenceRejectsThirdParty(t *testing.T) {
	_, _, _, ok := cmd.ParseGovardImageReferenceForTest("node:24-alpine")
	if ok {
		t.Fatal("expected parse to reject non-govard image")
	}
}

func TestLocalBuildSpecPHPMagento2(t *testing.T) {
	spec, err := cmd.ResolveLocalBuildSpecForTest("php-magento2", "8.4", "local/govard-")
	if err != nil {
		t.Fatalf("local build spec: %v", err)
	}

	if spec.ContextRel != "php" {
		t.Fatalf("unexpected context %q", spec.ContextRel)
	}
	if spec.DockerfileRel != "php/magento2/Dockerfile" {
		t.Fatalf("unexpected dockerfile %q", spec.DockerfileRel)
	}
	if spec.BuildArgs["PHP_VERSION"] != "8.4" {
		t.Fatalf("expected PHP_VERSION build arg, got %q", spec.BuildArgs["PHP_VERSION"])
	}
	if spec.BuildArgs["GOVARD_IMAGE_REPOSITORY"] != "local/govard-" {
		t.Fatalf(
			"expected GOVARD_IMAGE_REPOSITORY build arg local/govard-, got %q",
			spec.BuildArgs["GOVARD_IMAGE_REPOSITORY"],
		)
	}
}

func TestLocalBuildSpecVarnishLatest(t *testing.T) {
	spec, err := cmd.ResolveLocalBuildSpecForTest("varnish", "latest", "ddtcorex/govard-")
	if err != nil {
		t.Fatalf("local build spec: %v", err)
	}

	if spec.ContextRel != "varnish" {
		t.Fatalf("unexpected context %q", spec.ContextRel)
	}
	if spec.BuildArgs["VARNISH_VERSION"] != "7.6" {
		t.Fatalf("expected VARNISH_VERSION=7.6, got %q", spec.BuildArgs["VARNISH_VERSION"])
	}
	if spec.BuildArgs["VARNISH_IMAGE_TAG"] != "7.6" {
		t.Fatalf("expected VARNISH_IMAGE_TAG=7.6, got %q", spec.BuildArgs["VARNISH_IMAGE_TAG"])
	}
}

func TestLocalBuildSpecUnsupportedService(t *testing.T) {
	if _, err := cmd.ResolveLocalBuildSpecForTest("unknown", "latest", "ddtcorex/govard-"); err == nil {
		t.Fatal("expected unsupported service to return error")
	}
}

func TestLocalBuildSpecPHPMagento2Debug(t *testing.T) {
	spec, err := cmd.ResolveLocalBuildSpecForTest("php-magento2", "8.3-debug", "ddtcorex/govard-")
	if err != nil {
		t.Fatalf("local build spec: %v", err)
	}

	if spec.ContextRel != "php" {
		t.Fatalf("unexpected context %q", spec.ContextRel)
	}
	if spec.DockerfileRel != "php/debug/Dockerfile" {
		t.Fatalf("unexpected dockerfile %q, expected php/debug/Dockerfile", spec.DockerfileRel)
	}
	if spec.BuildArgs["BASE_IMAGE"] != "ddtcorex/govard-php-magento2:8.3" {
		t.Fatalf("expected BASE_IMAGE=ddtcorex/govard-php-magento2:8.3, got %q", spec.BuildArgs["BASE_IMAGE"])
	}
	if len(spec.Dependencies) == 0 || spec.Dependencies[0] != "ddtcorex/govard-php-magento2:8.3" {
		t.Fatalf("expected dependency on base image, got %v", spec.Dependencies)
	}
}
