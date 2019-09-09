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

package flasharray

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/util"
	"github.com/go-resty/resty"
	log "github.com/sirupsen/logrus"
)

// These are endpoint constants
const (
	APIPrefix                             = "/api"
	APIVersionEndpoint                    = "/api/api_version"
	ArrayEndpoint                         = "/array"
	ArrayCapacityMetricsEndpoint          = "/array?space=true"
	ArrayControllersEndpoint              = "/array?controllers=true"
	ArrayPerformanceMetricsEndpoint       = "/array?action=monitor&size=true"
	HostCountEndpoint                     = "/host?start=0&limit=1"
	MessageFlaggedEndpoint                = "/message?flagged=true"
	MessageTimelineEndpoint               = "/message?timeline=true"
	SessionEndpoint                       = "/auth/session"
	VolumeCapacityMetricsEndpoint         = "/volume?space=true"
	VolumeCountEndpoint                   = "/volume?start=0&limit=1"
	VolumePerformanceMetricsEndpoint      = "/volume?action=monitor"
	VolumePendingEradicationCountEndpoint = "/volume?pending_only=true&start=0&limit=1"
	VolumeSnapshotCountEndpoint           = "/volume?snap=true&start=0&limit=1"
	VolumeSnapshotsEndpoint               = "/volume?snap=true"
)

// These are other constants
const (
	PreferredAPIVersion  = "1.7"
	RequestAttemptCount  = 3
	TotalItemCountHeader = "x-total-item-count"
	UserAgent            = "Pure1 Unplugged FlashArray Client v1.0"
	UserAgentHeader      = "User-Agent"
)

// NewClient creates a new FlashArray client and initializes it by getting the API version,
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

	// Get the API versions and verify the preferred one is supported
	apiVersion, err := client.getAPIVersion()
	if err != nil {
		client.logCreationError()
		return nil, err
	}
	client.APIVersion = apiVersion

	log.WithFields(log.Fields{
		"api_version":  client.APIVersion,
		"display_name": client.DisplayName,
	}).Info("Successfully created FlashArray Client")
	return &client, nil
}

// GetAlertsFlagged returns only flagged alert messages from the array
func (client *Client) GetAlertsFlagged() ([]*AlertResponse, error) {
	return client.getAlerts(MessageFlaggedEndpoint)
}

// GetAlertsTimeline returns all alert messages from the array with a "closed" response field
func (client *Client) GetAlertsTimeline() ([]*AlertResponse, error) {
	return client.getAlerts(MessageTimelineEndpoint)
}

// GetArrayCapacityMetrics returns all capacity metrics for the array
func (client *Client) GetArrayCapacityMetrics() (*ArrayCapacityMetricsResponse, error) {
	url := client.createFullURL(ArrayCapacityMetricsEndpoint)
	response, _, err := client.performGet(url, []ArrayCapacityMetricsResponse{})
	if err != nil {
		return nil, err
	}

	result := response.(*[]ArrayCapacityMetricsResponse)
	return &(*result)[0], nil
}

// GetArrayInfo returns the basic array metadata
func (client *Client) GetArrayInfo() (*ArrayInfoResponse, error) {
	url := client.createFullURL(ArrayEndpoint)
	response, _, err := client.performGet(url, ArrayInfoResponse{})
	if err != nil {
		return nil, err
	}

	result := response.(*ArrayInfoResponse)
	return result, nil
}

// GetArrayPerformanceMetrics returns all capacity metrics for the array
func (client *Client) GetArrayPerformanceMetrics() (*ArrayPerformanceMetricsResponse, error) {
	url := client.createFullURL(ArrayPerformanceMetricsEndpoint)
	response, _, err := client.performGet(url, []ArrayPerformanceMetricsResponse{})
	if err != nil {
		return nil, err
	}

	result := response.(*[]ArrayPerformanceMetricsResponse)
	return &(*result)[0], nil
}

// GetHostCount returns the count of hosts on the array
func (client *Client) GetHostCount() (uint32, error) {
	return client.getResourceCount(HostCountEndpoint)
}

// GetModel returns the model of the primary controller (usually CT0)
func (client *Client) GetModel() (string, error) {
	url := client.createFullURL(ArrayControllersEndpoint)
	response, _, err := client.performGet(url, []ArrayControllersResponse{})
	if err != nil {
		return "", err
	}

	result := response.(*[]ArrayControllersResponse)

	// Return the model of the primary
	for _, controller := range *result {
		if controller.Mode == "primary" {
			return controller.Model, nil
		}
	}

	// If no primary is found, return the model of the first controller
	log.WithFields(log.Fields{
		"display_name": client.DisplayName,
		"url":          url,
	}).Warn("No primary controller found")
	return (*result)[0].Model, nil
}

// GetVolumeCapacityMetrics returns the capacity metrics for all volumes
func (client *Client) GetVolumeCapacityMetrics() ([]*VolumeCapacityMetricsResponse, error) {
	url := client.createFullURL(VolumeCapacityMetricsEndpoint)
	response, _, err := client.performGet(url, []*VolumeCapacityMetricsResponse{})
	if err != nil {
		return nil, err
	}

	result := response.(*[]*VolumeCapacityMetricsResponse)
	return *result, nil
}

