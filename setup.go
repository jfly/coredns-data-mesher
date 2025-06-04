package data_mesher

import (
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

// Register this plugin.
func init() { plugin.Register("data-mesher", setup) }

// `setup` is called when the config parser sees the token "data-mesher". It is responsible
// for parsing any extra options the plugin may have.
func setup(c *caddy.Controller) error {
	c.Next() // Give us the next token.

	// The first token we see had better be "data-mesher", otherwise something weird is going on.
	if c.Val() != "data-mesher" {
		return plugin.Error("data-mesher", c.ArgErr())
	}

	if c.NextArg() {
		// If there is another token, return an error, because we don't have any configuration.
		return plugin.Error("data-mesher", c.ArgErr())
	}

	// Add the Plugin to CoreDNS, so Servers can use it in their plugin chain.
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return DataMesher{
			stateDir: "/var/lib/data-mesher", // https://git.clan.lol/clan/data-mesher/src/commit/bca54baa18fcbfb73dada430cfdac8e55c0532a4/nix/nixosModules/data-mesher/module.nix#L85
			Next:     next,
		}
	})

	return nil
}
