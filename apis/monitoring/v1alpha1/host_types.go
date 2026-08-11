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

// HostParameters are the configurable fields of a Host. The Nagios host_name
// is not a field here - it is taken from the resource's external-name
// annotation, defaulted to metadata.name if unset.
//
// Field shapes mirror internal/client.Host's wire types directly (e.g.
// numeric-looking fields stay string) rather than being remapped to more
// idiomatic Go types, since Nagios XI's API itself expects string form for
// these values.
type HostParameters struct {
	// Address is the host's IP address or FQDN.
	Address string `json:"address"`

	// DisplayName overrides how the host is shown in the Nagios UI.
	// +optional
	DisplayName string `json:"displayName,omitempty"`

	// MaxCheckAttempts is the number of times Nagios retries a check before
	// treating the host as truly down.
	MaxCheckAttempts string `json:"maxCheckAttempts"`

	// CheckPeriod is the name of the time period during which active checks
	// of this host can occur.
	CheckPeriod string `json:"checkPeriod"`

	// NotificationInterval is the number of minutes to wait before
	// re-notifying a contact about a problem.
	NotificationInterval string `json:"notificationInterval"`

	// NotificationPeriod is the name of the time period during which
	// notifications about this host can be sent out.
	NotificationPeriod string `json:"notificationPeriod"`

	// Contacts are the short names of contacts that should be notified
	// whenever there are problems with this host.
	// +optional
	Contacts []string `json:"contacts,omitempty"`

	// ContactGroups are the short names of contact groups that should be
	// notified whenever there are problems with this host.
	// +optional
	ContactGroups []string `json:"contactGroups,omitempty"`

	// Alias is a longer name/description for the host.
	// +optional
	Alias string `json:"alias,omitempty"`

	// Templates are the names of host templates this host should inherit
	// from.
	// +optional
	Templates []string `json:"templates,omitempty"`

	// CheckCommand is the short name of the command used to check the
	// status of the host.
	// +optional
	CheckCommand string `json:"checkCommand,omitempty"`

	// Parents lists the short names of hosts this host depends on for
	// network reachability.
	// +optional
	Parents []string `json:"parents,omitempty"`

	// Hostgroups are the short names of hostgroups this host should be a
	// member of.
	// +optional
	Hostgroups []string `json:"hostgroups,omitempty"`

	// Notes are additional notes about the host.
	// +optional
	Notes string `json:"notes,omitempty"`

	// NotesURL is a URL for additional notes about the host.
	// +optional
	NotesURL string `json:"notesUrl,omitempty"`

	// ActionURL is a URL for actions to be performed on the host.
	// +optional
	ActionURL string `json:"actionUrl,omitempty"`

	// RetryInterval is the number of minutes to wait before scheduling a
	// re-check of the host after a problem is detected.
	// +optional
	RetryInterval string `json:"retryInterval,omitempty"`

	// ActiveChecksEnabled determines whether active checks of the host are
	// enabled.
	// +optional
	ActiveChecksEnabled *bool `json:"activeChecksEnabled,omitempty"`

	// PassiveChecksEnabled determines whether passive checks of the host
	// are enabled.
	// +optional
	PassiveChecksEnabled *bool `json:"passiveChecksEnabled,omitempty"`

	// EventHandler is the short name of the command run whenever a state
	// change occurs for the host.
	// +optional
	EventHandler string `json:"eventHandler,omitempty"`

	// EventHandlerEnabled determines whether the event handler is enabled.
	// +optional
	EventHandlerEnabled *bool `json:"eventHandlerEnabled,omitempty"`

	// FlapDetectionEnabled determines whether flap detection is enabled.
	// +optional
	FlapDetectionEnabled *bool `json:"flapDetectionEnabled,omitempty"`

	// FlapDetectionOptions are the host states used in flap detection.
	// +optional
	FlapDetectionOptions []string `json:"flapDetectionOptions,omitempty"`

	// LowFlapThreshold is the low state change threshold used in flap
	// detection.
	// +optional
	LowFlapThreshold string `json:"lowFlapThreshold,omitempty"`

	// HighFlapThreshold is the high state change threshold used in flap
	// detection.
	// +optional
	HighFlapThreshold string `json:"highFlapThreshold,omitempty"`

	// ProcessPerfData determines whether performance data processing is
	// enabled.
	// +optional
	ProcessPerfData *bool `json:"processPerfData,omitempty"`

	// RetainStatusInformation determines whether status-related information
	// about the host is retained across program restarts.
	// +optional
	RetainStatusInformation *bool `json:"retainStatusInformation,omitempty"`

	// RetainNonstatusInformation determines whether non-status information
	// about the host is retained across program restarts.
	// +optional
	RetainNonstatusInformation *bool `json:"retainNonstatusInformation,omitempty"`

	// CheckFreshness determines whether freshness checks are enabled for
	// this host.
	// +optional
	CheckFreshness *bool `json:"checkFreshness,omitempty"`

	// FreshnessThreshold is the freshness threshold in seconds for this
	// host.
	// +optional
	FreshnessThreshold string `json:"freshnessThreshold,omitempty"`

	// FirstNotificationDelay is the number of minutes to wait before
	// sending the first problem notification.
	// +optional
	FirstNotificationDelay string `json:"firstNotificationDelay,omitempty"`

	// NotificationOptions are the states for which notifications should be
	// sent out, comma-joined (e.g. "d,u,r").
	// +optional
	NotificationOptions string `json:"notificationOptions,omitempty"`

	// NotificationsEnabled determines whether notifications for the host
	// are enabled.
	// +optional
	NotificationsEnabled *bool `json:"notificationsEnabled,omitempty"`

	// StalkingOptions are the states for which stalking should be enabled.
	// +optional
	StalkingOptions string `json:"stalkingOptions,omitempty"`

	// IconImage is the image used in the Nagios UI's status/map views.
	// +optional
	IconImage string `json:"iconImage,omitempty"`

	// IconImageAlt is the alt text for IconImage.
	// +optional
	IconImageAlt string `json:"iconImageAlt,omitempty"`

	// VRMLImage is the image used for the host in 3D status map views.
	// +optional
	VRMLImage string `json:"vrmlImage,omitempty"`

	// StatusMapImage is the image used for the host in 2D status map
	// views.
	// +optional
	StatusMapImage string `json:"statusMapImage,omitempty"`

	// CoordsTwoD are the host's 2D status map coordinates ("x,y").
	// +optional
	CoordsTwoD string `json:"coords2d,omitempty"`

	// CoordsThreeD are the host's 3D status map coordinates ("x,y,z").
	// +optional
	CoordsThreeD string `json:"coords3d,omitempty"`

	// FreeVariables are custom `_`-prefixed macros attached to the host.
	// +optional
	FreeVariables map[string]string `json:"freeVariables,omitempty"`
}

