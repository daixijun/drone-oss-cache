package main

import (
	"path/filepath"
	"time"

	"github.com/drone/drone-cache-lib/archive/util"
	"github.com/drone/drone-cache-lib/cache"
	"github.com/drone/drone-cache-lib/storage"
	"github.com/sirupsen/logrus"
)

const (
	// RestoreMode for resotre mode string
	RestoreMode = "restore"
	// RebuildMode for rebuild mode string
	RebuildMode = "rebuild"
	// FlushMode for flush mode string
	FlushMode = "flush"
)

type Plugin struct {
	Rebuild      bool
	Restore      bool
	Flush        bool
	FlushAge     int
	Path         string
	FallBackPath string
	FileName     string
	Mode         string
	Mount        []string
	Storage      storage.Storage
}

func (p *Plugin) Exec() error {
	var err error

	at, err := util.FromFilename(p.FileName)
	if err != nil {
		return err
	}
	c := cache.New(p.Storage, at)

	path := filepath.Join(p.Path, p.FileName)
	if p.Mode == RebuildMode {
		logrus.Infof("Rebuilding cache at %s", path)
		err = c.Rebuild(p.Mount, path)

		if err == nil {
			logrus.Infof("Cache rebuilt")
		}
	}

	if p.Mode == RestoreMode {
		logrus.Infof("Restoring cache at %s", path)
		err = c.Restore(path, p.FallBackPath)

		if err == nil {
			logrus.Info("Cache restored")
		}
	}

	if p.Mode == FlushMode {
		logrus.Infof("Flushing cache items older than %d days at %s", p.FlushAge, path)
		f := cache.NewFlusher(p.Storage, genIsExpired(p.FlushAge))
		err = f.Flush(p.Path)

		if err == nil {
			logrus.Info("Cache flushed")
		}
	}

	return err
}

func genIsExpired(age int) cache.DirtyFunc {
	return func(file storage.FileEntry) bool {
		// Check if older than "age" days
		return file.LastModified.Before(time.Now().AddDate(0, 0, age*-1))
	}
}
