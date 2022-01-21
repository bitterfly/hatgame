## Prerequisites

See `psqlInfo.js` for expected db info.

## [How to play](go-presentation.md)

## Deploy backend

```
> CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w -extldflags "-static"'
> ssh root@hat.adjoint.fun 'systemctl stop hatgame.service'
> scp hatgame psqlInfo.json root@hat.adjoint.fun:/var/www/
> ssh root@hat.adjoint.fun 'systemctl start hatgame.service'
> ssh root@hat.adjoint.fun 'systemctl status hatgame.service'
```

## Local vs prod config

Local vs production configuration/picking is done via the `config.js` file.

This file is included in `index.html`.

By default, the local one is used, as it's expected that most of the time
a developer will locally be testing the project. If you want to deploy,
you need to make sure to copy `config_prod.js` and rename it to `config.js`.

## Deploy frontend

```
> elm make --output elm.js src/Main.elm
> scp index.html elm.js root@hat.adjoint.fun:/var/www/html
> scp config_prod.js root@hat.adjoint.fun:/var/www/html/config.js
> scp sass/output.css root@hat.adjoint.fun:/var/www/html/sass
```