// GetVolumeCount returns the count of volumes on the array (without getting all of the volumes)
func (client *Client) GetVolumeCount() (uint32, error) {
	return client.getResourceCount(VolumeCountEndpoint)
}

// GetVolumePerformanceMetrics returns the performance metrics for all volumes
func (client *Client) GetVolumePerformanceMetrics() ([]*VolumePerformanceMetricsResponse, error) {
	url := client.createFullURL(VolumePerformanceMetricsEndpoint)
	response, _, err := client.performGet(url, []*VolumePerformanceMetricsResponse{})
	if err != nil {
		return nil, err
	}

	result := response.(*[]*VolumePerformanceMetricsResponse)
	return *result, nil
}

// GetVolumePendingEradicationCount returns the count of volumes that are pending eradication
func (client *Client) GetVolumePendingEradicationCount() (uint32, error) {
	return client.getResourceCount(VolumePendingEradicationCountEndpoint)
}

// GetVolumeSnapshotCount returns the count of volume snapshots on the array (without getting all of the snapshots)
func (client *Client) GetVolumeSnapshotCount() (uint32, error) {
	return client.getResourceCount(VolumeSnapshotCountEndpoint)
}

// GetVolumeSnapshots returns volume snapshots
func (client *Client) GetVolumeSnapshots() ([]*VolumeSnapshotResponse, error) {
	url := client.createFullURL(VolumeSnapshotsEndpoint)
	response, _, err := client.performGet(url, []*VolumeSnapshotResponse{})
	if err != nil {
		return nil, err
	}

	result := response.(*[]*VolumeSnapshotResponse)
	return *result, nil
}

// createFullURL is a helper function that returns a URL for the specified endpoint/params with the
// management endpoint and API version
func (client *Client) createFullURL(endpoint string) string {
	return fmt.Sprintf("https://%s%s/%s%s", client.ManagementIP.String(), APIPrefix, client.APIVersion, endpoint)
}

// getAlerts is a helper function that returns alerts for the specified messages endpoint
func (client *Client) getAlerts(alertEndpoint string) ([]*AlertResponse, error) {
	url := client.createFullURL(alertEndpoint)
	response, _, err := client.performGet(url, []*AlertResponse{})
	if err != nil {
		return nil, err
	}

	result := response.(*[]*AlertResponse)
	return *result, nil
}

// getAPIVersion is a helper function that checks the available API versions and that the desired
// version is available; it warns if it is not
func (client *Client) getAPIVersion() (string, error) {
	url := fmt.Sprintf("https://%s%s", client.ManagementIP.String(), APIVersionEndpoint)
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

// getResourceCount is a helper function that returns the number of resources where we don't need
// the actual items by getting the count from the response header
func (client *Client) getResourceCount(endpoint string) (uint32, error) {
	url := client.createFullURL(endpoint)
	_, headers, err := client.performGet(url, []EmptyResponse{})
	if err != nil {
		return 0, err
	}

	// Read the count from the header
	countHeader := headers.Get(TotalItemCountHeader)
	count, err := strconv.ParseUint(countHeader, 10, 32) // base 10, 32 bit
	if err != nil {
		log.WithFields(log.Fields{
			"count_header": string(countHeader),
			"display_name": client.DisplayName,
			"url":          url,
		}).Error("Could not get total item count from header")
		return 0, err
	}
	return uint32(count), nil
}

// logCreationHeader is a helper function that logs errors originating from NewClient
func (client *Client) logCreationError() {
	log.WithFields(log.Fields{
		"display_name": client.DisplayName,
	}).Error("Could not create FlashArray Client")
}

// performGet is a helper function that encapsulates exit and retry cases for GET requests
// Returns the unmarshalled response data, response headers, and error
func (client *Client) performGet(url string, result interface{}) (interface{}, http.Header, error) {
	// Each request can be retried multiple times
	for i := 0; i < RequestAttemptCount; i++ {
		log.WithFields(log.Fields{
			"display_name": client.DisplayName,
			"url":          url,
		}).Trace("Making GET request")
		response, err := resty.R().SetHeader(UserAgentHeader, UserAgent).SetResult(result).Get(url)

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
		if response.StatusCode() == 401 {
			log.WithFields(log.Fields{
				"display_name": client.DisplayName,
				"status_code":  response.StatusCode(),
				"url":          url,
			}).Trace("Session expired; refreshing session and retrying")
			client.refreshSession()
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
	}
	// Log we failed
	log.WithFields(log.Fields{
		"display_name": client.DisplayName,
		"url":          url,
	}).Error("No successful GET request")
	return nil, nil, errors.New("No successful GET request")
}

// refreshSession is a helper function that makes a POST request to refresh the client session
func (client *Client) refreshSession() error {
	url := client.createFullURL(SessionEndpoint)
	log.WithFields(log.Fields{
		"display_name": client.DisplayName,
		"url":          url,
	}).Trace("Making POST request to refresh session")
	response, err := resty.R().SetHeader(UserAgentHeader, UserAgent).SetBody(map[string]interface{}{"api_token": client.APIToken}).Post(url)

	// Verify the request was successful
	if err != nil {
		log.WithFields(log.Fields{
			"display_name": client.DisplayName,
			"error":        err,
			"url":          url,
		}).Error("Error making refresh request")
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
	return nil
}
