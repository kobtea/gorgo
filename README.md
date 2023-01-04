# gorgo

GitHub Organization Organizer.

## Overview

**This project is unstable yet. So breaking change may be happened no notice.**

Gorgo improves a regulation against repositories of GitHub organizations.
It checks org repositories widely for policy compliance.

Gorgo uses [Conftest](https://github.com/open-policy-agent/conftest) as policy engine.
The reason why we use Conftest instead of OPA directly is that we want to follow Conftest's policy format and output format.

Use cases:

- Cross-cutting policy checks with Conftest
- Detecting inactive repositories
- Detecting repositories that do not have the expected CI configured

## Install

### Binary

Go to https://github.com/kobtea/gorgo/releases

### Go get

```bash
$ go install github.com/kobtea/gorgo@latest
```

### Docker

```bash
$ docker run -it --rm -v $(pwd)/path/to/config.yaml:/gorgo.yaml ghcr.io/kobtea/gorgo --help
```

## Usage

```bash
$ gorgo --help
GitHub Organization Organizer

Usage:
  gorgo [command]

Available Commands:
  check       Test policies
  clean       Remove contents at working directory
  completion  Generate the autocompletion script for the specified shell
  fetch       Retrieve repository metadata
  help        Help about any command
  version     Show version

Flags:
      --config string      config file (default "./gorgo.yaml")
  -h, --help               help for gorgo
      --log-level string   log level (default "info")

Use "gorgo [command] --help" for more information about a command.
```

Configuration format is below.

```yaml
# working_dir is where gorgo reads/writes temporary files
working_dir: <string> # default: `tmp`
# configuration for each GitHub endpoint, including GitHub Enterprise Server
github_configs:
  -
    # need domain and endpoints when you want to check other than `https://github.com`
    domain: <string> # default: github.com
    api_endpoint: <string>
    upload_endpoint: <string>
    # envvar name of github token, not token value itself
    token_envvar_name: <string> # default: GITHUB_TOKEN
    # configuration for each repository
    repo_configs:
      -
        # user or organization name
        owner: <string>
        # regex pattern for repository name
        # regex format is RE2 https://golang.org/s/re2syntax
        regex: <string>
        # configuration for Conftest
        conftest_configs:
          -
            # input file type
            # repo: response body of `/repo` in GitHub api
            # src: source code of the repository
            target: <string> # `repo`, `src`
            # input file path for conftest
            # root dir is repository root
            input: <string>
            # combine flag of conftest
            combine: <bool>
            # policy file path written in rego
            # policy format follows conftest format
            policies: [ <string> ]
```

example: https://github.com/kobtea/gorgo/blob/main/example/config.yaml

```bash
# download metadata and source code if needed
$ gorgo fetch --config ./example/config.yaml

# run conftest and check policy against each repository
$ gorgo check --config ./example/config.yaml 2> /dev/null
WARN - testdata/tmp/metadata/github.com/kobtea/jsonnet-libs/repo.json - github.repo - GitHub repository should be pushed at least once every 6 month
WARN - testdata/tmp/src/github.com/kobtea/setup-jsonnet-action/.github/workflows/test.yml - github.actions - GitHub actions should be defined `Install dependencies` step
WARN - testdata/tmp/metadata/github.com/kobtea/dns_lookup_exporter/repo.json - github.repo - GitHub repository should be pushed at least once every 6 month
WARN - testdata/tmp/metadata/github.com/kobtea/mysqld_exporter/repo.json - github.repo - GitHub repository should be pushed at least once every 6 month

8 tests, 4 passed, 4 warnings, 0 failures, 0 exceptions

# run conftest directly for debug
$ conftest test ./testdata/tmp/metadata/github.com/kobtea/dns_lookup_exporter/repo.json -p ./example/policy/github_repo.rego --all-namespaces
WARN - ./testdata/tmp/metadata/github.com/kobtea/dns_lookup_exporter/repo.json - github.repo - GitHub repository should be pushed at least once every 6 month

1 test, 0 passed, 1 warning, 0 failures, 0 exceptions
```

## License

Apache-2.0
