![GitHub](https://img.shields.io/github/license/drewhammond/chefbrowser)
[![Go Report Card](https://goreportcard.com/badge/github.com/drewhammond/chefbrowser)](https://goreportcard.com/report/github.com/drewhammond/chefbrowser)

# Chef Browser (2022)

A simple read-only web application for browsing objects on a Chef Infra Server (or [Cinc Server](https://cinc.sh/)).

Inspiration taken from the abandoned [chef-browser](https://github.com/3ofcoins/chef-browser) ruby/sinatra application.

>Disclaimer: This is mostly a way for me to tinker with Go and GitHub workflows, so this project should not be considered
stable until I remove this notice.

## Installation

Two methods of installation are planned:

1. Traditional deployment using systemd
2. Docker container

```shell
docker run -d -v $(pwd)/conf:/conf drewhammond/chefbrowser:latest
```

## Usage

```shell
A web application for viewing chef server resources

Usage:
  chefbrowser [flags]

Flags:
      --config string   path to config file
  -h, --help            help for chefbrowser
  -v, --version         version for chefbrowser
```

## Contributing

This project is in its infancy so any and all contributes are welcomed! If you're looking for something to work on,
I think the frontend could use some love.

## TODO

- [ ] Decide between `html/template` and a React SPA for the frontend UI
- [ ] Test suite
- [ ] Drop Cobra (do we need it?)
- [ ] Build pipeline
- [ ] Support browsing multiple chef organizations

## License

MIT
