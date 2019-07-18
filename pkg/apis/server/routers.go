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
	"net/http"
	"time"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/clients/elastic"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/clients/kube"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/db"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/logger"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

const (
	elasticRetryTime = 5 * time.Second
)

// NewRouter will create a router configured with all methods, names, paths, queries, and handlers which are defined in routes.go
func NewRouter() *mux.Router {
	secretAccess, err := kube.GetKubeSecretInterface("pure1-unplugged")
	if err != nil {
		log.WithError(err).Fatal("Error getting k8s secret interface")
		return nil
	}
	tokenStore := kube.NewKubeSecretAPITokenStore(secretAccess)

	elasticMeta, err := elastic.InitializeClient(APIServerEnv.ElasticHost, 0, elasticRetryTime)
	if err != nil {
		log.WithError(err).Fatal("Error getting Elastic connection")
		return nil
	}
	connection = db.MetadataConnection{DAO: elasticMeta, Tokens: tokenStore}

	// Essentially means that "/path" redirects to "/path/"
	// "your application will always see the path as specified in the route"
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler

		// Use logger to make a record every time the handler is called.
		handler = route.HandlerFunc
		handler = logger.Logger(handler, route.Name)

		router.
			Methods(route.Method).
			Name(route.Name).
			Path(route.Pattern).
			Queries(route.Query...).
			Handler(handler)
		// if query params is not correct, route will ignore this wrong query params.
		// It's a hack way. Still don't know whether it's gorilla mux' bug
		// https://stackoverflow.com/questions/45378566/gorilla-mux-optional-query-values
		router.
			Methods(route.Method).
			Name(route.Name).
			Path(route.Pattern).
			Handler(handler)
	}
	return router
}
