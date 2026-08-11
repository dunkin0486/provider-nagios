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
	nagiosclient "github.com/dunkin0486/provider-nagios/internal/client"

	v1alpha1 "github.com/dunkin0486/provider-nagios/apis/monitoring/v1alpha1"
)

// boolToNagios converts an optional bool to Nagios's "0"/"1" string
// convention, leaving it unset ("") when p is nil. Never dereference an
// optional bool field directly - Go's zero value for an unset pointer would
// otherwise be indistinguishable from an explicit false. See
// internal/client's package doc for the equivalent quirk in the ported
// client.
func boolToNagios(p *bool) string {
	if p == nil {
		return ""
	}
	if *p {
		return "1"
	}
	return "0"
}

// nagiosToBool is boolToNagios's inverse. Any value other than "0"/"1"
// (including "") is treated as unset.
func nagiosToBool(s string) *bool {
	switch s {
	case "1":
		v := true
		return &v
	case "0":
		v := false
		return &v
	default:
		return nil
	}
}

// hostFromParameters builds the client.Host payload sent to Nagios for name
// (the resource's external-name) and the desired spec.
func hostFromParameters(name string, p v1alpha1.HostParameters) *nagiosclient.Host {
	return &nagiosclient.Host{
		HostName:                   name,
		Address:                    p.Address,
		DisplayName:                p.DisplayName,
		MaxCheckAttempts:           p.MaxCheckAttempts,
		CheckPeriod:                p.CheckPeriod,
		NotificationInterval:       p.NotificationInterval,
		NotificationPeriod:         p.NotificationPeriod,
		Contacts:                   p.Contacts,
		ContactGroups:              p.ContactGroups,
		Alias:                      p.Alias,
		Templates:                  p.Templates,
		CheckCommand:               p.CheckCommand,
		Parents:                    p.Parents,
		Hostgroups:                 p.Hostgroups,
		Notes:                      p.Notes,
		NotesURL:                   p.NotesURL,
		ActionURL:                  p.ActionURL,
		RetryInterval:              p.RetryInterval,
		ActiveChecksEnabled:        boolToNagios(p.ActiveChecksEnabled),
		PassiveChecksEnabled:       boolToNagios(p.PassiveChecksEnabled),
		EventHandler:               p.EventHandler,
		EventHandlerEnabled:        boolToNagios(p.EventHandlerEnabled),
		FlapDetectionEnabled:       boolToNagios(p.FlapDetectionEnabled),
		FlapDetectionOptions:       p.FlapDetectionOptions,
		LowFlapThreshold:           p.LowFlapThreshold,
		HighFlapThreshold:          p.HighFlapThreshold,
		ProcessPerfData:            boolToNagios(p.ProcessPerfData),
		RetainStatusInformation:    boolToNagios(p.RetainStatusInformation),
		RetainNonstatusInformation: boolToNagios(p.RetainNonstatusInformation),
		CheckFreshness:             boolToNagios(p.CheckFreshness),
		FreshnessThreshold:         p.FreshnessThreshold,
		FirstNotificationDelay:     p.FirstNotificationDelay,
		NotificationOptions:        p.NotificationOptions,
		NotificationsEnabled:       boolToNagios(p.NotificationsEnabled),
		StalkingOptions:            p.StalkingOptions,
		IconImage:                  p.IconImage,
		IconImageAlt:               p.IconImageAlt,
		VRMLImage:                  p.VRMLImage,
		StatusMapImage:             p.StatusMapImage,
		TwoDCoords:                 p.CoordsTwoD,
		ThreeDCoords:               p.CoordsThreeD,
		FreeVariables:              p.FreeVariables,
	}
}

// observationFromHost converts a Host fetched from Nagios into the
// resource's observed status.
func observationFromHost(h *nagiosclient.Host) v1alpha1.HostObservation {
	return v1alpha1.HostObservation{
		Address:                    h.Address,
		DisplayName:                h.DisplayName,
		MaxCheckAttempts:           h.MaxCheckAttempts,
		CheckPeriod:                h.CheckPeriod,
		NotificationInterval:       h.NotificationInterval,
		NotificationPeriod:         h.NotificationPeriod,
		Contacts:                   h.Contacts,
		ContactGroups:              h.ContactGroups,
		Alias:                      h.Alias,
		Templates:                  h.Templates,
		CheckCommand:               h.CheckCommand,
		Parents:                    h.Parents,
		Hostgroups:                 h.Hostgroups,
		Notes:                      h.Notes,
		NotesURL:                   h.NotesURL,
		ActionURL:                  h.ActionURL,
		RetryInterval:              h.RetryInterval,
		ActiveChecksEnabled:        nagiosToBool(h.ActiveChecksEnabled),
		PassiveChecksEnabled:       nagiosToBool(h.PassiveChecksEnabled),
		EventHandler:               h.EventHandler,
		EventHandlerEnabled:        nagiosToBool(h.EventHandlerEnabled),
		FlapDetectionEnabled:       nagiosToBool(h.FlapDetectionEnabled),
		FlapDetectionOptions:       h.FlapDetectionOptions,
		LowFlapThreshold:           h.LowFlapThreshold,
		HighFlapThreshold:          h.HighFlapThreshold,
		ProcessPerfData:            nagiosToBool(h.ProcessPerfData),
		RetainStatusInformation:    nagiosToBool(h.RetainStatusInformation),
		RetainNonstatusInformation: nagiosToBool(h.RetainNonstatusInformation),
		CheckFreshness:             nagiosToBool(h.CheckFreshness),
		FreshnessThreshold:         h.FreshnessThreshold,
		FirstNotificationDelay:     h.FirstNotificationDelay,
		NotificationOptions:        h.NotificationOptions,
		NotificationsEnabled:       nagiosToBool(h.NotificationsEnabled),
		StalkingOptions:            h.StalkingOptions,
		IconImage:                  h.IconImage,
		IconImageAlt:               h.IconImageAlt,
		VRMLImage:                  h.VRMLImage,
		StatusMapImage:             h.StatusMapImage,
		CoordsTwoD:                 h.TwoDCoords,
		CoordsThreeD:               h.ThreeDCoords,
		FreeVariables:              h.FreeVariables,
	}
}
