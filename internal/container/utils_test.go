package container

import (
	"testing"
)

func TestContainerNameUsesPaguroPrefix(t *testing.T) {
	got := ContainerName()
	if prefix := "paguro."; len(got) <= len(prefix) || got[:len(prefix)] != prefix {
		t.Errorf("ContainerName() = %q, want a %q prefix", got, prefix)
	}
	if !isManagedName(got) {
		t.Errorf("ContainerName() = %q, want it to be recognised as managed", got)
	}
}

func TestIsManagedName(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"paguro.123", true},
		{"paguro.0", true},
		{"paguro.", false},
		{"paguro.12a", false},
		{"paguro.12-3", false},
		{"paguro-web", false},
		{"dv.123", false},
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
		{"name", Container{Names: []string{"/paguro.1"}}, true},
		{"project label", Container{Names: []string{"/renamed"}, Labels: map[string]string{projectLabel: "paguro.1"}}, true},
		{"sandbox label", Container{Names: []string{"/renamed"}, Labels: map[string]string{sandboxLabel: "true"}}, true},
		{"old name is not managed", Container{Names: []string{"/dv.1"}}, false},
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
