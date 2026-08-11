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

package clients

import "testing"

func TestNewClient_Valid(t *testing.T) {
	c, err := NewClient([]byte(`{"url":"http://nagios.example.com/nagiosxi","token":"TOKEN"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected a non-nil client")
	}
}

func TestNewClient_MissingURL(t *testing.T) {
	if _, err := NewClient([]byte(`{"token":"TOKEN"}`)); err == nil {
		t.Error("expected an error when url is missing, got nil")
	}
}

func TestNewClient_MissingToken(t *testing.T) {
	if _, err := NewClient([]byte(`{"url":"http://nagios.example.com/nagiosxi"}`)); err == nil {
		t.Error("expected an error when token is missing, got nil")
	}
}

func TestNewClient_InvalidJSON(t *testing.T) {
	if _, err := NewClient([]byte(`not json`)); err == nil {
		t.Error("expected an error for unparseable credentials, got nil")
	}
}
