package tests

import (
	"testing"

	"govard/internal/engine/remote"
)

func TestParseMagentoDBHostPort(t *testing.T) {
	testCases := []struct {
		name    string
		raw     string
		expectH string
		expectP int
	}{
		{name: "empty host", raw: "", expectH: "db", expectP: 3306},
		{name: "host only", raw: "database.internal", expectH: "database.internal", expectP: 3306},
		{name: "host and port", raw: "database.internal:3307", expectH: "database.internal", expectP: 3307},
		{name: "tcp prefix", raw: "tcp://db.example:3310", expectH: "db.example", expectP: 3310},
		{name: "ipv6 bracket host", raw: "[::1]:3309", expectH: "::1", expectP: 3309},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			host, port := remote.ParseMagentoDBHostPort(testCase.raw)
			if host != testCase.expectH {
				t.Fatalf("host mismatch: got %q want %q", host, testCase.expectH)
			}
			if port != testCase.expectP {
				t.Fatalf("port mismatch: got %d want %d", port, testCase.expectP)
			}
		})
	}
}

func TestNormalizeMagentoVersion(t *testing.T) {
	testCases := []struct {
		name   string
		raw    string
		expect string
	}{
		{name: "caret constraint", raw: "^2.4.7-p1", expect: "2.4.7-p1"},
		{name: "tilde constraint", raw: "~2.4.6", expect: "2.4.6"},
		{name: "comparison constraint", raw: ">=2.4.8 <2.5", expect: "2.4.8"},
		{name: "pipe constraint", raw: "2.4.6-p3 || 2.4.7", expect: "2.4.6-p3"},
		{name: "wildcard", raw: "2.4.x", expect: "2.4.x"},
		{name: "empty", raw: "", expect: ""},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			actual := remote.NormalizeMagentoVersion(testCase.raw)
			if actual != testCase.expect {
				t.Fatalf("version mismatch: got %q want %q", actual, testCase.expect)
			}
		})
	}
}
