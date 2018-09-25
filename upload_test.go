package main

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type B2UploaderTestSuite struct {
	suite.Suite
	AppplicationID     string
	ApplicationKey     string
	BucketID           string
	Object             string
	UploadRetyInterval int
	TestFile           string
}

func TestB2UploaderIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(B2UploaderTestSuite))
}

func (s *B2UploaderTestSuite) SetupSuite() {
	s.AppplicationID = os.Getenv("ESB_B2_APPLICATION_ID")
	if len(s.AppplicationID) == 0 {
		s.T().Skipf("ESB_B2_APPLICATION_ID does not exists")
	}

	s.ApplicationKey = os.Getenv("ESB_B2_APPLICATION_KEY")
	if len(s.AppplicationID) == 0 {
		s.T().Skipf("ESB_B2_APPLICATION_KEY does not exists")
	}

	s.BucketID = os.Getenv("ESB_B2_BUCKET_ID")
	if len(s.BucketID) == 0 {
		s.T().Skipf("ESB_B2_BUCKET does not exists")
	}

	s.Object = os.Getenv("ESB_B2_OBJECT")
	if len(s.Object) == 0 {
		s.T().Skipf("ESB_B2_OBJECT does not exists")
	}

	wd, err := os.Getwd()
	s.Require().NoError(err)
	s.TestFile = filepath.Join(wd, "testdata", "hello.txt")

}

func (s *B2UploaderTestSuite) Test_Upload_Success() {

	client := http.Client{
		Timeout: time.Second * 10,
	}

	uploader := B2Uploader{
		ApplicationID:       s.AppplicationID,
		ApplicationKey:      s.ApplicationKey,
		BucketID:            s.BucketID,
		Object:              s.Object,
		UploadRetryInterval: 5000,
		HTTPClient:          &client,
	}

	err := uploader.Upload(context.Background(), s.TestFile)
	s.NoError(err)
}

func (s *B2UploaderTestSuite) Test_Upload_Fail() {

	client := http.Client{
		Timeout: time.Second * 10,
	}

	uploader := B2Uploader{
		ApplicationID:       s.AppplicationID,
		ApplicationKey:      s.ApplicationKey,
		BucketID:            s.BucketID,
		Object:              s.Object,
		UploadRetryInterval: 1000,
		HTTPClient:          &client,
	}

	ctx, cancel := context.WithCancel(context.Background())
	errChan := make(chan error, 1)
	go func() {
		err := uploader.Upload(ctx, "NOTAFILE")
		errChan <- err
	}()

	go func() {
		time.Sleep(time.Second)
		cancel()
	}()

	var errUpload error
	t := time.NewTimer(time.Second * 5).C
L:
	for {
		select {
		case <-t:
			s.FailNow("Timeout")
		case errUpload = <-errChan:
			break L
		}
	}

	s.Require().Error(errUpload)
	s.Equal("Upload canceled", errUpload.Error())

}
