
### Prerequisites
The project uses postgres database. The postgres configuration is expected to be in working directory of the process in the `psqlInfo.js` file. The configuration should contain the following fields:

```
{
    "host": ...,
    "port": ...,
    "user": ...,
    "dbname": ...,
    "password": ...,
    "sslmode": ...
}
```

See `psqlInfo.js` for example db info.

### Run server locally

```
> go run ./main.go
```

### Deploy backend

```
> CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w -extldflags "-static"'
> ssh root@hat.adjoint.fun 'systemctl stop hatgame.service'
> scp hatgame psqlInfo.json root@hat.adjoint.fun:/var/www/
> ssh root@hat.adjoint.fun 'systemctl start hatgame.service'
> ssh root@hat.adjoint.fun 'systemctl status hatgame.service'
```