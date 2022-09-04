# chefbrowser (cinc)

A web application for browsing objects on a chef (cinc) server. Inspiration taken from the
abandoned [chef-browser](https://github.com/3ofcoins/chef-browser) ruby/sinatra application.

Disclaimer: This is mostly a way for me to tinker with Go and GitHub workflows, so this project should not be considered
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
      --config string   config file (default is $HOME/.chefbrowser.yaml)
  -h, --help            help for chefbrowser
```

## TODO

- [ ] Decide between `html/template` and a React SPA for the frontend UI
- [ ] Test suite
- [ ] Drop Cobra (do we need it?)
- [ ] Build pipeline
- [ ] Support browsing multiple chef organizations

## License

MIT
