package oss

import (
	"io"
	"io/ioutil"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/drone/drone-cache-lib/storage"
	"github.com/sirupsen/logrus"
)

type Options struct {
	Endpoint string
	Bucket   string
	Ak       string
	SK       string
}

type ossStorage struct {
	bucket *oss.Bucket
}

func New(opts *Options) (storage.Storage, error) {
	client, err := oss.New(opts.Endpoint, opts.Ak, opts.SK)
	if err != nil {
		return nil, err
	}
	bucket, err := client.Bucket(opts.Bucket)
	if err != nil {
		return nil, err
	}
	return &ossStorage{bucket: bucket}, nil
}

func (s *ossStorage) Get(p string, dst io.Writer) error {
	logrus.Infof("Retrieving file in %s at %s", s.bucket.BucketName, p)
	reader, err := s.bucket.GetObject(p)
	if err != nil {
		return err
	}
	defer func() {
		_ = reader.Close()
	}()

	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	if _, err := dst.Write(body); err != nil {
		return err
	}
	logrus.Infof("Downloaded cache file to server succeeded.")
	return nil
}

func (s *ossStorage) Put(p string, src io.Reader) error {

	logrus.Infof("Uploading to bucket %s at %s", s.bucket.BucketName, p)
	if err := s.bucket.PutObject(p, src); err != nil {
		logrus.Errorf("Upload cache file in %s at %s failed: %s", s.bucket.BucketName, p, err)
		return err
	}
	logrus.Infof("Uploaded cache file in %s at %s", s.bucket.BucketName, p)
	return nil
}

func (s *ossStorage) List(p string) ([]storage.FileEntry, error) {
	res, err := s.bucket.ListObjects(oss.Prefix(p), oss.Delimiter("/"))
	if err != nil {
		logrus.Errorf("List cache failed: %s", err)
		return nil, err
	}
	entries := make([]storage.FileEntry, 0)

	for _, obj := range res.Objects {
		entry := storage.FileEntry{
			Path:         obj.Key,
			Size:         obj.Size,
			LastModified: obj.LastModified,
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func (s *ossStorage) Delete(p string) error {
	logrus.Infof("Trying to delete old cache file %s in bucket %s", p, s.bucket.BucketName)
	if err := s.bucket.DeleteObject(p); err != nil {
		logrus.Errorf("Delete old cache %s failed: %s", p, err)
		return err
	}
	logrus.Infof("Deleted old cache %s", p)
	return nil
}
