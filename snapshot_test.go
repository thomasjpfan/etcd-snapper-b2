package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ory/dockertest"
	dc "github.com/ory/dockertest/docker"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	etcd "go.etcd.io/etcd/clientv3"
)

type SnapshotTestSuite struct {
	suite.Suite
	Pool         *dockertest.Pool
	Resource     *dockertest.Resource
	EClient      *etcd.Client
	Endpoint     string
	TempDir      string
	SnapshotPath string
}

type MockUploader struct {
	mock.Mock
}

func (m *MockUploader) Upload(ctx context.Context, path string) error {
	args := m.Called(ctx, path)
	return args.Error(0)
}

func TestSnapshotUnitTestSuite(t *testing.T) {
	suite.Run(t, new(SnapshotTestSuite))
}

func (s *SnapshotTestSuite) SetupSuite() {
	pool, err := dockertest.NewPool("")
	s.Require().NoError(err)

	s.Pool = pool

	options := &dockertest.RunOptions{
		Repository: "gcr.io/etcd-development/etcd",
		Tag:        "v3.3.9",
		Cmd: []string{"/usr/local/bin/etcd",
			"--name", "s1", "--data-dir", "/etcd-data",
			"--listen-client-urls", "http://0.0.0.0:2379",
			"--advertise-client-urls", "http://0.0.0.0:2379",
			"--listen-peer-urls", "http://0.0.0.0:2380",
			"--initial-advertise-peer-urls", "http://0.0.0.0:2380",
			"--initial-cluster", "s1=http://0.0.0.0:2380",
			"--initial-cluster-token", "tkn",
			"--initial-cluster-state", "new"},
		PortBindings: map[dc.Port][]dc.PortBinding{
			"2379": []dc.PortBinding{{HostPort: "2379"}},
		},
	}
	resource, err := pool.RunWithOptions(options)
	s.Require().NoError(err)
	s.Resource = resource

	s.Endpoint = fmt.Sprintf("localhost:%s", resource.GetPort("2379/tcp"))

	err = pool.Retry(func() error {
		cl, err := etcd.New(etcd.Config{
			Endpoints:   []string{s.Endpoint},
			DialTimeout: 10 * time.Second,
		})
		if err != nil {
			return err
		}
		s.EClient = cl
		return nil
	})
	s.Require().NoError(err)
}

func (s *SnapshotTestSuite) TearDownSuite() {
	s.Pool.Purge(s.Resource)
	os.RemoveAll(s.TempDir)
}

func (s *SnapshotTestSuite) SetupTest() {
	dir, err := ioutil.TempDir("", "snapshot")
	s.Require().NoError(err)
	s.TempDir = dir

	s.SnapshotPath = filepath.Join(dir, "snapshot.db")
}

func (s *SnapshotTestSuite) Test_Watch_TwoEventsWithinInterval() {

	doneChan := make(chan struct{})
	uploadMockCnt := 0
	uploadMock := new(MockUploader)
	uploadMock.On("Upload", mock.AnythingOfType("*context.cancelCtx"),
		s.SnapshotPath).Run(func(args mock.Arguments) {
		uploadMockCnt++
		if uploadMockCnt == 1 {
			doneChan <- struct{}{}
		}
	}).Return(nil)

	eSnapper := EtcdSnapper{
		Client:                             s.EClient,
		Uploader:                           uploadMock,
		WaitForAdditionalChangesIntervalMS: 2000,
		SnapshotPath:                       s.SnapshotPath,
		Prefix:                             "hello",
	}

	go eSnapper.Watch()

	go func() {
		time.Sleep(1 * time.Second)
		_, err := s.EClient.Put(context.Background(), "hello/there", "world")
		if err != nil {
			return
		}
		time.Sleep(1 * time.Second)
		_, err = s.EClient.Put(context.Background(), "hello", "sings")
		if err != nil {
			return
		}
	}()

	timer := time.NewTimer(time.Second * 5).C

L:
	for {
		select {
		case <-timer:
			s.FailNow("Timeout")
		case <-doneChan:
			break L
		}
	}

	s.Equal(1, uploadMockCnt)
	_, err := os.Stat(s.SnapshotPath)
	s.True(!os.IsNotExist(err))
	uploadMock.AssertExpectations(s.T())
}

func (s *SnapshotTestSuite) Test_Watch_TwoEventsOutsideOfInterval() {
	doneChan := make(chan struct{})
	uploadMockCnt := 0
	uploadMock := new(MockUploader)
	uploadMock.On("Upload", mock.AnythingOfType("*context.cancelCtx"),
		s.SnapshotPath).Run(func(args mock.Arguments) {
		uploadMockCnt++
		if uploadMockCnt == 2 {
			doneChan <- struct{}{}
		}
	}).Return(nil)

	eSnapper := EtcdSnapper{
		Client:                             s.EClient,
		Uploader:                           uploadMock,
		WaitForAdditionalChangesIntervalMS: 500,
		SnapshotPath:                       s.SnapshotPath,
		Prefix:                             "hello",
	}

	go eSnapper.Watch()

	go func() {
		time.Sleep(1 * time.Second)
		_, err := s.EClient.Put(context.Background(), "hello/wow", "world")
		if err != nil {
			return
		}
		time.Sleep(2 * time.Second)
		_, err = s.EClient.Put(context.Background(), "hello/there", "sings")
		if err != nil {
			return
		}
	}()

	timer := time.NewTimer(time.Second * 8).C

L:
	for {
		select {
		case <-timer:
			s.FailNow("Timeout")
		case <-doneChan:
			break L
		}
	}

	s.Equal(2, uploadMockCnt)
	_, err := os.Stat(s.SnapshotPath)
	s.True(!os.IsNotExist(err))
	uploadMock.AssertExpectations(s.T())

}
