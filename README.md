# etcd snapper backblaze b2

Watches etcd for changes and uploads snapshots to backblaze

## Environment variables

- `ESB_ETCD_ENDPOINT`: etcd endpoint
- `ESB_ETCD_PREFIX`: etcd prefix to watch
- `ESB_B2_APPLICATION_ID`: backblaze application id
- `ESB_B2_APPLICATION_KEY`: backblaze application key
- `ESB_B2_BUCKET_ID`: bucket id to store file in
- `ESB_B2_OBJECT`: name of object on backblaze
- `ESB_WAIT_FOR_CHANGES_INTERVAL`: interval to wait between etcd changes before creating a snapshot (in ms)
- `ESB_B2_UPLOAD_RETRY_INTERVAL`: interval to retry when upload fails (in ms)
