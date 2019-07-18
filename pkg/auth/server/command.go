// Copyright 2019, Pure Storage Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/version"

	pureoidc "github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/auth/server/oidc"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/auth/server/tokenstore/kube"
	purehttp "github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/http"

	oidc "github.com/coreos/go-oidc"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Cmd provides main Command of Auth Server
func Cmd() *cobra.Command {
	var (
		a         dexApp
		issuerURL string
		listen    string
		tlsCert   string
		tlsKey    string
		debug     bool
	)
	c := cobra.Command{
		Use:     "pure1-unplugged-auth-server",
		Short:   "Authentication service for Pure1 Unplugged",
		Long:    "Proxy authentication endpoint for Dex",
		Version: version.Get(),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return errors.New("surplus arguments provided")
			}

			_, err := url.Parse(a.redirectURI)
			if err != nil {
				return fmt.Errorf("parse redirect-uri: %v", err)
			}
			listenURL, err := url.Parse(listen)
			if err != nil {
				return fmt.Errorf("parse listen address: %v", err)
			}

			// Ignore invalid certs
			var transport http.RoundTripper = &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			}

			// If debugging, wrap the existing transport in a debug transport
			if debug {
				transport = &purehttp.DebugTransport{Tripper: transport}
			}

			// Set up the client to use the given transport
			a.client = &http.Client{
				Transport: transport,
			}

			ctx := oidc.ClientContext(context.Background(), a.client)

			var provider *oidc.Provider

			for provider == nil {
				newProvider, err := oidc.NewProvider(ctx, issuerURL)
				if err != nil {
					log.WithFields(log.Fields{
						"provider": issuerURL,
						"err":      err,
					}).Error("Failed to query provider, retrying in 5 seconds")
					time.Sleep(time.Second * 5)
				} else {
					log.WithField("provider", issuerURL).Info("Queried provider successfully!")
					provider = newProvider
				}
			}

			var s struct {
				// What scopes does a provider support?
				//
				// See: https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderMetadata
				ScopesSupported []string `json:"scopes_supported"`
			}
			if err := provider.Claims(&s); err != nil {
				return fmt.Errorf("Failed to parse provider scopes_supported: %v", err)
			}

			if len(s.ScopesSupported) == 0 {
				// scopes_supported is a "RECOMMENDED" discovery claim, not a required
				// one. If missing, assume that the provider follows the spec and has
				// an "offline_access" scope.
				a.offlineAsScope = true
			} else {
				// See if scopes_supported has the "offline_access" scope.
				a.offlineAsScope = func() bool {
					for _, scope := range s.ScopesSupported {
						if scope == oidc.ScopeOfflineAccess {
							return true
						}
					}
					return false
				}()
			}

			a.provider = provider
			verifier := pureoidc.NewOIDCVerifier(provider.Verifier(&oidc.Config{ClientID: a.clientID}))
			a.verifier = &verifier
			oauth2Config := a.NewOIDCOAuth2Config()
			a.oauth2config = &oauth2Config
			kubeConn, err := kube.GetKubeSecretInterface("pure1-unplugged")
			if err != nil {
				return err
			}
			a.apiTokenStore = kube.NewKubeSecretAPITokenStore(kubeConn)

			http.HandleFunc("/login", a.handleLogin)
			http.HandleFunc("/", a.handleVerify)
			http.HandleFunc("/callback", a.handleCallback)
			http.HandleFunc("/api-token", a.handleAPIToken)

			switch listenURL.Scheme {
			case "http":
				log.WithFields(log.Fields{
					"host":    listen,
					"version": version.Get(),
				}).Info("Server listening (HTTP)")
				return http.ListenAndServe(listenURL.Host, nil)
			case "https":
				log.WithFields(log.Fields{
					"host":    listen,
					"version": version.Get(),
				}).Info("Server listening (HTTPS)")
				return http.ListenAndServeTLS(listenURL.Host, tlsCert, tlsKey, nil)
			default:
				return fmt.Errorf("listen address %q is not using http or https", listen)
			}
		},
	}
	c.Flags().StringVar(&a.clientID, "client-id", AuthServerEnvConf.ClientID, "OAuth2 client ID of this application.")
	c.Flags().StringVar(&a.clientSecret, "client-secret", AuthServerEnvConf.ClientSecret, "OAuth2 client secret of this application.")
	c.Flags().StringVar(&a.redirectURI, "redirect-uri", AuthServerEnvConf.RedirectURL, "Callback URL for OAuth2 responses.")
	c.Flags().StringVar(&issuerURL, "issuer", AuthServerEnvConf.IssuerURL, "URL of the OpenID Connect issuer.")
	c.Flags().StringVar(&listen, "listen", AuthServerEnvConf.Listen, "HTTP(S) address to listen at.")
	c.Flags().StringVar(&tlsCert, "tls-cert", AuthServerEnvConf.TLSCert, "X509 cert file to present when serving HTTPS.")
	c.Flags().StringVar(&tlsKey, "tls-key", AuthServerEnvConf.TLSKey, "Private key for the HTTPS cert.")
	c.Flags().BoolVar(&debug, "debug", AuthServerEnvConf.Debug, "Print all request and responses from the OpenID Connect issuer.")
	return &c
}
