## Prerequisites

See `psqlInfo.js` for expected db info.

## Deploy

```
> CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w -extldflags "-static"'
> ssh root@hat.adjoint.fun 'systemctl stop hatgame.service'
> scp hatgame psqlInfo.json root@hat.adjoint.fun:/var/www/
> ssh root@hat.adjoint.fun 'systemctl start hatgame.service'
> ssh root@hat.adjoint.fun 'systemctl status hatgame.service'
```