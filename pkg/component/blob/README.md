# blob

Cloud storage abstraction over [gocloud.dev/blob](https://gocloud.dev/howto/blob/). Provides a uniform `*blob.Bucket` API across local filesystem, in-memory, S3, GCS, and Azure Blob Storage. Provider selection is driven by URL scheme.

## Supported backends

| URL scheme | Backend |
|---|---|
| `file:///path/to/dir` | Local filesystem |
| `mem://` | In-memory (tests) |
| `s3://bucket?region=us-east-1` | AWS S3 |
| `gs://bucket` | Google Cloud Storage |
| `azblob://container` | Azure Blob Storage |

S3/GCS/Azure require importing the respective gocloud driver package.

## Configuration

```go
type BlobConfig struct {
    Filesystems map[string]FilesystemConfig `config:"filesystems"`
}

type FilesystemConfig struct {
    URL string `config:"url"`
}
```

```yaml
blob:
  filesystems:
    uploads:
      url: "s3://my-bucket?region=us-east-1"
    avatars:
      url: "file:///var/storage/avatars"
```

## Usage

```go
factory := blob.NewDefaultBlobFactory()

bucket, err := factory.Create(ctx, blob.FilesystemConfig{
    URL: "file:///tmp/storage",
})
if err != nil {
    return err
}
defer bucket.Close()

// Write
err = bucket.WriteAll(ctx, "uploads/photo.jpg", data, nil)

// Read
data, err = bucket.ReadAll(ctx, "uploads/photo.jpg")

// Delete
err = bucket.Delete(ctx, "uploads/photo.jpg")
```

## FilesystemRegistry

When multiple named buckets are needed, use the registry:

```go
registry := blob.NewFilesystemRegistry(map[string]*cloudblob.Bucket{
    "uploads": uploadsBucket,
    "avatars": avatarsBucket,
})

bucket, err := registry.Get("uploads")
```

Use `fxblob` to wire multiple buckets from config into a registry via FX.
