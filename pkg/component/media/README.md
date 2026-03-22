# media

Media management component — **early stage / prototype**.

Handles media uploads, storage routing, CDN URL generation, and image preset generation (resize/crop). Storage is backed by `pkg/component/blob`; image resizing is not yet implemented.

## Configuration

```go
type MediaConfig struct {
    Resizers  map[string]ResizerConfig           `config:"resizers"`
    CDN       map[string]cdn.CDNSpec             `config:"cdn"`
    Providers map[string]provider.ProviderConfig `config:"providers"`
    Presets   map[string]PresetConfig            `config:"presets"`
}
```

```yaml
media:
  providers:
    s3:
      type: s3
      bucket: my-media-bucket
  cdn:
    cloudfront:
      base_url: "https://cdn.example.com"
  resizers:
    thumbnail:
      type: simple
  presets:
    avatar:
      providers: [s3]
      formats:
        thumb: { width: 128, height: 128 }
```

## Status

- Storage (upload/read/delete via blob provider): working
- CDN URL generation: working
- Image resizing/cropping: not implemented

HTTP handlers are in `pkg/module/media`. Use `fxmedia.FxMediaModule` to wire into FX.
