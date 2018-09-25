package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ParserTestSuite struct {
	suite.Suite
}

func TestParserUnitTestSuite(t *testing.T) {
	suite.Run(t, new(ParserTestSuite))
}

func (s *ParserTestSuite) TearDownTest() {
	os.Clearenv()
}

func (s *ParserTestSuite) Test_Parser() {

	etcdEndpoint := "etcd:2379"
	etcdPrefix := "traefik"
	b2AppID := "backblazeid"
	b2AppKey := "backblazekey"
	b2BucketID := "backblazebucket"
	b2Object := "backblazeobject"
	waitForChangesInterval := 1000
	b2UploadRetryInterval := 5000

	os.Setenv("ESB_ETCD_ENDPOINT", etcdEndpoint)
	os.Setenv("ESB_ETCD_PREFIX", etcdPrefix)
	os.Setenv("ESB_B2_APPLICATION_ID", b2AppID)
	os.Setenv("ESB_B2_APPLICATION_KEY", b2AppKey)
	os.Setenv("ESB_B2_BUCKET_ID", b2BucketID)
	os.Setenv("ESB_B2_OBJECT", b2Object)
	os.Setenv("ESB_WAIT_FOR_CHANGES_INTERVAL",
		fmt.Sprintf("%d", waitForChangesInterval))
	os.Setenv("ESB_B2_UPLOAD_RETRY_INTERVAL",
		fmt.Sprintf("%d", b2UploadRetryInterval))

	spec, err := ParseENV()
	s.Require().NoError(err)

	s.Equal(etcdEndpoint, spec.EtcdEndpoint)
	s.Equal(etcdPrefix, spec.EtcdPrefix)
	s.Equal(b2AppID, spec.B2ApplicationID)
	s.Equal(b2AppKey, spec.B2ApplicationKey)
	s.Equal(b2BucketID, spec.B2BucketID)
	s.Equal(b2Object, spec.B2Object)
	s.Equal(waitForChangesInterval, spec.WaitForChangesInterval)
	s.Equal(b2UploadRetryInterval, spec.B2UploadRetryInterval)
}
