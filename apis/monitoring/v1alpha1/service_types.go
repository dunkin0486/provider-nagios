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

package v1alpha1

import (
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"
)

// ServiceParameters are the configurable fields of a Service. The Nagios
// config_name is not a field here - it is taken from the resource's
// external-name annotation, defaulted to metadata.name if unset, mirroring
// HostParameters.
//
// Field shapes mirror internal/client.Service's wire types directly, with
// one deliberate exception preserved from the ported client: NotificationOptions
// is a []string here (and on the wire), while the equivalent field on Host is
// a comma-joined string - a real, intentionally-preserved API asymmetry, not
// a mistake to "fix" into consistency.
type ServiceParameters struct {
	// HostName lists the short names of the hosts this service is attached
	// to. Nagios's DELETE endpoint addresses the service by this full set
	// (comma-joined) rather than by config_name - see internal/client.Service's
	// doc comment.
	HostName []string `json:"hostName"`

	// HostgroupName are the short names of hostgroups this service should
	// additionally be applied to.
	// +optional
	HostgroupName []string `json:"hostgroupName,omitempty"`

	// DisplayName overrides how the service is shown in the Nagios UI.
	// +optional
	DisplayName string `json:"displayName,omitempty"`

	// Description is the service's description (Nagios's
	// service_description). Together with the external-name (config_name)
	// it forms the compound key Nagios's PUT endpoint addresses a service
	// by.
	Description string `json:"description"`

	// CheckCommand is the short name of the command used to check the
	// status of the service.
	CheckCommand string `json:"checkCommand"`

	// MaxCheckAttempts is the number of times Nagios retries a check before
	// treating the service as truly in a problem state.
	MaxCheckAttempts string `json:"maxCheckAttempts"`

	// CheckInterval is the number of minutes between regularly scheduled
	// checks of the service.
	CheckInterval string `json:"checkInterval"`

	// RetryInterval is the number of minutes to wait before scheduling a
	// re-check of the service after a problem is detected.
	RetryInterval string `json:"retryInterval"`

	// CheckPeriod is the name of the time period during which active checks
	// of this service can occur.
	CheckPeriod string `json:"checkPeriod"`

	// NotificationInterval is the number of minutes to wait before
	// re-notifying a contact about a problem.
	NotificationInterval string `json:"notificationInterval"`

	// NotificationPeriod is the name of the time period during which
	// notifications about this service can be sent out.
	NotificationPeriod string `json:"notificationPeriod"`

	// Contacts are the short names of contacts that should be notified
	// whenever there are problems with this service.
	// +optional
	Contacts []string `json:"contacts,omitempty"`

	// Templates are the names of service templates this service should
	// inherit from.
	// +optional
	Templates []string `json:"templates,omitempty"`

	// IsVolatile determines whether the service is treated as volatile.
	// +optional
	IsVolatile *bool `json:"isVolatile,omitempty"`

	// InitialState is the initial state of the service Nagios assumes on
	// startup.
	// +optional
	InitialState string `json:"initialState,omitempty"`

	// ActiveChecksEnabled determines whether active checks of the service
	// are enabled.
	// +optional
	ActiveChecksEnabled *bool `json:"activeChecksEnabled,omitempty"`

	// PassiveChecksEnabled determines whether passive checks of the service
	// are enabled.
	// +optional
	PassiveChecksEnabled *bool `json:"passiveChecksEnabled,omitempty"`

	// ObsessOverService determines whether Nagios "obsesses" over the
	// service for distributed monitoring purposes.
	// +optional
	ObsessOverService *bool `json:"obsessOverService,omitempty"`

	// CheckFreshness determines whether freshness checks are enabled for
	// this service.
	// +optional
	CheckFreshness *bool `json:"checkFreshness,omitempty"`

	// FreshnessThreshold is the freshness threshold in seconds for this
	// service.
	// +optional
	FreshnessThreshold string `json:"freshnessThreshold,omitempty"`

	// EventHandler is the short name of the command run whenever a state
	// change occurs for the service.
	// +optional
	EventHandler string `json:"eventHandler,omitempty"`

	// EventHandlerEnabled determines whether the event handler is enabled.
	// +optional
	EventHandlerEnabled *bool `json:"eventHandlerEnabled,omitempty"`

	// LowFlapThreshold is the low state change threshold used in flap
	// detection.
	// +optional
	LowFlapThreshold string `json:"lowFlapThreshold,omitempty"`

	// HighFlapThreshold is the high state change threshold used in flap
	// detection.
	// +optional
	HighFlapThreshold string `json:"highFlapThreshold,omitempty"`

	// FlapDetectionEnabled determines whether flap detection is enabled.
	// +optional
	FlapDetectionEnabled *bool `json:"flapDetectionEnabled,omitempty"`

	// FlapDetectionOptions are the service states used in flap detection.
	// +optional
	FlapDetectionOptions []string `json:"flapDetectionOptions,omitempty"`

	// ProcessPerfData determines whether performance data processing is
	// enabled.
	// +optional
	ProcessPerfData *bool `json:"processPerfData,omitempty"`

	// RetainStatusInformation determines whether status-related information
	// about the service is retained across program restarts.
	// +optional
	RetainStatusInformation *bool `json:"retainStatusInformation,omitempty"`

	// RetainNonstatusInformation determines whether non-status information
	// about the service is retained across program restarts.
	// +optional
	RetainNonstatusInformation *bool `json:"retainNonstatusInformation,omitempty"`

	// FirstNotificationDelay is the number of minutes to wait before
	// sending the first problem notification.
	// +optional
	FirstNotificationDelay string `json:"firstNotificationDelay,omitempty"`

	// NotificationOptions are the states for which notifications should be
	// sent out (e.g. ["w", "u", "c", "r"]). Unlike Host, this is a list, not
	// a comma-joined string - a real API asymmetry, see the type doc above.
	// +optional
	NotificationOptions []string `json:"notificationOptions,omitempty"`

	// NotificationsEnabled determines whether notifications for the service
	// are enabled.
	// +optional
	NotificationsEnabled *bool `json:"notificationsEnabled,omitempty"`

	// ContactGroups are the short names of contact groups that should be
	// notified whenever there are problems with this service.
	// +optional
	ContactGroups []string `json:"contactGroups,omitempty"`

	// Servicegroups are the short names of servicegroups this service
	// should be a member of.
	// +optional
	Servicegroups []string `json:"servicegroups,omitempty"`

	// Notes are additional notes about the service.
	// +optional
	Notes string `json:"notes,omitempty"`

	// NotesURL is a URL for additional notes about the service.
	// +optional
	NotesURL string `json:"notesUrl,omitempty"`

	// ActionURL is a URL for actions to be performed on the service.
	// +optional
	ActionURL string `json:"actionUrl,omitempty"`

	// IconImage is the image used for the service in the Nagios UI.
	// +optional
	IconImage string `json:"iconImage,omitempty"`

	// IconImageAlt is the alt text for IconImage.
	// +optional
	IconImageAlt string `json:"iconImageAlt,omitempty"`

	// FreeVariables are custom `_`-prefixed macros attached to the service.
	// +optional
	FreeVariables map[string]string `json:"freeVariables,omitempty"`
}

