// +build windows

package bundlecollection_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-agent/agent/applier/bundlecollection"
	"github.com/cloudfoundry/bosh-agent/agent/applier/bundlecollection/fakes"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
)

//go:generate counterfeiter -o fakes/fake_clock.go ../../../vendor/code.cloudfoundry.org/clock Clock

var _ = Describe("FileBundle", func() {
	var (
		fs          *fakesys.FakeFileSystem
		fakeClock   *fakes.FakeClock
		logger      boshlog.Logger
		sourcePath  string
		installPath string
		enablePath  string
		fileBundle  FileBundle
	)

	BeforeEach(func() {
		fs = fakesys.NewFakeFileSystem()
		fakeClock = new(fakes.FakeClock)

		err := fs.MkdirAll("/d-drive/data", os.ModePerm)
		Expect(err).ToNot(HaveOccurred())

		err = fs.MkdirAll("/var/vcap", os.ModePerm)
		Expect(err).ToNot(HaveOccurred())

		err = fs.Symlink("/d-drive/data", "/var/vcap/data")
		Expect(err).ToNot(HaveOccurred())

		err = fs.MkdirAll("/var/vcap/data/jobs", os.ModePerm)
		Expect(err).ToNot(HaveOccurred())

		err = fs.MkdirAll("/var/vcap/jobs", os.ModePerm)
		Expect(err).ToNot(HaveOccurred())

		installPath = "/var/vcap/data/jobs/job_name"
		enablePath = "/var/vcap/jobs/job_name"
		logger = boshlog.NewLogger(boshlog.LevelNone)
		fileBundle = NewFileBundle(installPath, enablePath, os.FileMode(0750), fs, fakeClock, logger)
	})

	createSourcePath := func() string {
		path := "/source-path"
		err := fs.MkdirAll(path, os.ModePerm)
		Expect(err).ToNot(HaveOccurred())

		return path
	}

	BeforeEach(func() {
		sourcePath = createSourcePath()
	})

	Describe("Disable", func() {
		Context("where the enabled path target is the same installed version", func() {
			BeforeEach(func() {
				_, err := fileBundle.Install(sourcePath)
				Expect(err).NotTo(HaveOccurred())
				//Expect(fs.FileExists("/d-drive/data/jobs/job_name")).To(BeTrue())

				_, err = fileBundle.Enable()
				Expect(err).NotTo(HaveOccurred())

				Expect(fs.FileExists(enablePath)).To(BeTrue())
			})
			It("does not return error and removes the symlink", func() {
				err := fileBundle.Disable()
				Expect(err).NotTo(HaveOccurred())
				Expect(fs.FileExists(enablePath)).To(BeFalse())
			})
		})
	})
})
