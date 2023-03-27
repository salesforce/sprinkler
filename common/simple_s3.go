package common

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Basics encapsulates the Amazon Simple Storage Service (Amazon S3) actions
// used in the examples.
// It contains S3Client, an Amazon S3 service client that is used to perform bucket
// and object actions.
type S3Basics struct {
	S3Client *s3.Client
}

type S3BucketPath struct {
	Bucket string
	Path   string
}

func DefaultS3Client() (*s3.Client, error) {
	sdkConfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("Couldn't load default configuration; check AWS account setup.  Error: %w\n", err)
	}
	s3Client := s3.NewFromConfig(sdkConfig)
	return s3Client, nil
}

// ListObjects lists the objects in a bucket.
func (b S3Basics) ListObjects(bucketName string, prefix string) ([]types.Object, error) {
	result, err := b.S3Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(prefix),
	})
	var contents []types.Object
	if err != nil {
		return contents, fmt.Errorf("Couldn't list objects in bucket %v. Error: %v\n", bucketName, err)
	}
	contents = result.Contents
	return contents, nil
}

// DownloadFile gets an object from a bucket and stores it in a local file.
func (b S3Basics) DownloadFile(bucketName string, objectKey string, fileName string) error {
	result, err := b.S3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return fmt.Errorf("Couldn't get object %v:%v.  Error:%w\n", bucketName, objectKey, err)
	}
	defer result.Body.Close()
	if err = os.MkdirAll(filepath.Dir(fileName), 0770); err != nil {
		return err
	}
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("Couldn't create file %v. Error: %w\n", fileName, err)
	}
	defer file.Close()
	body, err := io.ReadAll(result.Body)
	if err != nil {
		return fmt.Errorf("Couldn't read object body from %v. Error: %w\n", objectKey, err)
	}
	_, err = file.Write(body)
	if err != nil {
		return fmt.Errorf("Couldn't write file.  Error: %w\n", err)
	}
	return nil
}

// parse S3 url to get bucket and path
func (b S3Basics) GetBucketPath(s3Url string) (S3BucketPath, error) {
	patt, err := regexp.Compile(`^s3://([^ /]+)/([^ ]+)$`)
	if err != nil {
		return S3BucketPath{"", ""}, fmt.Errorf("regex with s3Url %v.  Error: %w\n", s3Url, err)
	}
	res := patt.FindStringSubmatch(s3Url)
	if len(res) != 3 {
		return S3BucketPath{"", ""}, fmt.Errorf("artifact parsing failed")
	}
	return S3BucketPath{res[1], res[2]}, nil
}

func (b S3Basics) GetLastSegment(path string) (string, error) {
	if !strings.Contains(path, "/") {
		return path, nil
	}
	p, err := regexp.Compile(`^.*\/([^ ]+)$`)
	if err != nil {
		return "", fmt.Errorf("regex with path %v. Error: %w\n", path, err)
	}
	res := p.FindStringSubmatch(path)
	if len(res) != 2 {
		return "", fmt.Errorf("s3 path parse failed:%v", path)
	}
	return res[1], nil
}
