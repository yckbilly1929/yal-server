Yal Server
===========

Minimal Development Server with **Live Reload** Capability.  
(**Y**et **A**nother **Live Server**)

- Rewritten in Golang
- Supportive tool for esbuild

Roadmap
--------

- github actions
- npm publish
- support recursive watch
- accept json config file, and more dynamic config options like live-server
- refresh css only / possibility to support HMR
- example project with esbuild

Get Started
------------

```bash
  # Global
  $ npm i -g yal-server

  # As node package
  $ npm i -D yal-server
```

Usage from command line
-----------------------

```bash
  # Help
  $ yal-server

  # example dev command
  $ yal-server dev -c='{"root": "dist", "port": 5501, "cors": true, "https": true, "fallback": true, "proxy": [{"prefix": "/api", "target": "https://backend", "changeOrigin": true}]}'
```

Usage from node
---------------

```javascript
const yalServer = require("yal-server")

const conf = {
  root: 'dist',
  port: 5501,
  cors: true,
  https: true,
  fallback: true,
  proxy: [],
}

yalServer.dev(conf)
```
