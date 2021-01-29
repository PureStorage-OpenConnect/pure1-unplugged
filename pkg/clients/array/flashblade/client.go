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

package flashblade

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/util"
	"github.com/go-resty/resty"
	log "github.com/sirupsen/logrus"
)

// Endpoint constants
const (
	AlertsEndpoint                  = "/alerts"
	APIPrefix                       = "/api"
	APIVersionEndpoint              = "/api/api_version"
	ArraysEndpoint                  = "/arrays"
	ArraysPerformanceEndpoint       = "/arrays/performance"
	ArraysSpaceEndpoint             = "/arrays/space"
	FileSystemCountEndpoint         = "/file-systems?limit=1"
	FileSystemsEndpoint             = "/file-systems"
	FileSystemsPerformanceEndpoint  = "/file-systems/performance?protocol=nfs&limit=5"
	FileSystemSnapshotCountEndpoint = "/file-system-snapshots?limit=1"
	FileSystemSnapshotsEndpoint     = "/file-system-snapshots"
	LoginEndpoint                   = "/api/login"
)

// Other constants
const (
	APITokenHeader                  = "api-token"
	AuthTokenHeader                 = "x-auth-token"
	FileSystemPerformanceResolution = 30000 // ms
	PreferredAPIVersion             = "1.5"
	RequestAttemptCount             = 3
	UserAgent                       = "Pure1 Unplugged FlashBlade Client v1.0"
	UserAgentHeader                 = "User-Agent"
)

// NewClient creates a new FlashBlade client and initializes it by getting the API version,
// refreshing a new session, and getting the array metadata
func NewClient(displayName string, managementEndpoint string, apiToken string) (ArrayClient, error) {
	// Ignore the verification for using HTTPS
	resty.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	// Convert the endpoint to an IP
	ip, err := util.ParseEndpoint(displayName, managementEndpoint)
	if err != nil {
		return nil, err
	}

	client := Client{
		APIToken:     apiToken,
		DisplayName:  displayName,
		ManagementIP: ip,
	}

	// Get the API Versions and verify the preferred one is supported
	apiVersion, err := client.getAPIVersion()
	if err != nil {
		client.logCreationError()
		return nil, err
	}
	client.APIVersion = apiVersion

	log.WithFields(log.Fields{
		"api_version":  client.APIVersion,
		"display_name": client.DisplayName,
	}).Info("Successfully created FlashBlade Client")
	return &client, nil
}

// GetAlerts returns all alerts from the array
func (client *Client) GetAlerts() ([]*AlertResponse, error) {
	url := client.createFullURL(AlertsEndpoint)
	response, _, err := client.performGet(url, AlertGenericResponse{})
	if err != nil {
		return nil, err
	}

	result := response.(*AlertGenericResponse)
	return result.Items, nil
}

// GetArrayCapacityMetrics returns the capacity metrics from the array
func (client *Client) GetArrayCapacityMetrics() (*ArrayCapacityMetricsResponse, error) {
	url := client.createFullURL(ArraysSpaceEndpoint)
	response, _, err := client.performGet(url, ArrayCapacityMetricsGenericResponse{})
	if err != nil {
		return nil, err
	}

	result := response.(*ArrayCapacityMetricsGenericResponse)
	return result.Items[0], nil
}

// GetArrayInfo returns the array metadata
func (client *Client) GetArrayInfo() (*ArrayInfoResponse, error) {
	url := client.createFullURL(ArraysEndpoint)
	response, _, err := client.performGet(url, ArrayInfoGenericResponse{})
	if err != nil {
		return nil, err
	}

	result := response.(*ArrayInfoGenericResponse)
	return result.Items[0], nil
}

// GetArrayPerformanceMetrics returns the performance metrics from the array
func (client *Client) GetArrayPerformanceMetrics() (*ArrayPerformanceMetricsResponse, error) {
	url := client.createFullURL(ArraysPerformanceEndpoint)
	response, _, err := client.performGet(url, ArrayPerformanceMetricsGenericResponse{})
	if err != nil {
		return nil, err
	}

	result := response.(*ArrayPerformanceMetricsGenericResponse)
	return result.Items[0], nil
}

// GetFileSystemCapacityMetrics returns the capacity metrics for the file systems
func (client *Client) GetFileSystemCapacityMetrics() ([]*FileSystemCapacityMetricsResponse, error) {
	url := client.createFullURL(FileSystemsEndpoint)
	response, _, err := client.performGet(url, FileSystemCapacityMetricsGenericResponse{})
	if err != nil {
		return nil, err
	}

	result := response.(*FileSystemCapacityMetricsGenericResponse)
	return result.Items, nil
}

