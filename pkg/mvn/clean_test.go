package mvn

import (
	"testing"

	"yoreyuan/m2LocalRepoTools/pkg/config"
)

func TestMatchDirRegex(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		dirName  string
		want     bool
	}{
		{
			name:     "match ${antlr4.version}",
			patterns: []string{"${*.version}"},
			dirName:  "${antlr4.version}",
			want:     true,
		},
		{
			name:     "match ${guava.version}",
			patterns: []string{"${*.version}"},
			dirName:  "${guava.version}",
			want:     true,
		},
		{
			name:     "no match normal version",
			patterns: []string{"${*.version}"},
			dirName:  "4.10.1",
			want:     false,
		},
		{
			name:     "no match partial pattern",
			patterns: []string{"${*.version}"},
			dirName:  "${antlr4.release}",
			want:     false,
		},
		{
			name:     "match with multiple patterns",
			patterns: []string{"${*.version}", "${*.release}"},
			dirName:  "${foo.release}",
			want:     true,
		},
		{
			name:     "no match with empty patterns",
			patterns: []string{},
			dirName:  "${antlr4.version}",
			want:     false,
		},
		{
			name:     "match snapshot placeholder",
			patterns: []string{"${*.version}"},
			dirName:  "${project.version}",
			want:     true,
		},
		{
			name:     "no match without closing brace",
			patterns: []string{"${*.version}"},
			dirName:  "${antlr4.version",
			want:     false,
		},
		{
			name:     "Match the specific folder name",
			patterns: []string{"unknown"},
			dirName:  "unknown",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf = &config.Config{
				Clean: config.RepoClean{
					DirRegex: tt.patterns,
				},
			}

			got := matchDirRegex(tt.dirName)
			if got != tt.want {
				t.Errorf("matchDirRegex(%q) = %v, want %v", tt.dirName, got, tt.want)
			}
		})
	}
}
