package aferoS3

import (
	"fmt"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
	"github.com/spf13/afero"
	"github.com/spf13/afero/mem"
	"io/ioutil"
	"mime"
	"os"
	"strings"
	"time"
)

/*

I think here I should create a s3 to file system mapping?

Example:

	"private":                   600,
	"public-read":               664,
	"public-read-write":         666,
	"authenticated-read":        660,
	"bucket-owner-read":         660,
	"bucket-owner-full-control": 666,

*/

type S3Fs struct {
	Bucket *s3.Bucket
	auth   *aws.Auth
}

var USEast = aws.USEast
var USGovWest = aws.USGovWest
var USWest = aws.USWest
var USWest2 = aws.USWest2
var EUWest = aws.EUWest
var EUCentral = aws.EUCentral
var APSoutheast = aws.APSoutheast
var APSoutheast2 = aws.APSoutheast2
var APNortheast = aws.APNortheast
var SAEast = aws.SAEast
var CNNorth = aws.CNNorth

func GetBucket(name string, region aws.Region) (afero.Fs, error) {
	a, err := aws.EnvAuth()
	if err != nil {
		return nil, err
	}

	client := s3.New(a, region)
	bucket := client.Bucket(name)

	return S3Fs{
		Bucket: bucket,
		auth:   &a,
	}, nil
}

func (S3Fs) Name() string {
	return "S3Fs"
}

func (S3Fs) Create(name string) (afero.File, error) {
	return mem.CreateFile(name), nil
}

// Read from s3, and bring down whole file? or torrent?
func (s S3Fs) Open(name string) (afero.File, error) {

	memFile, err := s.Create(getNameFromPath(name))
	if err != nil {
		return nil, err
	}

	torrent, err := s.Bucket.GetTorrent(name)
	if err != nil {
		return memFile, err
	}

	memFile.Write(torrent)

	return memFile, err
}

func (s S3Fs) Push(f afero.File, path string) error {

	body, err := ioutil.ReadAll(f)

	if err != nil {
		s.Bucket.Put(path, body, mime.TypeByExtension(path), "")
	}

	return err
}

func getNameFromPath(fileName string) string {
	var name string
	tokens := strings.Split(fileName, ".")
	ext := tokens[len(tokens)-1]

	if len(tokens) > 2 {
		name = strings.Join(tokens[:len(tokens)-1], ".")
	} else {
		name = tokens[0]
	}

	return fmt.Sprintf("%s.%s", name, ext)
}

type S3FileInfo struct {
	os.FileInfo
	file *afero.File
}

// Maybe different between Open is its torrent, and this is the actuall file
func (s S3Fs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	file, err := s.Open(name)
	s.Chmod(name, perm)
	return file, err
}

// Set ACL Perms
func (s S3Fs) Chmod(name string, mode os.FileMode) error {
	return nil
}

func (S3Fs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return nil
}

func (s S3Fs) Stat(name string) (os.FileInfo, error) {
	f, err := s.Open(name)
	return S3FileInfo{file: &f}, err
}

// Renames a file
func (s S3Fs) Rename(oldname, newname string) error {
	return s.Bucket.Copy(oldname, newname, s3.ACL(""))
}

// Removes a file
func (s S3Fs) Remove(name string) error {
	return s.Bucket.Del(name)
}

// Dont think we can do much here
func (S3Fs) Mkdir(name string, perm os.FileMode) error    { return nil }
func (S3Fs) MkdirAll(path string, perm os.FileMode) error { return nil }
func (S3Fs) RemoveAll(path string) error                  { return nil }