// GetFileSystemCount returns the count of file systems
func (client *Client) GetFileSystemCount() (uint32, error) {
	url := client.createFullURL(FileSystemCountEndpoint)
	response, _, err := client.performGet(url, FileSystemCapacityMetricsGenericResponse{})
	if err != nil {
		return 0, err
	}

	result := response.(*FileSystemCapacityMetricsGenericResponse)
	return result.PaginationInfo.TotalItemCount, nil
}

// GetFileSystemPerformanceMetrics returns the performance metrics for the file systems
func (client *Client) GetFileSystemPerformanceMetrics(window int64) ([]*FileSystemPerformanceMetricsResponse, error) {
	// File system performance metrics are limited to 5 per response, so multiple requests and the use of a continuation
	// token are needed to gather all metrics
	baseURL := client.createFullURL(FileSystemsPerformanceEndpoint)
	var metricsResponses []*FileSystemPerformanceMetricsResponse
	var continuationToken string

	// Set the time parameters (they are required in ms)
	endTime := time.Now()
	// Subtract an extra second so we can be sure to encompass at least one data point
	startTime := endTime.Add(time.Duration(-window-1) * time.Second)
	baseURL = fmt.Sprintf("%s&resolution=%d&start_time=%d&end_time=%d",
		baseURL, FileSystemPerformanceResolution, startTime.Unix()*1000, endTime.Unix()*1000)

	// Make initial request
	responseItems, token, err := client.fetchFileSystemPerformanceMetrics(baseURL)
	if err != nil {
		return nil, err
	}
	metricsResponses = append(metricsResponses, responseItems...)
	continuationToken = token

	// Make subsequent requests until the list is exhausted
	for continuationToken != "" {
		fullURL := fmt.Sprintf("%s&token=%s", baseURL, continuationToken)
		responseItems, token, err := client.fetchFileSystemPerformanceMetrics(fullURL)
		if err != nil {
			return nil, err
		}
		metricsResponses = append(metricsResponses, responseItems...)
		continuationToken = token
	}

	return metricsResponses, nil
}

// GetFileSystemSnapshotCount returns the count of file system snapshots
func (client *Client) GetFileSystemSnapshotCount() (uint32, error) {
	url := client.createFullURL(FileSystemSnapshotCountEndpoint)
	response, _, err := client.performGet(url, FileSystemSnapshotGenericResponse{})
	if err != nil {
		return 0, err
	}

	result := response.(*FileSystemSnapshotGenericResponse)
	return result.PaginationInfo.TotalItemCount, nil
}

// GetFileSystemSnapshots returns all file system snapshots
func (client *Client) GetFileSystemSnapshots() ([]*FileSystemSnapshotResponse, error) {
	url := client.createFullURL(FileSystemSnapshotsEndpoint)
	response, _, err := client.performGet(url, FileSystemSnapshotGenericResponse{})
	if err != nil {
		return nil, err
	}

	result := response.(*FileSystemSnapshotGenericResponse)
	return result.Items, nil
}

// createFullURL is a helper function that returns a URL for the specified endpoint/params with the
// management endpoint and API version
func (client *Client) createFullURL(endpoint string) string {
	return fmt.Sprintf("https://%s%s/%s%s", client.ManagementIP, APIPrefix, client.APIVersion, endpoint)
}

// fetchFileSystemPerformanceMetrics is a helper function to make one single request to get file system performance
// metrics (using a continuation token) and return the subset of metrics and next continuation token
func (client *Client) fetchFileSystemPerformanceMetrics(fullURL string) ([]*FileSystemPerformanceMetricsResponse, string, error) {
	response, _, err := client.performGet(fullURL, FileSystemPerformanceMetricsGenericResponse{})
	if err != nil {
		return nil, "", err
	}
	result := response.(*FileSystemPerformanceMetricsGenericResponse)
	return result.Items, result.PaginationInfo.ContinuationToken, nil
}

