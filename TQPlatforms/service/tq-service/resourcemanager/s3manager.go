package resourcemanager

import (
	"fmt"
	"os"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
  util "server/service/tq-service/util"
  env  "server/service/tq-service/env"
)

type S3Manager struct {
	access_key string
	secret_key string
	end_point  string
  default_bucket string
  sess *session.Session
}

func (sm *S3Manager) Init() {

	sm.access_key = env.S3_ACCESSKEY      // admin
	sm.secret_key = env.S3_SECRETKEY      // secret
	sm.end_point = env.S3_ENDPOINT        // addr
  sm.default_bucket = env.S3_BUCKET     // bucket

	sess,_ := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(sm.access_key, sm.secret_key, ""),
		Endpoint:         aws.String(sm.end_point),
		Region:           aws.String("cn-north-1"),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(false),
	})

  sm.sess = sess
}

//showallbuket
func (sm *S3Manager) ListBucket() {

	svc := s3.New(sm.sess)
	result, err := svc.ListBuckets(nil)
	if err != nil {
		util.Infof("Unable to list buckets, %v", err)
	}

	fmt.Println("Buckets:")

	for _, b := range result.Buckets {
		fmt.Printf("* %s created on %s\n", aws.StringValue(b.Name), aws.TimeValue(b.CreationDate))
	}
	for _, b := range result.Buckets {
		fmt.Printf("%s\n", aws.StringValue(b.Name))
	}
}

// showall
func (sm *S3Manager) ListbucketFile(bucket string) {

	svc := s3.New(sm.sess)

	params := &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
	}

	result, err := svc.ListObjects(params)

	if err != nil {
		util.Infof("Unable to list items in bucket %q, %v", bucket, err)
	}

	for _, item := range result.Contents {
		fmt.Println("Name:         ", *item.Key)
		fmt.Println("List modified:", *item.LastModified)
		fmt.Println("Size:         ", *item.Size)
		fmt.Println("Storage class:", *item.StorageClass)
		fmt.Println("")
	}
}

// createbucket
func (sm *S3Manager) CreateBucketFile(bucket string) bool {

	svc := s3.New(sm.sess)

	params := &s3.CreateBucketInput{Bucket: aws.String(bucket)}
	_, err := svc.CreateBucket(params)

	if err != nil {
		util.Infof("Unable to create bucket %q, %v", bucket, err)
		return false
	}

	// Wait until bucket is created before finishing
	fmt.Printf("Waiting for bucket %q to be created...\n", bucket)

	err = svc.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})

	if err != nil {
		util.Infof("Error occurred while waiting for bucket to be created, %v", bucket)
		return false
	}

	fmt.Printf("Bucket %q successfully created\n", bucket)
	return true
}

// uploaddefault
func (sm *S3Manager) UploadDefaultbucket(filename string) {

	file, err := os.Open(filename)
	if err != nil {
		util.Infof("Unable to open file %v, %v",filename, err)
	}

	defer file.Close()

	uploader := s3manager.NewUploader(sm.sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(sm.default_bucket),
		Key:    aws.String(filename),
		Body:   file,
	})
	if err != nil {
		util.Infof("Unable to upload %q to %v, %v", filename, sm.default_bucket, err)
	}

	util.Infof("Successfully uploaded %v to %v", filename, sm.default_bucket)
}

// upload
func (sm *S3Manager) Upload(bucket string, filename string) {

	file, err := os.Open(filename)
	if err != nil {
		util.Infof("Unable to open file %v, %v",filename, err)
	}

	defer file.Close()

	uploader := s3manager.NewUploader(sm.sess)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filename),
		Body:   file,
	})
	if err != nil {
		// Print the error and exit.
		util.Infof("Unable to upload %q to %v, %v", filename, bucket, err)
	}

	util.Infof("Successfully uploaded %v to %v", filename, bucket)
}

// download
func (sm *S3Manager) Download(bucket string, filename string) {

	file, err := os.Create(filename)
	if err != nil {
		util.Infof("Unable to open file %q, %v", err)
	}

	defer file.Close()
	downloader := s3manager.NewDownloader(sm.sess)

	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(filename),
		})
	if err != nil {
		util.Infof("Unable to download item %v, %v", filename, err)
	}

	fmt.Println("Downloaded", file.Name(), numBytes, "bytes")

}

// 
