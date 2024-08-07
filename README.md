![GitHub](https://img.shields.io/github/license/drewhammond/chefbrowser)
[![Go Report Card](https://goreportcard.com/badge/github.com/drewhammond/chefbrowser)](https://goreportcard.com/report/github.com/drewhammond/chefbrowser)
[![go-test](https://github.com/drewhammond/chefbrowser/actions/workflows/go-test.yml/badge.svg)](https://github.com/drewhammond/chefbrowser/actions/workflows/go-test.yml)

# Chef Browser (2024)

A simple read-only web application for browsing objects on a Chef Infra Server (or [Cinc Server](https://cinc.sh/)).

Inspiration taken from the abandoned [chef-browser](https://github.com/3ofcoins/chef-browser) ruby/sinatra application.

## Installation

### Configuration

All configurable settings are documented in [`defaults.ini`](defaults.ini).

Two methods of installation are planned:

1. Traditional deployment using systemd
2. Docker container

```shell
docker run -d \
  -p 8080:8080 \
  -v /path/to/example.pem:/example.pem:ro \
  -v /path/to/config.ini:/config.ini:ro \
  drewhammond/chefbrowser:latest --config /config.ini
```

## Usage

```
chefbrowser is a read-only web application for viewing
Chef Infra Server (or Cinc Server) resources

Usage:
  chefbrowser --config /path/to/config.ini [flags]

Flags:
      --config string   path to config file
  -h, --help            help for chefbrowser
  -v, --version         version for chefbrowser
```

### Adding custom links to node pages

You can add custom links to node pages to easily pivot between internal systems.

Links can be static or templated using node attributes (in the `title` and `href` fields) by enclosing the desired
attribute key between `{{` `}}`, for example:

- `{{fqdn}}`
- `{{ipaddress}}`
- `{{ec2.instance_id}}`

Nested attributes should use dots as delimiters. If you reference a node attribute that does not exist, it will be replaced by `undefined` in the final URL.

**Config:**

```ini
[custom_links.nodes.0]
title = "Open Example Dashboard (Grafana)"
href = "https://grafana.example.com/d/000001/example-dashboard?var-hostname={{fqdn}}"
new_tab = true

[custom_links.nodes.1]
title = "Another Internal Resource (ID: {{ec2.instance_id}})"
href = "https://docs.example.com/foo/?instance-id={{ec2.instance_id}}"
```


## Contributing

This project is in its infancy so any and all contributes are welcome! If you're looking for something to work on,
I think the frontend could use some love.

## Development

Set `app_mode = development` in your config file to enable developer mode. This mode does the following:

- Go Templates are loaded from the file system instead of embedded into the backend. They can be changed without
  recompiling.
- CSS/JS links in the HTML are updated to point to the local Vite dev server for live reloading.

Install UI dependencies and start the frontend development server:

```shell
make start-ui-dev
```

Build and start the backend server:

```shell
make build-backend
./dist/chefbrowser --config development.ini
```

Access the dev server at http://localhost:8080.

CSS/JS changes will trigger automatic rebuilds
as long as you have the frontend development server running.

> Note: Go changes will not be live reloaded. Rebuild backend for changes to take effect.

## TODO

- [ ] Test suite
- [ ] Drop Cobra (do we need it?)
- [ ] Support browsing multiple chef organizations
- [ ] Windows support? (if you are interested, please file an issue!)

## License

MIT
