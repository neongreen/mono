package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParsePRURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantOwner string
		wantRepo  string
		wantPRNum int
		wantErr   bool
	}{{name: "full URL with https", url: "https://github.com/neongreen/mono/pull/123", wantOwner: "neongreen", wantRepo: "mono", wantPRNum: 123, wantErr: false}, {name: "URL without protocol", url: "github.com/owner/repo/pull/456", wantOwner: "owner", wantRepo: "repo", wantPRNum: 456, wantErr: false}, {name: "invalid URL", url: "not-a-github-url", wantOwner: "", wantRepo: "", wantPRNum: 0, wantErr: true}, {name: "missing PR number", url: "github.com/owner/repo/pull/", wantOwner: "", wantRepo: "", wantPRNum: 0, wantErr: true}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePRURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePRURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if got.Owner != tt.wantOwner {
					t.Errorf("parsePRURL() Owner = %v, want %v", got.Owner, tt.wantOwner)
				}
				if got.Repo != tt.wantRepo {
					t.Errorf("parsePRURL() Repo = %v, want %v", got.Repo, tt.wantRepo)
				}
				if got.PRNum != tt.wantPRNum {
					t.Errorf("parsePRURL() PRNum = %v, want %v", got.PRNum, tt.wantPRNum)
				}
			}
		})
	}
}
func TestGetCacheDir(t *testing.
	T) {
	cacheDir,
		err := getCacheDir()
	if err != nil {
		t.Fatalf("getCacheDir() error = %v",

			err)
	}
	homeDir,
		_ := os.UserHomeDir()
	expectedDir := filepath.Join(homeDir, ".cache", "prrun")
	if cacheDir != expectedDir {
		t.Errorf("getCacheDir() = %v, want %v", cacheDir, expectedDir)
	}
}