// HostObservation are the observable fields of a Host, as last read back
// from Nagios XI.
type HostObservation struct {
	Address                    string            `json:"address,omitempty"`
	DisplayName                string            `json:"displayName,omitempty"`
	MaxCheckAttempts           string            `json:"maxCheckAttempts,omitempty"`
	CheckPeriod                string            `json:"checkPeriod,omitempty"`
	NotificationInterval       string            `json:"notificationInterval,omitempty"`
	NotificationPeriod         string            `json:"notificationPeriod,omitempty"`
	Contacts                   []string          `json:"contacts,omitempty"`
	ContactGroups              []string          `json:"contactGroups,omitempty"`
	Alias                      string            `json:"alias,omitempty"`
	Templates                  []string          `json:"templates,omitempty"`
	CheckCommand               string            `json:"checkCommand,omitempty"`
	Parents                    []string          `json:"parents,omitempty"`
	Hostgroups                 []string          `json:"hostgroups,omitempty"`
	Notes                      string            `json:"notes,omitempty"`
	NotesURL                   string            `json:"notesUrl,omitempty"`
	ActionURL                  string            `json:"actionUrl,omitempty"`
	RetryInterval              string            `json:"retryInterval,omitempty"`
	ActiveChecksEnabled        *bool             `json:"activeChecksEnabled,omitempty"`
	PassiveChecksEnabled       *bool             `json:"passiveChecksEnabled,omitempty"`
	EventHandler               string            `json:"eventHandler,omitempty"`
	EventHandlerEnabled        *bool             `json:"eventHandlerEnabled,omitempty"`
	FlapDetectionEnabled       *bool             `json:"flapDetectionEnabled,omitempty"`
	FlapDetectionOptions       []string          `json:"flapDetectionOptions,omitempty"`
	LowFlapThreshold           string            `json:"lowFlapThreshold,omitempty"`
	HighFlapThreshold          string            `json:"highFlapThreshold,omitempty"`
	ProcessPerfData            *bool             `json:"processPerfData,omitempty"`
	RetainStatusInformation    *bool             `json:"retainStatusInformation,omitempty"`
	RetainNonstatusInformation *bool             `json:"retainNonstatusInformation,omitempty"`
	CheckFreshness             *bool             `json:"checkFreshness,omitempty"`
	FreshnessThreshold         string            `json:"freshnessThreshold,omitempty"`
	FirstNotificationDelay     string            `json:"firstNotificationDelay,omitempty"`
	NotificationOptions        string            `json:"notificationOptions,omitempty"`
	NotificationsEnabled       *bool             `json:"notificationsEnabled,omitempty"`
	StalkingOptions            string            `json:"stalkingOptions,omitempty"`
	IconImage                  string            `json:"iconImage,omitempty"`
	IconImageAlt               string            `json:"iconImageAlt,omitempty"`
	VRMLImage                  string            `json:"vrmlImage,omitempty"`
	StatusMapImage             string            `json:"statusMapImage,omitempty"`
	CoordsTwoD                 string            `json:"coords2d,omitempty"`
	CoordsThreeD               string            `json:"coords3d,omitempty"`
	FreeVariables              map[string]string `json:"freeVariables,omitempty"`
}

// A HostSpec defines the desired state of a Host.
type HostSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`
	ForProvider              HostParameters `json:"forProvider"`
}

// A HostStatus represents the observed state of a Host.
type HostStatus struct {
	xpv2.ManagedResourceStatus `json:",inline"`
	AtProvider                 HostObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Host is a Nagios XI host object: a monitored device or system.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,nagios}
type Host struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HostSpec   `json:"spec"`
	Status HostStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HostList contains a list of Host
type HostList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Host `json:"items"`
}

// Host type metadata.
var (
	HostKind             = reflect.TypeOf(Host{}).Name()
	HostGroupKind        = schema.GroupKind{Group: Group, Kind: HostKind}.String()
	HostKindAPIVersion   = HostKind + "." + SchemeGroupVersion.String()
	HostGroupVersionKind = SchemeGroupVersion.WithKind(HostKind)
)

func init() {
	SchemeBuilder.Register(&Host{}, &HostList{})
}
