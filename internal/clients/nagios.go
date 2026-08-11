/*
Copyright 2025 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package clients constructs a Nagios XI API client from ProviderConfig
// credentials.
package clients

import (
	"encoding/json"
	"fmt"

	"github.com/dunkin0486/provider-nagios/internal/client"
)

// Credentials is the JSON shape expected in the Secret referenced by a
// ProviderConfig or ClusterProviderConfig:
//
//	{"url": "http://nagios.example.com/nagiosxi", "token": "..."}
type Credentials struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

// NewClient parses ProviderConfig credential data and returns a client
// configured to talk to the referenced Nagios XI instance.
func NewClient(data []byte) (*client.Client, error) {
	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("cannot parse nagios credentials: %w", err)
	}
	if creds.URL == "" {
		return nil, fmt.Errorf("nagios credentials missing required field %q", "url")
	}
	if creds.Token == "" {
		return nil, fmt.Errorf("nagios credentials missing required field %q", "token")
	}
	return client.NewClient(creds.URL, creds.Token), nil
}
