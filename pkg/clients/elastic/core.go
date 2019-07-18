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

package elastic

import (
	"context"
	"fmt"
	"time"

	"github.com/olivere/elastic"
	log "github.com/sirupsen/logrus"
)

// tryRepeatReturnErrorOnly is a utility method that tries the given function
// until it runs out of attempts or succeeds
func (e *Client) tryRepeatReturnErrorOnly(fn errorReturnOnlyFunction) error {
	infiniteRetry := e.maxAttempts == 0
	try := uint(0)
	var lastError error

	for infiniteRetry || try < e.maxAttempts {
		err := fn()
		if err == nil {
			return nil
		}
		lastError = err
		time.Sleep(e.retryTime)
		try++
	}
	return fmt.Errorf("Max number of attempts reached. Last error: %v", lastError)
}

// tryRepeatReturnBoolError is a utility method that tries the given function
// until it runs out of attempts or succeeds, which also returns a boolean result
func (e *Client) tryRepeatReturnBoolError(fn boolErrorReturnFunction) (bool, error) {
	infiniteRetry := e.maxAttempts == 0
	try := uint(0)
	var lastError error

	for infiniteRetry || try < e.maxAttempts {
		res, err := fn()
		if err == nil {
			return res, nil
		}
		lastError = err
		time.Sleep(e.retryTime)
		try++
	}
	return false, fmt.Errorf("Max number of attempts reached. Last error: %v", lastError)
}

// tryRepeatReturnStringSliceError is a utility method that tries the given function
// until it runs out of attempts or succeeds, which also returns a boolean result
func (e *Client) tryRepeatReturnStringSliceError(fn stringSliceErrorReturnFunction) ([]string, error) {
	infiniteRetry := e.maxAttempts == 0
	try := uint(0)
	var lastError error

	for infiniteRetry || try < e.maxAttempts {
		res, err := fn()
		if err == nil {
			return res, nil
		}
		lastError = err
		time.Sleep(e.retryTime)
		try++
	}
	return nil, fmt.Errorf("Max number of attempts reached. Last error: %v", lastError)
}

// InitializeClient creates a new Client and has it attempt to connect
func InitializeClient(host string, maxAttempts uint, attemptDelay time.Duration) (*Client, error) {
	client := &Client{
		host:        host,
		maxAttempts: maxAttempts,
		retryTime:   attemptDelay,
		infoLog:     log.New(),
		errorLog:    log.New(),
	}
	client.infoLog.SetLevel(log.InfoLevel)
	client.errorLog.SetLevel(log.ErrorLevel)
	err := client.Connect()
	if err != nil {
		return nil, err
	}
	return client, nil
}

// Connect will attempt to connect this Client instance to the Elastic server
func (e *Client) Connect() error {
	return e.tryRepeatReturnErrorOnly(func() error {
		newclient, err := elastic.NewClient(elastic.SetSniff(false),
			elastic.SetURL("http://"+e.host),
			elastic.SetErrorLog(e.errorLog),
			elastic.SetInfoLog(e.infoLog),
		)
		if err != nil {
			return err
		}
		e.esclient = newclient
		return nil
	})
}

// Connected checks if this Client is connected with a couple basic checks
func (e *Client) Connected(ctx context.Context) bool {
	if e.esclient == nil {
		return false
	}

	// Run a simple request, just basically ping elasticsearch using the client,
	// to see if it's connected and can even reach it
	_, err := e.esclient.PerformRequest(ctx, elastic.PerformRequestOptions{
		Path:   "/",
		Method: "GET",
	})
	return err == nil
}

// EnsureConnected provides a method such that, once called, the client is guaranteed
// to be connected or an error will be thrown
func (e *Client) EnsureConnected(ctx context.Context) error {
	if !e.Connected(ctx) {
		log.Debug("Client not connected, connecting")
		err := e.Connect()
		if err != nil {
			log.WithError(err).Error("Error connecting to Elasticsearch")
			return err
		}
	}

	return nil
}

// DeleteIndices deletes the given indices
func (e *Client) DeleteIndices(ctx context.Context, indexNames []string) error {
	return e.tryRepeatReturnErrorOnly(func() error {
		_, err := e.esclient.DeleteIndex().Index(indexNames).Do(ctx)
		return err
	})
}

// DeleteByQuery deletes documents by query in the given index
func (e *Client) DeleteByQuery(ctx context.Context, indexName string, query elastic.Query) error {
	return e.tryRepeatReturnErrorOnly(func() error {
		_, err := e.esclient.DeleteByQuery(indexName).Query(query).Do(ctx)
		return err
	})
}

// Client returns the inner Elastic client to perform standard Elasticsearch calls on
func (e *Client) Client() *elastic.Client {
	return e.esclient
}

func (e *Client) createTemplate(ctx context.Context, name string, body interface{}) error {
	result, err := e.esclient.IndexPutTemplate(name).BodyJson(body).Do(ctx)
	if err != nil {
		return err
	}
	if !result.Acknowledged {
		return fmt.Errorf("Result was not acknowledged")
	}
	return nil
}
