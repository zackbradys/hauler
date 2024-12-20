package cosign

import (
	"bufio"
	"context"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"oras.land/oras-go/pkg/content"

	"hauler.dev/go/hauler/internal/flags"
	"hauler.dev/go/hauler/pkg/artifacts/image"
	"hauler.dev/go/hauler/pkg/consts"
	"hauler.dev/go/hauler/pkg/log"
	"hauler.dev/go/hauler/pkg/store"
)

// VerifyFileSignature verifies the digital signature of a file using Sigstore/Cosign.
func VerifySignature(ctx context.Context, s *store.Layout, keyPath string, ref string, rso *flags.StoreRootOpts, ro *flags.CliRootOpts) error {
	operation := func() error {
		cosignBinaryPath, err := getCosignPath(ro.HaulerDir)
		if err != nil {
			return err
		}

		cmd := exec.Command(cosignBinaryPath, "verify", "--insecure-ignore-tlog", "--key", keyPath, ref)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("error verifying signature: %v\n%s", err, output)
		}

		return nil
	}

	return RetryOperation(ctx, rso, ro, operation)
}

// SaveImage saves image and any signatures/attestations to the store.
func SaveImage(ctx context.Context, s *store.Layout, ref string, platform string, rso *flags.StoreRootOpts, ro *flags.CliRootOpts) error {
	l := log.FromContext(ctx)

	if !ro.IgnoreErrors {
		envVar := os.Getenv(consts.HaulerIgnoreErrors)
		if envVar == "true" {
			ro.IgnoreErrors = true
		}
	}

	operation := func() error {
		cosignBinaryPath, err := getCosignPath(ro.HaulerDir)
		if err != nil {
			return err
		}

		// check to see if the image is multi-arch
		isMultiArch, err := image.IsMultiArchImage(ref)
		if err != nil {
			return err
		}
		l.Debugf("multi-arch image [%v]", isMultiArch)

		cmd := exec.Command(cosignBinaryPath, "save", ref, "--dir", s.Root)
		// Conditionally add platform.
		if platform != "" && isMultiArch {
			l.Debugf("platform for image [%s]", platform)
			cmd.Args = append(cmd.Args, "--platform", platform)
		}

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return err
		}
		// start the command after having set up the pipe
		if err := cmd.Start(); err != nil {
			return err
		}

		// read command's stdout line by line
		output := bufio.NewScanner(stdout)
		for output.Scan() {
			l.Debugf(output.Text()) // write each line to your log, or anything you need
		}
		if err := output.Err(); err != nil {
			cmd.Wait()
			return err
		}

		// read command's stderr line by line
		errors := bufio.NewScanner(stderr)
		for errors.Scan() {
			if ro.IgnoreErrors {
				l.Warnf(errors.Text())
			}
			l.Errorf(errors.Text())
		}
		if err := errors.Err(); err != nil {
			cmd.Wait()
			return err
		}

		// Wait for the command to finish
		err = cmd.Wait()
		if err != nil {
			return err
		}

		return nil
	}

	return RetryOperation(ctx, rso, ro, operation)
}

// LoadImage loads store to a remote registry.
func LoadImages(ctx context.Context, s *store.Layout, registry string, ropts content.RegistryOptions, ro *flags.CliRootOpts) error {
	l := log.FromContext(ctx)

	cosignBinaryPath, err := getCosignPath(ro.HaulerDir)
	if err != nil {
		return err
	}

	cmd := exec.Command(cosignBinaryPath, "load", "--registry", registry, "--dir", s.Root)

	// Conditionally add extra registry flags.
	if ropts.Insecure {
		cmd.Args = append(cmd.Args, "--allow-insecure-registry=true")
	}
	if ropts.PlainHTTP {
		cmd.Args = append(cmd.Args, "--allow-http-registry=true")
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	// start the command after having set up the pipe
	if err := cmd.Start(); err != nil {
		return err
	}

	// read command's stdout line by line
	output := bufio.NewScanner(stdout)
	for output.Scan() {
		l.Infof(output.Text()) // write each line to your log, or anything you need
	}
	if err := output.Err(); err != nil {
		cmd.Wait()
		return err
	}

	// read command's stderr line by line
	errors := bufio.NewScanner(stderr)
	for errors.Scan() {
		l.Errorf(errors.Text()) // write each line to your log, or anything you need
	}
	if err := errors.Err(); err != nil {
		cmd.Wait()
		return err
	}

	// Wait for the command to finish
	err = cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}

// RegistryLogin - performs cosign login
func RegistryLogin(ctx context.Context, s *store.Layout, registry string, ropts content.RegistryOptions, ro *flags.CliRootOpts) error {
	log := log.FromContext(ctx)
	cosignBinaryPath, err := getCosignPath(ro.HaulerDir)
	if err != nil {
		return err
	}

	cmd := exec.Command(cosignBinaryPath, "login", registry, "-u", ropts.Username, "-p", ropts.Password)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error logging into registry: %v, output: %s", err, output)
	}
	log.Infof(strings.Trim(string(output), "\n"))

	return nil
}