// ServiceObservation are the observable fields of a Service, as last read
// back from Nagios XI.
type ServiceObservation struct {
	HostName                   []string          `json:"hostName,omitempty"`
	HostgroupName              []string          `json:"hostgroupName,omitempty"`
	DisplayName                string            `json:"displayName,omitempty"`
	Description                string            `json:"description,omitempty"`
	CheckCommand               string            `json:"checkCommand,omitempty"`
	MaxCheckAttempts           string            `json:"maxCheckAttempts,omitempty"`
	CheckInterval              string            `json:"checkInterval,omitempty"`
	RetryInterval              string            `json:"retryInterval,omitempty"`
	CheckPeriod                string            `json:"checkPeriod,omitempty"`
	NotificationInterval       string            `json:"notificationInterval,omitempty"`
	NotificationPeriod         string            `json:"notificationPeriod,omitempty"`
	Contacts                   []string          `json:"contacts,omitempty"`
	Templates                  []string          `json:"templates,omitempty"`
	IsVolatile                 *bool             `json:"isVolatile,omitempty"`
	InitialState               string            `json:"initialState,omitempty"`
	ActiveChecksEnabled        *bool             `json:"activeChecksEnabled,omitempty"`
	PassiveChecksEnabled       *bool             `json:"passiveChecksEnabled,omitempty"`
	ObsessOverService          *bool             `json:"obsessOverService,omitempty"`
	CheckFreshness             *bool             `json:"checkFreshness,omitempty"`
	FreshnessThreshold         string            `json:"freshnessThreshold,omitempty"`
	EventHandler               string            `json:"eventHandler,omitempty"`
	EventHandlerEnabled        *bool             `json:"eventHandlerEnabled,omitempty"`
	LowFlapThreshold           string            `json:"lowFlapThreshold,omitempty"`
	HighFlapThreshold          string            `json:"highFlapThreshold,omitempty"`
	FlapDetectionEnabled       *bool             `json:"flapDetectionEnabled,omitempty"`
	FlapDetectionOptions       []string          `json:"flapDetectionOptions,omitempty"`
	ProcessPerfData            *bool             `json:"processPerfData,omitempty"`
	RetainStatusInformation    *bool             `json:"retainStatusInformation,omitempty"`
	RetainNonstatusInformation *bool             `json:"retainNonstatusInformation,omitempty"`
	FirstNotificationDelay     string            `json:"firstNotificationDelay,omitempty"`
	NotificationOptions        []string          `json:"notificationOptions,omitempty"`
	NotificationsEnabled       *bool             `json:"notificationsEnabled,omitempty"`
	ContactGroups              []string          `json:"contactGroups,omitempty"`
	Servicegroups              []string          `json:"servicegroups,omitempty"`
	Notes                      string            `json:"notes,omitempty"`
	NotesURL                   string            `json:"notesUrl,omitempty"`
	ActionURL                  string            `json:"actionUrl,omitempty"`
	IconImage                  string            `json:"iconImage,omitempty"`
	IconImageAlt               string            `json:"iconImageAlt,omitempty"`
	FreeVariables              map[string]string `json:"freeVariables,omitempty"`
}

// A ServiceSpec defines the desired state of a Service.
type ServiceSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`
	ForProvider              ServiceParameters `json:"forProvider"`
}

// A ServiceStatus represents the observed state of a Service.
type ServiceStatus struct {
	xpv2.ManagedResourceStatus `json:",inline"`
	AtProvider                 ServiceObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Service is an example API type.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,nagios}
type Service struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceSpec   `json:"spec"`
	Status ServiceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ServiceList contains a list of Service
type ServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Service `json:"items"`
}

// Service type metadata.
var (
	ServiceKind             = reflect.TypeOf(Service{}).Name()
	ServiceGroupKind        = schema.GroupKind{Group: Group, Kind: ServiceKind}.String()
	ServiceKindAPIVersion   = ServiceKind + "." + SchemeGroupVersion.String()
	ServiceGroupVersionKind = SchemeGroupVersion.WithKind(ServiceKind)
)

func init() {
	SchemeBuilder.Register(&Service{}, &ServiceList{})
}
