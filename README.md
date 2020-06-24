mkr
===
[![Latest Version](https://img.shields.io/github/release/mackerelio/mkr.svg?style=flat-square)][release]
[![Build Status](https://img.shields.io/travis/mackerelio/mkr.svg?style=flat-square)][travis]
[![Go Documentation](https://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)][godocs]

[release]: https://github.com/mackerelio/mkr/releases
[travis]: http://travis-ci.org/mackerelio/mkr
[godocs]: http://godoc.org/github.com/mackerelio/mkr

mkr - Command Line Tool For Mackerel written in Go.

# DESCRIPTION

mkr is a command-line interface tool for the [Mackerel API](https://mackerel.io/api-docs/) written in Go.
mkr helps to automate tedious daily server operations to best leverage Mackerel's and Unix's tools.
mkr output format is JSON, so it can be filtered with a JSON processor such as [jq](http://stedolan.github.io/jq/).

# INSTALLATION

Install the plugin package from either the yum or the apt repository.

## CentOS 5/6

```bash
curl -fsSL https://mackerel.io/assets/files/scripts/setup-yum.sh | sh
yum install mkr
```

## Debian 6/7

```bash
curl -fsSL https://mackerel.io/assets/files/scripts/setup-apt.sh | sh
apt-get install mkr
```

## Homebrew
You can also install from the brew rule we maintain, but we don't officially support the environment.
```bash
brew tap mackerelio/mackerel-agent
brew install mkr
```

## Build from source

```bash
$ GO111MODULE=on go get github.com/mackerelio/mkr
```

# USAGE

First the MACKEREL_APIKEY environment variable must be set. It is not necessary to set the MACKEREL_APIKEY on hosts running [mackerel-agent](https://github.com/mackerelio/mackerel-agent). For more details, see below.

```bash
export MACKEREL_APIKEY=<Put your API key>
```

## EXAMPLES

```
mkr status <hostId>
{
    "id": "2eQGEaLxiYU",
    "name": "myproxy001",
    "status": "standby",
    "roleFullnames": [
        "My-Service:proxy"
    ],
    "isRetired": false,
    "createdAt": "2014-11-15T21:41:00+09:00"
}
```

```
mkr hosts --service My-Service --role proxy
[
    {
        "id": "2eQGEaLxiYU",
        "name": "myproxy001",
        "status": "standby",
        "roleFullnames": [
            "My-Service:proxy"
        ],
        "isRetired": false,
        "createdAt": "2014-11-15T21:41:00+09:00"
    },
    {
        "id": "2eQGDXqtoXs",
        "name": "myproxy002",
        "status": "standby",
        "roleFullnames": [
            "My-Service:proxy"
        ],
        "isRetired": false,
        "createdAt": "2014-11-15T21:41:00+09:00"
    },
]
```

The format of `createdAt` uses ISO-8601 in the current time zone.

The `mkr hosts` command has an '-f' option to format the output.

```
mkr hosts -f '{{range .}}{{if (len .Interfaces)}}{{(index .Interfaces 0).IPAddress}}{{end}}{{"\t"}}{{.Name}}{{"\n"}}{{end}}'
10.0.1.1  myproxy001
10.0.1.2  myproxy002
...
```

```
mkr create --status working -R My-Service:db-master mydb001
mkr update --status maintenance --roleFullname My-Service:db-master <hostId>
```

```
cat <<EOF | mkr throw --host <hostId>
<name>  <value> <time>
<name>  <value> <time>
EOF
...

cat <<EOF | mkr throw --service My-Service
<name>  <value> <time>
<name>  <value> <time>
EOF
...
```

```
mkr fetch --name loadavg5 2eQGDXqtoXs
{
    "2eQGDXqtoXs": {
        "loadavg5": {
            "time": 1416061500,
            "value": 0.025
        }
    }
}
```

```
mkr retire <hostId> ...
```

### Examples (on hosts running mackerel-agent)

Specifying the <hostId> and MACKEREL_APIKEY is not necessary because mkr refers to /var/lib/mackerel-agent/id and /etc/mackerel-agent/mackerel-agent.conf instead of specifying manually.

```
mkr status
```

```
mkr update --status maintenance <hostIds>...
```

```
mkr fetch --name loadavg5 <hostId>
```

```bash
cat <<EOF | mkr throw --host <hostId>
<name>  <value> <time>
EOF
```

```
mkr retire
```

## ADVANCED USAGE

```bash
$ mkr update --st working $(mkr hosts -s My-Service -r proxy | jq -r '.[].id')
```

## Using Docker Image

https://registry.hub.docker.com/u/mackerel/mkr/

```bash
$ docker run --rm --env MACKEREL_APIKEY=<API key> mackerel/mkr help
```

# CONTRIBUTION

1. Fork ([https://github.com/mackerelio/mkr/fork](https://github.com/mackerelio/mkr/fork))
2. Create a feature branch
3. Commit your changes
4. Rebase your local changes against the master branch
5. Run test suite with the `go test ./...` command and confirm that it passes
6. Run `gofmt -s`
7. Create new Pull Request


License
----------

Copyright 2014 Hatena Co., Ltd.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
