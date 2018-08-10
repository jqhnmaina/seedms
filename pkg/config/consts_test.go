package config_test

import (
	"testing"

	"bitbucket.org/doveria/care-server"
)

func TestVersionMajorPrefixed(t *testing.T) {
	tcs := []struct {
		name        string
		versionFull string
		sep         string
		expect      string
	}{
		{name: "major version", versionFull: "5.4.1", sep: "+", expect: "v5"},
		{name: "minor version", versionFull: "0.4.1", sep: "+", expect: "v0+4"},
		{name: "patch version", versionFull: "0.0.1", sep: "+", expect: "v0+0+1"},
		{name: "patch version suffixed", versionFull: "0.0.1-alpha", sep: "+", expect: "v0+0+1"},
		{name: "short with suffix", versionFull: "0.1-alpha", sep: "+", expect: "v0+1"},
		{name: "all zero", versionFull: "0.0.0", sep: "+", expect: "v0"},
		{name: "too many integers", versionFull: "0.0.0.1", sep: "+", expect: "v0"},
		{name: "all zero short", versionFull: "0.0", sep: "+", expect: "v0"},
		{name: "empty version", versionFull: "", sep: "+", expect: "v0"},
		{name: "empty sep", versionFull: "0.4.1", sep: "", expect: "v04"},
		{name: "different sep", versionFull: "0.4.1", sep: "_", expect: "v0_4"},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			act := care_server.VersionMajorPrefixed(tc.versionFull, tc.sep)
			if act != tc.expect {
				t.Errorf("Expected '%s' but got '%s'", tc.expect, act)
			}
		})
	}
}