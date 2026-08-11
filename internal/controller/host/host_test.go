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

package host

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	nagiosclient "github.com/dunkin0486/provider-nagios/internal/client"

	v1alpha1 "github.com/dunkin0486/provider-nagios/apis/monitoring/v1alpha1"
)

// Unlike many Kubernetes projects Crossplane does not use third party testing
// libraries, per the common Go test review comments. Crossplane encourages the
// use of table driven unit tests. The tests of the crossplane-runtime project
// are representative of the testing style Crossplane encourages.
//
// https://github.com/golang/go/wiki/TestComments
// https://github.com/crossplane/crossplane/blob/master/CONTRIBUTING.md#contributing-code

func hostWithExternalName(name string, p v1alpha1.HostParameters) *v1alpha1.Host {
	cr := &v1alpha1.Host{
		ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{}},
		Spec:       v1alpha1.HostSpec{ForProvider: p},
	}
	meta.SetExternalName(cr, name)
	return cr
}

func TestObserve_NoExternalName(t *testing.T) {
	e := external{client: nagiosclient.NewClient("http://unused.invalid", "TOKEN")}
	cr := &v1alpha1.Host{}

	got, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ResourceExists {
		t.Error("expected ResourceExists=false when external-name is unset")
	}
}

func TestObserve_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	e := external{client: nagiosclient.NewClient(srv.URL, "TOKEN")}
	cr := hostWithExternalName("myhost", v1alpha1.HostParameters{Address: "127.0.0.1"})

	got, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ResourceExists {
		t.Error("expected ResourceExists=false for a host Nagios doesn't know about")
	}
}

func TestObserve_FoundAndUpToDate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"host_name":"myhost","address":"127.0.0.1","max_check_attempts":"3","check_period":"24x7","notification_interval":"30","notification_period":"24x7"}]`))
	}))
	defer srv.Close()

	e := external{client: nagiosclient.NewClient(srv.URL, "TOKEN")}
	cr := hostWithExternalName("myhost", v1alpha1.HostParameters{
		Address:              "127.0.0.1",
		MaxCheckAttempts:     "3",
		CheckPeriod:          "24x7",
		NotificationInterval: "30",
		NotificationPeriod:   "24x7",
	})

	got, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Observe(...): -want, +got:\n%s", diff)
	}
	if cr.Status.AtProvider.Address != "127.0.0.1" {
		t.Errorf("AtProvider.Address = %q, want %q", cr.Status.AtProvider.Address, "127.0.0.1")
	}
}

func TestObserve_FoundButOutOfDate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"host_name":"myhost","address":"127.0.0.1","max_check_attempts":"3","check_period":"24x7","notification_interval":"30","notification_period":"24x7"}]`))
	}))
	defer srv.Close()

	e := external{client: nagiosclient.NewClient(srv.URL, "TOKEN")}
	cr := hostWithExternalName("myhost", v1alpha1.HostParameters{
		Address:              "10.0.0.9", // drifted from the "127.0.0.1" Nagios reports
		MaxCheckAttempts:     "3",
		CheckPeriod:          "24x7",
		NotificationInterval: "30",
		NotificationPeriod:   "24x7",
	})

	got, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.ResourceExists {
		t.Fatal("expected ResourceExists=true")
	}
	if got.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=false when address has drifted")
	}
}
