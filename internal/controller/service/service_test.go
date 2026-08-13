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

package service

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

const (
	testHost1       = "host1"
	testDescription = "CPU Load"
)

// Unlike many Kubernetes projects Crossplane does not use third party testing
// libraries, per the common Go test review comments. Crossplane encourages the
// use of table driven unit tests. The tests of the crossplane-runtime project
// are representative of the testing style Crossplane encourages.
//
// https://github.com/golang/go/wiki/TestComments
// https://github.com/crossplane/crossplane/blob/master/CONTRIBUTING.md#contributing-code

func serviceWithExternalName(name string, p v1alpha1.ServiceParameters) *v1alpha1.Service {
	cr := &v1alpha1.Service{
		ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{}},
		Spec:       v1alpha1.ServiceSpec{ForProvider: p},
	}
	meta.SetExternalName(cr, name)
	return cr
}

func TestObserve_NoExternalName(t *testing.T) {
	e := external{client: nagiosclient.NewClient("http://unused.invalid", "TOKEN")}
	cr := &v1alpha1.Service{}

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
	cr := serviceWithExternalName("my_service", v1alpha1.ServiceParameters{
		HostName:    []string{testHost1},
		Description: testDescription,
	})

	got, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ResourceExists {
		t.Error("expected ResourceExists=false for a service Nagios doesn't know about")
	}
}

func TestObserve_FoundAndUpToDate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"config_name":"my_service","host_name":["host1"],"service_description":"CPU Load","check_command":"check_load","max_check_attempts":"3","check_interval":"5","retry_interval":"1","check_period":"24x7","notification_interval":"30","notification_period":"24x7"}]`))
	}))
	defer srv.Close()

	e := external{client: nagiosclient.NewClient(srv.URL, "TOKEN")}
	cr := serviceWithExternalName("my_service", v1alpha1.ServiceParameters{
		HostName:             []string{testHost1},
		Description:          testDescription,
		CheckCommand:         "check_load",
		MaxCheckAttempts:     "3",
		CheckInterval:        "5",
		RetryInterval:        "1",
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
	if cr.Status.AtProvider.Description != testDescription {
		t.Errorf("AtProvider.Description = %q, want %q", cr.Status.AtProvider.Description, testDescription)
	}
}

func TestObserve_FoundButOutOfDate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"config_name":"my_service","host_name":["host1"],"service_description":"CPU Load","check_command":"check_load","max_check_attempts":"3","check_interval":"5","retry_interval":"1","check_period":"24x7","notification_interval":"30","notification_period":"24x7"}]`))
	}))
	defer srv.Close()

	e := external{client: nagiosclient.NewClient(srv.URL, "TOKEN")}
	cr := serviceWithExternalName("my_service", v1alpha1.ServiceParameters{
		HostName:             []string{testHost1},
		Description:          testDescription,
		CheckCommand:         "check_load",
		MaxCheckAttempts:     "5", // drifted from the "3" Nagios reports
		CheckInterval:        "5",
		RetryInterval:        "1",
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
		t.Error("expected ResourceUpToDate=false when max_check_attempts has drifted")
	}
}

// TestUpdate_AddressesOldDescriptionFromAtProvider confirms Update sources
// the PUT's compound-key oldDescription from cr.Status.AtProvider (last
// observed by a preceding Observe call in the same reconcile), not from the
// desired cr.Spec.ForProvider.Description - which may already hold the
// caller's newly-requested value. See internal/client.Service's doc comment
// for why the PUT key must reference the description Nagios still has on
// record, not the one being requested.
func TestUpdate_AddressesOldDescriptionFromAtProvider(t *testing.T) {
	var gotPUTPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPut:
			gotPUTPath = r.URL.Path
			_, _ = w.Write([]byte(`{"success":"Service successfully updated"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/system/applyconfig":
			_, _ = w.Write([]byte(`{"success":"ok"}`))
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer srv.Close()

	e := external{client: nagiosclient.NewClient(srv.URL, "TOKEN")}
	cr := serviceWithExternalName("my_service", v1alpha1.ServiceParameters{
		HostName:    []string{testHost1},
		Description: "CPU Load v2", // the newly desired description
	})
	cr.Status.AtProvider.Description = testDescription // what Observe last saw

	if _, err := e.Update(context.Background(), cr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "/api/v1/config/service/my_service/CPU Load"
	if gotPUTPath != want {
		t.Errorf("PUT path = %q, want %q (must address the old description, not the new one)", gotPUTPath, want)
	}
}

// TestDelete_AddressesLiveHostAssociationFromAtProvider confirms Delete
// sources both the host set and description from cr.Status.AtProvider -
// what's actually live in Nagios - rather than cr.Spec.ForProvider, which
// may have already drifted from what Nagios has on record.
func TestDelete_AddressesLiveHostAssociationFromAtProvider(t *testing.T) {
	var gotHostName, gotDescription string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodDelete:
			gotHostName = r.URL.Query().Get("host_name")
			gotDescription = r.URL.Query().Get("service_description")
			_, _ = w.Write([]byte(`{"success":"Service successfully deleted"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/system/applyconfig":
			_, _ = w.Write([]byte(`{"success":"ok"}`))
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer srv.Close()

	e := external{client: nagiosclient.NewClient(srv.URL, "TOKEN")}
	cr := serviceWithExternalName("my_service", v1alpha1.ServiceParameters{
		HostName:    []string{"host3"}, // desired spec has already changed...
		Description: "CPU Load v2",
	})
	cr.Status.AtProvider.HostName = []string{testHost1, "host2"} // ...but this is what's live
	cr.Status.AtProvider.Description = testDescription

	if _, err := e.Delete(context.Background(), cr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotHostName != "host1,host2" {
		t.Errorf("host_name = %q, want %q", gotHostName, "host1,host2")
	}
	if gotDescription != testDescription {
		t.Errorf("service_description = %q, want %q", gotDescription, testDescription)
	}
}
