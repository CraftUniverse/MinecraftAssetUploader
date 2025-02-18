package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var targets = []struct {
	OS   string
	Arch string
	Ext  string
}{
	{"windows", "arm64", ".exe"},
	{"windows", "amd64", ".exe"},
	{"linux", "amd64", ""},
	{"linux", "arm64", ""},
	{"darwin", "arm64", ""},
}

func main() {
	// Check for version argument
	var version string
	if len(os.Args) > 1 {
		version = os.Args[1]
	} else {
		var err error
		version, err = getGitTag()
		if err != nil {
			fmt.Println("⚠️  Failed to retrieve version:", err)
			version = "dev"
		}
	}

	fmt.Println("🔧 Building version:", version)

	// Create output directory
	outDir := "./dist"
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Println("⚠️  Failed to create output directory:", err)
		return
	}

	// Build for each target
	for _, target := range targets {
		buildTarget(version, target.OS, target.Arch, target.Ext, outDir)
	}

	fmt.Println("✅ Build completed!")
}

// Get latest Git tag
func getGitTag() (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// Build for a specific OS/Architecture
func buildTarget(version, goos, goarch, ext, outDir string) {
	fmt.Printf("🚀 Building for %s/%s...\n", goos, goarch)

	// Set environment variables for cross-compilation
	env := os.Environ()
	env = append(env, "GOOS="+goos, "GOARCH="+goarch)

	// Generate output filename
	outputName := filepath.Join(outDir, fmt.Sprintf("assetuploader-%s-%s%s", goos, goarch, ext))

	// Run build command
	cmd := exec.Command("go", "build", "-ldflags", "-w -X 'main.version="+version+"'", "-o", outputName, "./cmd/assetuploader/main.go")
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Execute build
	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Failed to build for %s/%s: %v\n", goos, goarch, err)
	} else {
		fmt.Printf("✅ Successfully built: %s\n", outputName)
	}
}
