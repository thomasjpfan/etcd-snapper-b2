package main

import (
	"context"
	"io"
	"log"
	"os"
	"sync"
	"time"

	etcd "go.etcd.io/etcd/clientv3"
)

// EtcdSnapper watches etcd and takes snapshots
type EtcdSnapper struct {
	Client                             *etcd.Client
	Uploader                           Uploader
	WaitForAdditionalChangesIntervalMS int
	SnapshotPath                       string
	Prefix                             string
	SnapLock                           sync.RWMutex
	Cancel                             context.CancelFunc
}

// Watch watches etcd cluster for changes and makes snapshots
func (e *EtcdSnapper) Watch() {
	ctx := context.Background()

	rch := e.Client.Watch(ctx, e.Prefix, etcd.WithPrefix())

	waitDuration := time.Millisecond * time.Duration(e.WaitForAdditionalChangesIntervalMS)

	for range rch {
		time.AfterFunc(waitDuration, func() {
			err := e.SnapshotAndUpload()
			if err != nil {
				log.Printf("%v", err)
				return
			}
			log.Print("Snapshot uploaded")
		})
	}
}

// SnapshotAndUpload makes a snapshot and uploads it
func (e *EtcdSnapper) SnapshotAndUpload() error {
	if e.Cancel != nil {
		e.Cancel()
	}

	e.SnapLock.Lock()
	defer e.SnapLock.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	e.Cancel = cancel

	err := e.snapshot(ctx)
	if err != nil {
		return err
	}

	return e.Uploader.Upload(ctx, e.SnapshotPath)
}

func (e *EtcdSnapper) snapshot(ctx context.Context) error {
	snapReader, err := e.Client.Snapshot(ctx)
	if err != nil {
		return err
	}
	file, err := os.Create(e.SnapshotPath)
	if err != nil {
		return err
	}
	_, err = io.Copy(file, snapReader)
	if err != nil {
		return err
	}
	return nil
}