func RetryOperation(ctx context.Context, rso *flags.StoreRootOpts, ro *flags.CliRootOpts, operation func() error) error {
	l := log.FromContext(ctx)

	if !ro.IgnoreErrors {
		envVar := os.Getenv(consts.HaulerIgnoreErrors)
		if envVar == "true" {
			ro.IgnoreErrors = true
		}
	}

	// Validate retries and fall back to a default
	retries := rso.Retries
	if retries <= 0 {
		retries = consts.DefaultRetries
	}

	for attempt := 1; attempt <= rso.Retries; attempt++ {
		err := operation()
		if err == nil {
			// If the operation succeeds, return nil (no error)
			return nil
		}

		if ro.IgnoreErrors {
			l.Warnf("warning (attempt %d/%d)... %v", attempt, rso.Retries, err)
		} else {
			l.Errorf("error (attempt %d/%d)... %v", attempt, rso.Retries, err)
		}

		// If this is not the last attempt, wait before retrying
		if attempt < rso.Retries {
			time.Sleep(time.Second * consts.RetriesInterval)
		}
	}

	// If all attempts fail, return an error
	return fmt.Errorf("operation unsuccessful after %d attempts", rso.Retries)
}

func EnsureBinaryExists(ctx context.Context, bin embed.FS, ro *flags.CliRootOpts) error {
	l := log.FromContext(ctx)

	// Set up a path for the binary to be copied
	binaryPath, err := getCosignPath(ro.HaulerDir)
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}

	l.Debugf("using hauler directory at [%s]", filepath.Dir(binaryPath))

	// Determine the architecture so that we pull the correct embedded binary
	arch := runtime.GOARCH
	rOS := runtime.GOOS
	binaryName := "cosign"
	if rOS == "windows" {
		binaryName = fmt.Sprintf("cosign-%s-%s.exe", rOS, arch)
	} else {
		binaryName = fmt.Sprintf("cosign-%s-%s", rOS, arch)
	}

	// retrieve the embedded binary
	f, err := bin.ReadFile(fmt.Sprintf("binaries/%s", binaryName))
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}

	// write the binary to the filesystem
	err = os.WriteFile(binaryPath, f, 0755)
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}

	return nil
}

// getCosignPath returns the binary path
func getCosignPath(haulerDir string) (string, error) {

	if haulerDir == "" {
		haulerDir = os.Getenv(consts.HaulerDir)
	}

	if haulerDir == "" {
		// Get the current user's information
		currentUser, err := user.Current()
		if err != nil {
			return "", fmt.Errorf("error retrieving user information: %v", err)
		}

		// Get the current user's home directory
		homeDir := currentUser.HomeDir
		haulerDir = filepath.Join(homeDir, consts.DefaultHaulerDirName)
	}

	// Create the .hauler directory (if it doesn't exist)
	if _, err := os.Stat(haulerDir); os.IsNotExist(err) {
		if err := os.MkdirAll(haulerDir, 0755); err != nil {
			return "", fmt.Errorf("error creating %s directory: %v", consts.DefaultHaulerDirName, err)
		}
	}

	// Determine the binary name
	rOS := runtime.GOOS
	binaryName := "cosign"
	if rOS == "windows" {
		binaryName = "cosign.exe"
	}

	// Construct the path to the binary
	binaryPath := filepath.Join(haulerDir, binaryName)

	return binaryPath, nil
}
