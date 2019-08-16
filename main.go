package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/drone/drone-cache-lib/storage"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/daixijun/drone-oss-cache/storage/oss"
)

var (
	version      = "unknown"
	allowFormats = []string{"tgz", "tar", "tar.gz"}
)

func main() {
	app := cli.NewApp()
	app.Name = "drone oss cache plugin"
	app.Usage = "cache plugin"
	app.Version = version
	app.Action = run
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "filename",
			Usage:  "Filename for the cache",
			EnvVar: "PLUGIN_FILENAME",
		},
		cli.StringFlag{
			Name:   "path",
			Usage:  "path",
			EnvVar: "PLUGIN_PATH",
		},
		cli.StringFlag{
			Name:   "fallback_path",
			Usage:  "fallback_path",
			EnvVar: "PLUGIN_FALLBACK_PATH",
		},
		cli.StringFlag{
			Name:   "format",
			Usage:  "the cache archive format",
			EnvVar: "PLUGIN_FORMAT",
		},
		cli.StringSliceFlag{
			Name:   "mount",
			Usage:  "cache directories",
			EnvVar: "PLUGIN_MOUNT",
		},
		cli.BoolFlag{
			Name:   "rebuild",
			Usage:  "rebuild the cache directories",
			EnvVar: "PLUGIN_REBUILD",
		},
		cli.BoolFlag{
			Name:   "restore",
			Usage:  "restore the cache directories",
			EnvVar: "PLUGIN_RESTORE",
		},
		cli.BoolFlag{
			Name:   "flush",
			Usage:  "flush the cache",
			EnvVar: "PLUGIN_FLUSH",
		},
		cli.StringFlag{
			Name:   "flush_age",
			Usage:  "flush cache files older than # days",
			EnvVar: "PLUGIN_FLUSH_AGE",
			Value:  "30",
		},
		cli.BoolFlag{
			Name:   "debug",
			Usage:  "debug plugin output",
			EnvVar: "PLUGIN_DEBUG",
		},
		cli.StringFlag{
			Name:   "endpoint",
			Usage:  "Aliyun oss endpoint",
			EnvVar: "PLUGIN_ENDPOINT",
		},
		cli.StringFlag{
			Name:   "access-key-id",
			Usage:  "Aliyun access key id",
			EnvVar: "PLUGIN_ACCESS_KEY_ID",
		},
		cli.StringFlag{
			Name:   "access-key-secret",
			Usage:  "Alyun access key secret",
			EnvVar: "PLUGIN_ACCESS_KEY_SECRET",
		},
		cli.StringFlag{
			Name:   "bucket",
			Usage:  "Aliyun oss bucket name",
			EnvVar: "PLUGIN_BUCKET",
		},
		// Build information (for setting defaults)

		cli.StringFlag{
			Name:   "repo.owner",
			Usage:  "repository owner",
			EnvVar: "DRONE_REPO_OWNER",
		},
		cli.StringFlag{
			Name:   "repo.name",
			Usage:  "repository name",
			EnvVar: "DRONE_REPO_NAME",
		},
		cli.StringFlag{
			Name:   "repo.branch",
			Value:  "master",
			Usage:  "repository default branch",
			EnvVar: "DRONE_REPO_BRANCH",
		},
		cli.StringFlag{
			Name:   "commit.branch",
			Value:  "master",
			Usage:  "git commit branch",
			EnvVar: "DRONE_COMMIT_BRANCH",
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func run(c *cli.Context) error {
	if c.Bool("debug") {
		logrus.SetLevel(logrus.DebugLevel)
	}

	rebuild := c.Bool("rebuild")
	restore := c.Bool("restore")
	flush := c.Bool("flush")

	if rebuild && restore {
		return errors.New("Must use a single mode: rebuild, restore")
	} else if !restore && !rebuild && !flush {
		return errors.New("No action specified")
	}

	var mode string
	var mount []string

	if rebuild {
		// Look for the mount points to rebuild
		mount = c.StringSlice("mount")

		if len(mount) == 0 {
			return errors.New("No mounts specified")
		}

		mode = RebuildMode
	} else if flush {
		mode = FlushMode
	} else {
		mode = RestoreMode
	}

	path := c.GlobalString("path")

	if len(path) == 0 {
		logrus.Info("No path specified. Creating default")
		path = fmt.Sprintf("%s/%s", c.String("repo.owner"), c.String("repo.name"))
	}

	fallbackpath := c.GlobalString("fallback_path")
	if len(fallbackpath) == 0 {
		logrus.Info("No fallback_path specified. Creating default")
		fallbackpath = path
	}

	filename := c.GlobalString("filename")
	format := c.GlobalString("format")

	if len(format) == 0 {
		format = "tgz"
	}
	if !inAllowFormat(format) {
		return fmt.Errorf("Unkown format: %s, olny tgz,tar.gz,tar are allowed.", format)
	}
	logrus.Infof("Set archive format to %s", format)

	if len(filename) == 0 {
		switch format {
		case "tgz", "tar.gz":
			filename = "archive.tgz"
		default:
			filename = "archive.tar"
		}
	}

	flushAge, err := strconv.Atoi(c.String("flush_age"))
	if err != nil {
		return err
	}

	s, err := ossStorage(c)

	if err != nil {
		return err
	}
	p := &Plugin{
		Rebuild:      rebuild,
		Restore:      restore,
		Mount:        mount,
		FileName:     filename,
		Path:         path,
		FallBackPath: fallbackpath,
		Flush:        flush,
		FlushAge:     flushAge,
		Storage:      s,
		Mode:         mode,
	}

	return p.Exec()
}

func ossStorage(c *cli.Context) (storage.Storage, error) {

	endpoint := c.GlobalString("endpoint")
	if len(endpoint) == 0 {
		endpoint = "oss-cn-hangzhou.aliyuncs.com"
	}

	bucket := c.GlobalString("bucket")
	if len(bucket) == 0 {
		return nil, errors.New("bucket must be set")
	}
	accessKeyId := c.GlobalString("access-key-id")
	accessKeySecret := c.GlobalString("access-key-secret")
	if accessKeyId == "" || accessKeySecret == "" {
		return nil, errors.New("access-key-id or access-key-secret are not be empty")
	}

	return oss.New(&oss.Options{
		Endpoint: endpoint,
		Bucket:   bucket,
		Ak:       accessKeyId,
		SK:       accessKeySecret,
	})
}

func inAllowFormat(format string) bool {
	for _, item := range allowFormats {
		if item == format {
			return true
		}
	}
	return false
}
