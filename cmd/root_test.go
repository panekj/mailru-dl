package cmd

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testJob struct {
	dir  string
	file string
}

var (
	wait     = time.Second * 15
	workdir  = "."
	loglevel = "debug"
)

func TestDownload(t *testing.T) {
	paths := []testJob{
		{filepath.Join("linux", "alpine", "extended"), "alpine-extended-3.13.5-x86_64.iso"},
		{filepath.Join("linux", "alpine", "extended"), "alpine-extended-3.14.0-x86_64.iso"},
		{filepath.Join("linux", "alpine", "standard"), "alpine-standard-3.13.5-x86_64.iso"},
		{filepath.Join("linux", "alpine", "standard"), "alpine-standard-3.14.0-x86_64.iso"},
		{filepath.Join("linux", "alpine", "virt"), "alpine-virt-3.14.0_rc2-aarch64.iso"},
		{filepath.Join("linux", "debian"), "debian-10.10.0-amd64-netinst.iso"},
		{"standard", "alpine-standard-3.13.5-x86_64.iso"},
		{"standard", "alpine-standard-3.14.0-x86_64.iso"},
	}

	conf = &flags{
		wait:     wait,
		workDir:  workdir,
		logLevel: loglevel,
		prefix:   false,
		retry:    true,
	}
	down([]string{"https://cloud.mail.ru/public/RgA6/8FEhtCsn6", "https://cloud.mail.ru/public/RgA6/8FEhtCsn6/alpine/standard/"})

	for _, p := range paths {
		assert.Equal(t,
			&flags{
				wait:     wait,
				workDir:  workdir,
				logLevel: loglevel,
				prefix:   false,
				retry:    true,
			}, conf)
		assert.DirExistsf(t,
			filepath.Join(conf.workDir, p.dir),
			"Failed to find directory: '%s'!",
			filepath.Join(conf.workDir, p.dir),
		)
		assert.FileExistsf(t,
			filepath.Join(conf.workDir, p.dir, p.file),
			"Failed to find file: %s",
			filepath.Join(conf.workDir, p.dir, p.file),
		)
	}

	conf = &flags{
		wait:     wait,
		workDir:  workdir,
		logLevel: loglevel,
		prefix:   true,
		retry:    true,
	}
	down([]string{"https://cloud.mail.ru/public/RgA6/8FEhtCsn6", "https://cloud.mail.ru/public/RgA6/8FEhtCsn6/alpine/standard/"})

	for _, p := range paths {
		assert.Equal(t,
			&flags{
				wait:     wait,
				workDir:  workdir,
				logLevel: loglevel,
				prefix:   true,
				retry:    true,
			}, conf)
		assert.DirExistsf(t,
			filepath.Join(conf.workDir, "RgA6", "8FEhtCsn6", p.dir),
			"Failed to find directory: '%s'!",
			filepath.Join(conf.workDir, "RgA6", "8FEhtCsn6", p.dir),
		)
		assert.FileExistsf(t,
			filepath.Join(conf.workDir, "RgA6", "8FEhtCsn6", p.dir, p.file),
			"Failed to find file: %s",
			filepath.Join(conf.workDir, "RgA6", "8FEhtCsn6", p.dir, p.file),
		)
	}
}
