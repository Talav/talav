# fxmedia

FX module for [pkg/component/media](../../component/media). Wires storage providers, CDN routing, image resizers, and CQRS media command/query handlers.

**Status: early stage / prototype** — API may change.

## Quick start

```go
fx.New(
    fxconfig.FxConfigModule,
    fxlogger.FxLoggerModule,
    fxorm.FxORMModule,
    fxblob.FxBlobModule,
    fxmedia.FxMediaModule,
)
```

## Configuration

```yaml
media:
  providers:
    default:
      filesystem: uploads     # name of a blob filesystem from fxblob config
      cdn: cloudfront
  cdn:
    cloudfront:
      base_url: "https://cdn.example.com"
  resizers:
    thumbnail:
      type: simple
  presets:
    avatar:
      providers: [default]
      formats:
        thumb: { width: 128, height: 128 }
```

The `blob.filesystems` section (in `fxblob` config) must define any filesystem name referenced by a provider.

## What's wired

| Type | Description |
|---|---|
| `*command.CreateMediaHandler` | Upload a media file |
| `*command.UpdateMediaHandler` | Update media metadata |
| `*command.DeleteMediaHandler` | Delete a media file |
| `*query.GetMediaQueryHandler` | Fetch single media record |
| `*query.ListMediaQueryHandler` | List media records |

## Registering custom resizers

```go
fxmedia.AsResizer("watermark", func(codec resizer.ImageCodec) resizer.Resizer {
    return myWatermarkResizer{codec: codec}
})
```

## Dependencies

- `fxblob.FxBlobModule` — blob filesystem registry
- `fxorm.FxORMModule` — database and repository registry
