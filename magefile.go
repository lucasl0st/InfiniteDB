//go:build mage

/*
 * Copyright (c) 2023 Lucas Pape
 */

package main

import (
	"errors"
	"fmt"
	"github.com/magefile/mage/sh"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var Default = Build

const binaryName = "infinitedb-server"
const buildDir = "build"

var tools = []string{
	"idbdump",
	"idbcli",
	"idbimport",
}

type dockerImage struct {
	tag       string
	multiArch bool
}

var dockerImages = []dockerImage{
	{
		tag:       "ghcr.io/lucasl0st/infinitedb",
		multiArch: false,
	},
	{
		tag:       "lucasl0st/infinitedb",
		multiArch: true,
	},
}

type architecture string

const archArm64 architecture = "arm64"
const archAmd64 architecture = "amd64"

type operating_system string

const osDarwin operating_system = "darwin"
const osLinux operating_system = "linux"
const osFreeBsd operating_system = "freebsd"
const osWindows operating_system = "windows"

type dockerPlatform string

const dockerPlatformLinuxArm64 dockerPlatform = "linux/amd64"
const dockerPlatformLinuxAmd64 dockerPlatform = "linux/arm64/v8"

type target struct {
	Arch architecture
	Os   operating_system
}

type result struct {
	t          target
	path       string
	binaryName string
}

var supportedTargets = []target{
	{
		Arch: archArm64,
		Os:   osDarwin,
	},
	{
		Arch: archAmd64,
		Os:   osDarwin,
	},
	{
		Arch: archArm64,
		Os:   osLinux,
	},
	{
		Arch: archAmd64,
		Os:   osLinux,
	},
	{
		Arch: archArm64,
		Os:   osFreeBsd,
	},
	{
		Arch: archAmd64,
		Os:   osFreeBsd,
	},
	{
		Arch: archArm64,
		Os:   osWindows,
	},
	{
		Arch: archAmd64,
		Os:   osWindows,
	},
}

func Docker_all(push bool, tagLatest bool) error {
	results, err := buildForOs(osLinux)

	if err != nil {
		return err
	}

	v, err := version()

	if err != nil {
		return err
	}

	binary := fmt.Sprintf("%s/%s_%s_%s", buildDir, binaryName, v, osLinux)

	platform := ""

	for _, result := range results {
		p, err := targetToDockerPlatform(result.t)

		if err != nil {
			return err
		}

		platform += fmt.Sprint(*p) + ","
	}

	platform = strings.TrimSuffix(platform, ",")

	args := []string{
		"buildx", "build", ".",
		"--platform", fmt.Sprint(platform),
		"--build-arg", fmt.Sprintf("binary=%s", binary),
		"--provenance", "false",
		"--progress", "plain",
		"--no-cache",
	}

	if push {
		args = append(args, "--push")
	}

	for _, image := range dockerImages {
		if image.multiArch {
			args = append(args, []string{"-t", fmt.Sprintf("%s:%s", image.tag, v)}...)

			if tagLatest {
				args = append(args, []string{"-t", fmt.Sprintf("%s:%s", image.tag, "latest")}...)
			}
		}
	}

	return sh.RunV("docker", args...)
}

func Docker_noarch(push bool, tagLatest bool) error {
	//need to build for all linux because docker can run on a different machine than the go compiler
	err := Build_linux()

	if err != nil {
		return err
	}

	v, err := version()

	if err != nil {
		return err
	}

	binary := fmt.Sprintf("%s/%s_%s_%s", buildDir, binaryName, v, osLinux)

	args := []string{
		"buildx", "build", ".",
		"--build-arg", fmt.Sprintf("binary=%s", binary),
		"--provenance", "false",
		"--progress", "plain",
		"--no-cache",
	}

	if push {
		args = append(args, "--push")
	}

	for _, image := range dockerImages {
		if !image.multiArch {
			tag := fmt.Sprintf("%s:%s", image.tag, v)
			args = append(args, []string{"-t", tag}...)

			if tagLatest {
				args = append(args, []string{"-t", fmt.Sprintf("%s:%s", image.tag, "latest")}...)
			}
		}
	}

	return sh.RunV("docker", args...)
}

func targetToDockerPlatform(t target) (*dockerPlatform, error) {
	if t.Os != osLinux {
		return nil, errors.New("can only build for docker platform linux")
	}

	switch t.Arch {
	case archArm64:
		return ptr(dockerPlatformLinuxArm64), nil
	case archAmd64:
		return ptr(dockerPlatformLinuxAmd64), nil
	}

	return nil, errors.New("unsupported docker architecture")
}

func Build() error {
	_, err := buildForRunningArch()

	return err
}

func Build_all() error {
	for _, t := range supportedTargets {
		_, err := buildServerAndTools(t)

		if err != nil {
			return err
		}
	}

	return nil
}

func Build_darwin() error {
	_, err := buildForOs(osDarwin)
	return err
}

func Build_linux() error {
	_, err := buildForOs(osLinux)
	return err
}

func Build_freebsd() error {
	_, err := buildForOs(osFreeBsd)
	return err
}

func Build_windows() error {
	_, err := buildForOs(osWindows)
	return err
}

func Test() error {
	return sh.RunV("go", "test", "-v", "-coverprofile", "cover.out", "./...")
}

func Integration_tests() error {
	root, err := os.Getwd()

	if err != nil {
		return err
	}

	out, err := filepath.Abs(buildDir + "/server_integration_test")

	if err != nil {
		return err
	}

	coverageOutDir, err := filepath.Abs("./integration_tests_coverage_out")

	if err != nil {
		return err
	}

	if _, err = os.Stat(coverageOutDir); !os.IsNotExist(err) {
		err = os.RemoveAll(coverageOutDir)

		if err != nil {
			return err
		}
	}

	err = os.Mkdir(coverageOutDir, os.ModePerm)

	if err != nil {
		return err
	}

	err = os.Chdir("server/cmd/")

	if err != nil {
		return err
	}

	err = sh.RunV("go", "build", "-cover", "-o", out, "-coverpkg=all")

	if err != nil {
		return err
	}

	err = os.Chdir(root)

	if err != nil {
		return err
	}

	err = os.Chdir("integration_tests/")

	if err != nil {
		return err
	}

	err = sh.RunV("go", "run", "cmd/main.go", out, coverageOutDir)

	if err != nil {
		return err
	}

	out, err = sh.Output("go", "tool", "covdata", "percent", "-i", coverageOutDir)

	if err != nil {
		return err
	}

	lines := strings.Split(out, "\n")

	for _, l := range lines {
		if strings.Contains(l, "github.com/lucasl0st/InfiniteDB") {
			fmt.Println(l)
		}
	}

	return nil
}

func Clean() error {
	return os.RemoveAll(buildDir)
}

func version() (string, error) {
	return sh.Output("git", "describe", "--tags", "--always")
}

func buildForOs(o operating_system) ([]result, error) {
	var results []result

	for _, t := range supportedTargets {
		if t.Os != o {
			continue
		}

		r, err := buildServerAndTools(t)

		if err != nil {
			return nil, err
		}

		for _, rr := range r {
			if rr.binaryName == binaryName {
				results = append(results, rr)
			}
		}
	}

	return results, nil
}

func getRunningTarget() (*target, error) {
	var t *target = nil

	for _, supportedTarget := range supportedTargets {
		if fmt.Sprint(supportedTarget.Arch) == runtime.GOARCH && fmt.Sprint(supportedTarget.Os) == runtime.GOOS {
			t = &supportedTarget
			break
		}
	}

	if t == nil {
		return nil, errors.New("this architecture or operating system is not supported")
	}

	return t, nil
}

func buildForRunningArch() ([]result, error) {
	t, err := getRunningTarget()

	if err != nil {
		return nil, err
	}

	return buildServerAndTools(*t)
}

func buildServerAndTools(t target) ([]result, error) {
	var results []result

	r, err := buildGo(t, "./server/cmd", binaryName)

	if err != nil {
		return nil, err
	}

	results = append(results, *r)

	for _, tool := range tools {
		r, err := buildGo(t, fmt.Sprintf("./tools/%s", tool), tool)

		if err != nil {
			return nil, err
		}

		results = append(results, *r)
	}

	return results, nil
}

func buildGo(t target, src string, binaryName string) (*result, error) {
	v, err := version()

	if err != nil {
		return nil, err
	}

	out, err := filepath.Abs(fmt.Sprintf("%s/%s_%s_%s_%s", buildDir, binaryName, v, t.Os, t.Arch))

	if err != nil {
		return nil, err
	}

	root, err := os.Getwd()

	if err != nil {
		return nil, err
	}

	err = os.Chdir(src)

	if err != nil {
		return nil, err
	}

	defer os.Chdir(root)

	env := map[string]string{
		"CGO_ENABLED": fmt.Sprint("0"),
		"GOOS":        fmt.Sprint(t.Os),
		"GOARCH":      fmt.Sprint(t.Arch),
	}

	err = sh.RunWithV(env, "go", "build", "-o", out)

	if err != nil {
		return nil, err
	}

	fmt.Printf("built %s \n", out)

	return &result{
		t:          t,
		path:       out,
		binaryName: binaryName,
	}, err
}

func ptr[T any](v T) *T {
	return &v
}
