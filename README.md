# Go Starter

A Gin sample application for [Kamui Platform](https://kamui-platform.com)

Please refer to the [documentation for Go app](https://docs.kamui-platform.com/gui/apps/go.html).

## Local Development

```bash
go build -o out/app ./cmd/server
go build -o out/db ./cmd/db
out/db
out/app
```

Open http://localhost:8000 in your browser.

## Deploy to Kamui Platform

### Dashboard Settings

| Setting | Value |
|---------|-------|
| Setup Command | `go build -trimpath -ldflags='-s -w' -o out/app ./cmd/server;go build -trimpath -ldflags='-s -w' -o out/db ./cmd/db` |
| Pre-deploy Command | `out/db` |
| Start Command | `out/app` |
| Health Check Path | `/health` |

### Environment Variables

`PORT` is automatically set when you deploy app. <br>
`DATABASE_URL` is automatically set when you link a database to your app.

## Links

- [Kamui Platform](https://kamui-platform.com/)
