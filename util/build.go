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

type target struct {
	OS   string
	Arch string
	Ext  string
}

var targets = []target{
	{OS: "linux", Arch: "amd64"},
	{OS: "linux", Arch: "arm64"},
	{OS: "darwin", Arch: "amd64"},
	{OS: "darwin", Arch: "arm64"},
	{OS: "windows", Arch: "amd64", Ext: ".exe"},
	{OS: "windows", Arch: "arm64", Ext: ".exe"},
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "build failed:", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "clean":
			return clean()

		case "dist":
			return buildDist()

		case "local", "build":
			return buildLocal()

		case "bundle":
			return buildBundle(os.Args[2:])

		case "run":
			return runInterpreter(os.Args[2:])

		case "test":
			return goTest()

		case "fmt":
			return goFmt()

		case "help", "-h", "--help":
			printHelp()
			return nil

		default:
			return fmt.Errorf("unknown command %q", os.Args[1])
		}
	}

	return buildDist()
}

func printHelp() {
	fmt.Println("Usage:")
	fmt.Println("  go run ./util/build.go                         build all dist binaries")
	fmt.Println("  go run ./util/build.go dist                    build all dist binaries")
	fmt.Println("  go run ./util/build.go local                   build local interpreter binary")
	fmt.Println("  go run ./util/build.go build                   build local interpreter binary")
	fmt.Println("  go run ./util/build.go bundle [args...]        run interpreter build [args...]")
	fmt.Println("  go run ./util/build.go run [args...]           run interpreter [args...]")
	fmt.Println("  go run ./util/build.go test                    run tests")
	fmt.Println("  go run ./util/build.go fmt                     format source")
	fmt.Println("  go run ./util/build.go clean                   remove build outputs")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run ./util/build.go run examples/1.stult")
	fmt.Println("  go run ./util/build.go bundle examples/bool -o bool-app")
}

func buildDist() error {
	if err := clean(); err != nil {
		return err
	}

	if err := os.MkdirAll(distDir, 0755); err != nil {
		return err
	}

	for _, t := range targets {
		out := filepath.Join(distDir, fmt.Sprintf("%s-%s-%s%s", appName, t.OS, t.Arch, t.Ext))
		fmt.Printf("building %s/%s -> %s\n", t.OS, t.Arch, out)

		if err := goBuild(out, t.OS, t.Arch); err != nil {
			return err
		}
	}

	fmt.Println("done")
	return nil
}

func buildLocal() error {
	out := appName
	if runtime.GOOS == "windows" {
		out += ".exe"
	}

	fmt.Printf("building local interpreter binary -> %s\n", out)
	return goBuild(out, runtime.GOOS, runtime.GOARCH)
}

func buildBundle(args []string) error {
	cmdArgs := append([]string{"run", "./" + srcDir, "build"}, args...)
	return runCommand("go", cmdArgs...)
}

func clean() error {
	fmt.Println("cleaning build outputs")

	if err := os.RemoveAll(distDir); err != nil {
		return err
	}

	if err := os.Remove(appName); err != nil && !os.IsNotExist(err) {
		return err
	}

	if err := os.Remove(appName + ".exe"); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func runInterpreter(args []string) error {
	cmdArgs := append([]string{"run", "./" + srcDir}, args...)
	return runCommand("go", cmdArgs...)
}

func goTest() error {
	return runCommand("go", "test", "./"+srcDir+"/...")
}

func goFmt() error {
	return runCommand("go", "fmt", "./"+srcDir+"/...")
}

func goBuild(output, goos, goarch string) error {
	cmd := exec.Command("go", "build", "-ldflags", "-s -w", "-o", output, "./"+srcDir)
	cmd.Env = append(os.Environ(),
		"CGO_ENABLED=0",
		"GOOS="+goos,
		"GOARCH="+goarch,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
