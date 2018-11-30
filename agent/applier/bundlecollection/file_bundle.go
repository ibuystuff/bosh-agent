package bundlecollection

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/clock"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

const (
	fileBundleLogTag = "FileBundle"
)

type FileBundle struct {
	installPath  string
	enablePath   string
	fileMode     os.FileMode
	fs           boshsys.FileSystem
	timeProvider clock.Clock
	logger       boshlog.Logger
}

func NewFileBundle(
	installPath, enablePath string,
	fileMode os.FileMode,
	fs boshsys.FileSystem,
	timeProvider clock.Clock,
	logger boshlog.Logger,
) FileBundle {
	return FileBundle{
		installPath:  installPath,
		enablePath:   enablePath,
		fileMode:     fileMode,
		fs:           fs,
		timeProvider: timeProvider,
		logger:       logger,
	}
}

func (b FileBundle) InstallWithoutContents() (string, error) {
	b.logger.Debug(fileBundleLogTag, "Installing without contents %v", b)

	// MkdirAll MUST be the last possibly-failing operation
	// because IsInstalled() relies on installPath presence.
	err := b.fs.MkdirAll(b.installPath, b.fileMode)
	if err != nil {
		return "", bosherr.WrapError(err, "Creating installation directory")
	}
	err = b.fs.Chown(b.installPath, "root:vcap")
	if err != nil {
		return "", bosherr.WrapError(err, "Setting ownership on installation directory")
	}

	return b.installPath, nil
}

func (b FileBundle) GetInstallPath() (string, error) {
	path := b.installPath
	if !b.fs.FileExists(path) {
		return "", bosherr.Error("install dir does not exist")
	}

	return path, nil
}

func (b FileBundle) IsInstalled() (bool, error) {
	return b.fs.FileExists(b.installPath), nil
}

func (b FileBundle) Enable() (string, error) {
	b.logger.Debug(fileBundleLogTag, "Enabling %v", b)

	if !b.fs.FileExists(b.installPath) {
		return "", bosherr.Error("bundle must be installed")
	}

	err := b.fs.MkdirAll(filepath.Dir(b.enablePath), b.fileMode)
	if err != nil {
		return "", bosherr.WrapError(err, "failed to create enable dir")
	}

	err = b.fs.Chown(filepath.Dir(b.enablePath), "root:vcap")
	if err != nil {
		return "", bosherr.WrapError(err, "Setting ownership on source directory")
	}

	err = b.fs.Symlink(b.installPath, b.enablePath)
	if err != nil {
		return "", bosherr.WrapError(err, "failed to enable")
	}

	return b.enablePath, nil
}

func (b FileBundle) Disable() error {
	b.logger.Debug(fileBundleLogTag, "Disabling %v", b)
	fmt.Fprintf(os.Stdout, "Disabling %v\n", b)

	//TODO: Replace ReadAndFollowLink with ReadLink as ReadLink doesn't seem to recursively follow SymLinks
	target, err := b.fs.Readlink(b.enablePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return bosherr.WrapError(err, "Reading symlink")
	}

	fmt.Fprintf(os.Stdout, "enabledPath Target %s\n", target)

	installPath := b.installPath
	if strings.HasPrefix(installPath, "/") {
		installPath, err = filepath.Abs(installPath)
		fmt.Fprintf(os.Stdout, "installPath Abs %s\n", installPath)
		if err != nil {
			return bosherr.WrapError(err, "Failed to convert install path to native OS path")
		}
		//installPath, err = b.fs.ReadAndFollowLink(installPath)
		fmt.Fprintf(os.Stdout, "installPath ReadAndFollowLink %s\n", installPath)
	}

	b.logger.Debug(fileBundleLogTag, "Comparing path %s to %s", installPath, target)
	fmt.Fprintf(os.Stdout, "Comparing path %s to %s\n", installPath, target)

	if target == installPath {
		return b.fs.RemoveAll(b.enablePath)
	}

	return nil
}
