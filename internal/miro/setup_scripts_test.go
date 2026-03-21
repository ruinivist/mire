package miro

import (
	"path/filepath"
	"testing"

	"miro/internal/testutil"
)

func TestDiscoverSetupScriptsFindsRootToLeafSetupFiles(t *testing.T) {
	testDir := filepath.Join(t.TempDir(), "e2e")
	scenarioDir := filepath.Join(testDir, "suite", "spec")
	rootSetup := filepath.Join(testDir, setupScriptName)
	suiteSetup := filepath.Join(testDir, "suite", setupScriptName)
	scenarioSetup := filepath.Join(scenarioDir, setupScriptName)

	testutil.WriteFile(t, rootSetup, "export ROOT=1\n")
	testutil.WriteFile(t, suiteSetup, "export SUITE=1\n")
	testutil.WriteFile(t, scenarioSetup, "export SPEC=1\n")

	got, err := discoverSetupScripts(testDir, scenarioDir)
	if err != nil {
		t.Fatalf("discoverSetupScripts() error = %v", err)
	}

	want := []string{rootSetup, suiteSetup, scenarioSetup}
	if len(got) != len(want) {
		t.Fatalf("len(discoverSetupScripts()) = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("discoverSetupScripts()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
