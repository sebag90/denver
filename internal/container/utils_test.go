package container

import (
	"os"
	"testing"
)

func TestContainerNameUsesDvPrefix(t *testing.T) {
	got := ContainerName()
	if prefix := "dv."; len(got) <= len(prefix) || got[:len(prefix)] != prefix {
		t.Errorf("ContainerName() = %q, want a %q prefix", got, prefix)
	}
	if !isManagedName(got) {
		t.Errorf("ContainerName() = %q, want it to be recognised as managed", got)
	}
}

func TestLegacyContainerNameUsesDenverPrefix(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	got := LegacyContainerName()
	want := legacyContainerNamePrefix + hash(dir)
	if got != want {
		t.Errorf("LegacyContainerName() = %q, want %q", got, want)
	}
}

func TestIsManagedName(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"dv.123", true},
		{"dv.0", true},
		{"denver.123", true},
		{"dv.", false},
		{"denver.", false},
		{"dv.12a", false},
		{"denver.12-3", false},
		{"dv-web", false},
		{"unrelated", false},
	}
	for _, tc := range tests {
		if got := isManagedName(tc.name); got != tc.want {
			t.Errorf("isManagedName(%q) = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestIsManagedContainer(t *testing.T) {
	tests := []struct {
		name string
		c    Container
		want bool
	}{
		{"current name", Container{Names: []string{"/dv.1"}}, true},
		{"legacy name", Container{Names: []string{"/denver.1"}}, true},
		{"current project label", Container{Names: []string{"/renamed"}, Labels: map[string]string{projectLabel: "dv.1"}}, true},
		{"sandbox label", Container{Names: []string{"/renamed"}, Labels: map[string]string{sandboxLabel: "true"}}, true},
		{"unrelated", Container{Names: []string{"/unrelated"}}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := isManagedContainer(tc.c); got != tc.want {
				t.Errorf("isManagedContainer(%v) = %v, want %v", tc.c.Names, got, tc.want)
			}
		})
	}
}
