# coredns-data-mesher

## Name

*data-mesher* - responds to DNS queries with [Data Mesher's](https://data-mesher.docs.clan.lol)'s distributed DNS

## Description

Data Mesher is an eventually consistent, decentralised data service that
provides DNS for a cluster. It's a part of the [Clan](https://clan.lol) computer management
framework.

Data Mesher works great on machines that are managed by Clan. However, if you
want to add a mobile device, you need some way for it to learn about DNS in
your cluster. See <https://git.clan.lol/clan/clan-core/issues/1268> for more
information.

## Compilation

This package is compiled as part of CoreDNS and not in a standalone
way.

The [manual](https://coredns.io/manual/toc/#what-is-coredns) has more
information about how to configure and extend the server with external plugins.

A simple way to consume this plugin is by adding the following to
[plugin.cfg](https://github.com/coredns/coredns/blob/master/plugin.cfg), and
recompiling it as [detailed on
coredns.io](https://coredns.io/2017/07/25/compile-time-enabling-or-disabling-plugins/#build-with-compile-time-configuration-file).

~~~
data-mesher:github.com/jfly/coredns-data-mesher
~~~

Put this early in the plugin list, so that *example* is executed before any of the other plugins.

After this you can compile `coredns` by:

``` sh
make
```

## Syntax

~~~ txt
data-mesher
~~~

## Metrics

If monitoring is enabled (via the *prometheus* directive) the following metric is exported:

* `coredns_data_mesher_request_count_total{server}` - query count to the *data_mesher* plugin.

The `server` label indicated which server handled the request, see the *metrics* plugin for details.

## Ready

This plugin reports readiness to the ready plugin. It will be immediately ready.

## Examples

In this configuration, we first query Data Mesher, if Data Mesher does not
have a response, we then forward queries to 9.9.9.9.

~~~ corefile
. {
  debug
  log

  data-mesher
  forward . 8.8.8.8 9.9.9.9
}
~~~
