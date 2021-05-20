package main

import (
	mPlugin "github.com/oidc-proxy-ecosystem/oidc-plugin-sdk/plugin/session"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/session"
	"github.com/oidc-proxy-ecosystem/oidc-session-memory/memory"
)

func main() {
	mPlugin.Sever(&mPlugin.ServerOpts{
		GRPCSessionFunc: func() session.Session {
			return memory.New()
		},
	})
}
