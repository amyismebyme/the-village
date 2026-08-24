# Build metadata

The API build supports linker-injected metadata:

- `runtime.BuildVersion`
- `runtime.GitCommit`
- `runtime.BuildTimestamp`
- `runtime.Environment`

Docker accepts these through build arguments and injects them with Go `-ldflags -X`.

The Zscaler corporate CA is optional. Docker BuildKit can receive it as a secret:

```sh
docker build --secret id=zscaler,src=zscaler.crt .
```

When the secret is omitted, the image builds using the normal system CA bundle. The certificate is never copied into the runtime image.
