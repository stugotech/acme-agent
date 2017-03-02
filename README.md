# acme-agent

Responds to [ACME http-01 challenges](https://github.com/ietf-wg-acme/acme/blob/master/draft-ietf-acme-acme.md#http) by looking up a KV store.

## Usage

Run the command with no arguments to get the built-in help:

    NAME:
      acme-agent - solve ACME HTTP challenges using a KV store

    USAGE:
      acme-agent [global options] command [command options] [arguments...]

    COMMANDS:
        serve    Run a HTTP server to respond to challenges
        help, h  Shows a list of commands or help for one command

    GLOBAL OPTIONS:
      --log value           logging level [debug|info|warning|error|fatal|panic] (default: "info")
      --store value         key value store to use [etcd|consul|boltdb|zookeeper] (default: "etcd")
      --store-nodes value   comma-separated list of KV nodes (authority only) (default: "127.0.0.1:2379")
      --store-prefix value  prefix in KV store (default: "acme-agent")
      --help, -h            show help
      --version, -v         print the version

The cli only supports one command at present:

    NAME:
      acme-agent serve - Run a HTTP server to respond to challenges

    USAGE:
      acme-agent serve [command options] [arguments...]

    OPTIONS:
      --listen value      TCP interface or unix socket to listen on (default: ":8080")
      --path-prefix value TCP interface or unix socket to listen on (default: ".well-known/acme-challenge")


## Example 

Start the server:

    acme-agent --store=etcd --store-nodes=127.0.0.1:2379 --store-prefix=acme-agent \
        serve --listen=":8080" --path-prefix=".well-known/acme-challenge" &

Set a value in etcd:

    etcdctl set /acme-agent/fred flintstone

Test the retrieval:

    curl http://127.0.0.1:8080/.well-known/acme-challenge/fred

The server should respond with:

    flintstone
