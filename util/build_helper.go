//go:build ignore

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const appName = "stult"
const distDir = "dist"
const srcDir = "src"

type buildTarget struct {
	OS   string
	Arch string
	Ext  string
}

var buildTargets = []buildTarget{
	{OS: "linux", Arch: "amd64"},
	{OS: "linux", Arch: "arm64"},
	{OS: "darwin", Arch: "amd64"},
	{OS: "darwin", Arch: "arm64"},
	{OS: "windows", Arch: "amd64", Ext: ".exe"},
	{OS: "windows", Arch: "arm64", Ext: ".exe"},
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "build helper failed:", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) <= 1 {
		printUsage()
		return nil
	}

	switch os.Args[1] {
	case "local":
		return buildLocal()

	case "dist":
		return buildDist()

	case "clean":
		return clean()

	case "help", "-h", "--help":
		printUsage()
		return nil

	default:
		printUsage()
		return fmt.Errorf("unknown command %q", os.Args[1])
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  go run ./util/build_helper.go local   build local stult executable")
	fmt.Println("  go run ./util/build_helper.go dist    build all dist stult executables")
	fmt.Println("  go run ./util/build_helper.go clean   remove build outputs")
	fmt.Println("  go run ./util/build_helper.go help    show this usage message")
}

func buildLocal() error {
	output := appName
	if runtime.GOOS == "windows" {
		output += ".exe"
	}

	fmt.Printf("building local executable -> %s\n", output)

	return goBuild(output, runtime.GOOS, runtime.GOARCH)
}

func buildDist() error {
	if err := cleanDist(); err != nil {
		return err
	}

	if err := os.MkdirAll(distDir, 0755); err != nil {
		return fmt.Errorf("could not create dist directory: %w", err)
	}

	for _, target := range buildTargets {
		output := filepath.Join(
			distDir,
			fmt.Sprintf(
				"%s-%s-%s%s",
				appName,
				target.OS,
				target.Arch,
				target.Ext,
			),
		)

		fmt.Printf("building %s/%s -> %s\n", target.OS, target.Arch, output)

		if err := goBuild(output, target.OS, target.Arch); err != nil {
			return err
		}
	}

	fmt.Println("done")

	return nil
}

func clean() error {
	if err := cleanDist(); err != nil {
		return err
	}

	if err := removeIfExists(appName); err != nil {
		return err
	}

	if err := removeIfExists(appName + ".exe"); err != nil {
		return err
	}

	fmt.Println("cleaned build outputs")

	return nil
}

func cleanDist() error {
	if err := os.RemoveAll(distDir); err != nil {
		return fmt.Errorf("could not remove dist directory: %w", err)
	}

	return nil
}

func removeIfExists(filename string) error {
	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("could not remove %q: %w", filename, err)
	}

	return nil
}

func goBuild(output string, goos string, goarch string) error {
	command := exec.Command(
		"go",
		"build",
		"-trimpath",
		"-ldflags",
		"-s -w",
		"-o",
		output,
		"./"+srcDir,
	)

	command.Env = append(
		os.Environ(),
		"CGO_ENABLED=0",
		"GOOS="+goos,
		"GOARCH="+goarch,
	)

	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	if err := command.Run(); err != nil {
		return fmt.Errorf("go build failed for %s/%s: %w", goos, goarch, err)
	}

	return nil
}
