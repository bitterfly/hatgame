### Local vs prod config

Local vs production configuration/picking is done via the `config.js` file.

This file is included in `index.html`.

By default, the local one is used, as it's expected that most of the time
a developer will locally be testing the project. If you want to deploy,
you need to make sure to copy `config_prod.js` and rename it to `config.js`.

### Run locally

The server is expected to be running on `localhost:8080`.

```
> elm make --output elm.js src/Main.elm
```
Open `index.html` in a browser.

### Deploy frontend

```
> elm make --output elm.js src/Main.elm
> scp index.html elm.js root@hat.adjoint.fun:/var/www/html
> scp config_prod.js root@hat.adjoint.fun:/var/www/html/config.js
> scp sass/output.css root@hat.adjoint.fun:/var/www/html/sass
```