// getAPIVersion is a helper function that checks the available API versions and that the desired
// version is available; it warns if it is not
func (client *Client) getAPIVersion() (string, error) {
	url := fmt.Sprintf("https://%s%s", client.ManagementIP, APIVersionEndpoint)
	response, _, err := client.performGet(url, APIVersionResponse{})
	if err != nil {
		log.WithFields(log.Fields{
			"display_name": client.DisplayName,
			"url":          url,
		}).Error("Could not get proper API version")
		return "", err
	}

	result := response.(*APIVersionResponse)

	// If the preferred API version exists, we'll use that
	for _, ver := range result.Version {
		if ver == PreferredAPIVersion {
			return PreferredAPIVersion, nil
		}
	}

	// Otherwise, default to latest API version
	log.WithFields(log.Fields{
		"display_name":          client.DisplayName,
		"latest_api_version":    result.Version[len(result.Version)-1],
		"preferred_api_version": PreferredAPIVersion,
	}).Warn("Could not use preferred API version; defaulting to latest")
	return result.Version[len(result.Version)-1], nil
}

// logCreationHeader is a helper function that logs errors originating from NewClient
func (client *Client) logCreationError() {
	log.WithFields(log.Fields{
		"display_name": client.DisplayName,
	}).Error("Could not create FlashBlade Client")
}

// performGet is a helper function that encapsulates exit and retry cases for GET requests
// Returns the unmarshaled response data, response headers, and error
func (client *Client) performGet(url string, resultType interface{}) (interface{}, http.Header, error) {
	// Each request can be retried multiple times
	for i := 0; i < RequestAttemptCount; i++ {
		log.WithFields(log.Fields{
			"display_name": client.DisplayName,
			"url":          url,
		}).Trace("Making GET request")
		response, err := resty.R().SetHeader(UserAgentHeader, UserAgent).SetHeader(AuthTokenHeader, client.AuthToken).SetResult(resultType).Get(url)

		// If there was a client error we quit
		if err != nil {
			log.WithFields(log.Fields{
				"display_name": client.DisplayName,
				"error":        err,
				"url":          url,
			}).Error("Client error with GET request")
			return nil, nil, err
		}

		// Cases where we try again
		if response.StatusCode() == 401 || response.StatusCode() == 403 {
			log.WithFields(log.Fields{
				"display_name": client.DisplayName,
				"status_code":  response.StatusCode(),
				"url":          url,
			}).Trace("Session expired; refreshing session and retrying")
			_ = client.refreshSession()
			continue
		}
		if response.StatusCode() == 500 {
			log.WithFields(log.Fields{
				"display_name": client.DisplayName,
				"status_code":  response.StatusCode(),
				"url":          url,
			}).Warn("Array internal server error; waiting 500ms and retrying")
			time.Sleep(500 * time.Millisecond)
			continue
		}

		// Cases where we return
		if response.StatusCode() == 200 {
			return response.Result(), response.Header(), nil
		}

		// No error, and not a handled status code: let's print it for debugging
		log.WithFields(log.Fields{
			"display_name": client.DisplayName,
			"status_code":  response.StatusCode(),
			"url":          url,
			"body":         response.String(),
		}).Trace("Unhandled status code returned")
	}
	// Log we failed
	log.WithFields(log.Fields{
		"display_name": client.DisplayName,
		"url":          url,
	}).Error("No successful GET request")
	return nil, nil, errors.New("No successful GET request")
}

// refreshSession is a helper function that makes a POST request to refresh the client session
// and saves the X-Auth-Token header
func (client *Client) refreshSession() error {
	// Make a request to create a new session
	url := fmt.Sprintf("https://%s%s", client.ManagementIP.String(), LoginEndpoint)
	log.WithFields(log.Fields{
		"display_name": client.DisplayName,
		"url":          url,
	}).Trace("Making POST request")
	response, err := resty.R().SetHeader(UserAgentHeader, UserAgent).SetHeader(APITokenHeader, client.APIToken).Post(url)

	// Read the response
	if err != nil {
		log.WithFields(log.Fields{
			"display_name": client.DisplayName,
			"error":        err,
			"url":          url,
		}).Error("Error making request to start new session")
		return err
	}
	if response.StatusCode() != 200 {
		log.WithFields(log.Fields{
			"display_name": client.DisplayName,
			"status_code":  response.StatusCode(),
			"url":          url,
		}).Error("Could not start new session")
		return errors.New("Could not start new session")
	}
	tokenHeader := response.Header().Get(AuthTokenHeader)
	if tokenHeader == "" {
		log.WithFields(log.Fields{
			"display_name": client.DisplayName,
			"url":          url,
		}).Error("Error getting auth token header from response")
	}

	// Save the new header to use for future requests
	client.AuthToken = tokenHeader
	return nil
}
