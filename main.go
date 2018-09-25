package main

import (
	"log"
	"net/http"
	"os"
	"time"

	etcd "go.etcd.io/etcd/clientv3"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: ./etcd-snapper-b2 [snapshot_path]")
	}
	snapshotPath := os.Args[1]

	spec, err := ParseENV()
	if err != nil {
		log.Fatalf("%v", err)
	}

	b2Uploader := B2Uploader{
		ApplicationID:       spec.B2ApplicationID,
		ApplicationKey:      spec.B2ApplicationKey,
		BucketID:            spec.B2BucketID,
		Object:              spec.B2Object,
		UploadRetryInterval: spec.B2UploadRetryInterval,
		HTTPClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}

	cl, err := etcd.New(etcd.Config{
		Endpoints:   []string{spec.EtcdEndpoint},
		DialTimeout: 10 * time.Second,
	})
	if err != nil {
		log.Fatalf("Unable to connect to etcd: %v", err)
	}

	etcdSnapper := EtcdSnapper{
		Client:                             cl,
		Uploader:                           b2Uploader,
		WaitForAdditionalChangesIntervalMS: spec.WaitForChangesInterval,
		SnapshotPath:                       snapshotPath,
		Prefix:                             spec.EtcdPrefix,
	}

	log.Printf("Snapshot initial etcd snapshot")
	err = etcdSnapper.SnapshotAndUpload()
	if err != nil {
		log.Fatalf("Unable to upload initial etcd snapshot: %v", err)
	}
	log.Printf("Starting to watch etcd")
	go etcdSnapper.Watch()

}
