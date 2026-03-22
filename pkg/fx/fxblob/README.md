# fxblob

FX module for [pkg/component/blob](../../component/blob). Builds a `*blob.FilesystemRegistry` from config and/or programmatically registered buckets.

## Quick start

```go
fx.New(
    fxconfig.FxConfigModule,
    fxblob.FxBlobModule,
    fx.Invoke(func(registry *blob.FilesystemRegistry) {
        bucket, _ := registry.Get("uploads")
        _ = bucket.WriteAll(ctx, "file.txt", data, nil)
    }),
)
```

## Configuration

```yaml
blob:
  filesystems:
    uploads:
      url: "s3://my-bucket?region=us-east-1"
    avatars:
      url: "file:///var/storage/avatars"
```

All filesystems listed under `blob.filesystems` are opened at startup and added to the registry.

## Registering additional buckets programmatically

```go
fxblob.AsFilesystem("custom", func(ctx context.Context) (*cloudblob.Bucket, error) {
    return cloudblob.OpenBucket(ctx, "mem://")
})
```

Programmatically registered buckets take precedence over config if names collide.

## Injected types

- `*blob.FilesystemRegistry` — look up buckets by name with `registry.Get("name")`
