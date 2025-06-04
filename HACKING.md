# Hacking

For reasons I don't understand, CoreDNS plugins can't seem to include `go.mod`
and `go.sum` files. See
https://github.com/jfly/coredns-data-mesher/commit/a63bcbec520fbff198f714ad4e7f112f04ab9960
for details. Please let me know if you know why this is, or if there's a better way to handle this.

## Enter dev shell

To allow for development to actually happen, `go.mod` and `go.sum` are
gitignored. They get generated automatically when you enter the dev shell:

```console
nix develop
```

## Run tests

In a dev shell:

```console
go test ./...
```

## Format code

```console
nix fmt
```
