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
	nagiosclient "github.com/dunkin0486/provider-nagios/internal/client"

	v1alpha1 "github.com/dunkin0486/provider-nagios/apis/monitoring/v1alpha1"
)

// boolToNagios converts an optional bool to Nagios's "0"/"1" string
// convention, leaving it unset ("") when p is nil. Never dereference an
// optional bool field directly - Go's zero value for an unset pointer would
// otherwise be indistinguishable from an explicit false. See
// internal/client's package doc for the equivalent quirk in the ported
// client, and internal/controller/host/convert.go for the same helper.
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

// serviceFromParameters builds the client.Service payload sent to Nagios for
// name (the resource's external-name, i.e. config_name) and the desired
// spec.
func serviceFromParameters(name string, p v1alpha1.ServiceParameters) *nagiosclient.Service {
	return &nagiosclient.Service{
		ServiceName:                name,
		HostName:                   p.HostName,
		HostgroupName:              p.HostgroupName,
		DisplayName:                p.DisplayName,
		Description:                p.Description,
		CheckCommand:               p.CheckCommand,
		MaxCheckAttempts:           p.MaxCheckAttempts,
		CheckInterval:              p.CheckInterval,
		RetryInterval:              p.RetryInterval,
		CheckPeriod:                p.CheckPeriod,
		NotificationInterval:       p.NotificationInterval,
		NotificationPeriod:         p.NotificationPeriod,
		Contacts:                   p.Contacts,
		Templates:                  p.Templates,
		IsVolatile:                 boolToNagios(p.IsVolatile),
		InitialState:               p.InitialState,
		ActiveChecksEnabled:        boolToNagios(p.ActiveChecksEnabled),
		PassiveChecksEnabled:       boolToNagios(p.PassiveChecksEnabled),
		ObsessOverService:          boolToNagios(p.ObsessOverService),
		CheckFreshness:             boolToNagios(p.CheckFreshness),
		FreshnessThreshold:         p.FreshnessThreshold,
		EventHandler:               p.EventHandler,
		EventHandlerEnabled:        boolToNagios(p.EventHandlerEnabled),
		LowFlapThreshold:           p.LowFlapThreshold,
		HighFlapThreshold:          p.HighFlapThreshold,
		FlapDetectionEnabled:       boolToNagios(p.FlapDetectionEnabled),
		FlapDetectionOptions:       p.FlapDetectionOptions,
		ProcessPerfData:            boolToNagios(p.ProcessPerfData),
		RetainStatusInformation:    boolToNagios(p.RetainStatusInformation),
		RetainNonStatusInformation: boolToNagios(p.RetainNonstatusInformation),
		FirstNotificationDelay:     p.FirstNotificationDelay,
		NotificationOptions:        p.NotificationOptions,
		NotificationsEnabled:       boolToNagios(p.NotificationsEnabled),
		ContactGroups:              p.ContactGroups,
		Servicegroups:              p.Servicegroups,
		Notes:                      p.Notes,
		NotesURL:                   p.NotesURL,
		ActionURL:                  p.ActionURL,
		IconImage:                  p.IconImage,
		IconImageAlt:               p.IconImageAlt,
		FreeVariables:              p.FreeVariables,
	}
}

// observationFromService converts a Service fetched from Nagios into the
// resource's observed status.
func observationFromService(s *nagiosclient.Service) v1alpha1.ServiceObservation {
	return v1alpha1.ServiceObservation{
		HostName:                   s.HostName,
		HostgroupName:              s.HostgroupName,
		DisplayName:                s.DisplayName,
		Description:                s.Description,
		CheckCommand:               s.CheckCommand,
		MaxCheckAttempts:           s.MaxCheckAttempts,
		CheckInterval:              s.CheckInterval,
		RetryInterval:              s.RetryInterval,
		CheckPeriod:                s.CheckPeriod,
		NotificationInterval:       s.NotificationInterval,
		NotificationPeriod:         s.NotificationPeriod,
		Contacts:                   s.Contacts,
		Templates:                  s.Templates,
		IsVolatile:                 nagiosToBool(s.IsVolatile),
		InitialState:               s.InitialState,
		ActiveChecksEnabled:        nagiosToBool(s.ActiveChecksEnabled),
		PassiveChecksEnabled:       nagiosToBool(s.PassiveChecksEnabled),
		ObsessOverService:          nagiosToBool(s.ObsessOverService),
		CheckFreshness:             nagiosToBool(s.CheckFreshness),
		FreshnessThreshold:         s.FreshnessThreshold,
		EventHandler:               s.EventHandler,
		EventHandlerEnabled:        nagiosToBool(s.EventHandlerEnabled),
		LowFlapThreshold:           s.LowFlapThreshold,
		HighFlapThreshold:          s.HighFlapThreshold,
		FlapDetectionEnabled:       nagiosToBool(s.FlapDetectionEnabled),
		FlapDetectionOptions:       s.FlapDetectionOptions,
		ProcessPerfData:            nagiosToBool(s.ProcessPerfData),
		RetainStatusInformation:    nagiosToBool(s.RetainStatusInformation),
		RetainNonstatusInformation: nagiosToBool(s.RetainNonStatusInformation),
		FirstNotificationDelay:     s.FirstNotificationDelay,
		NotificationOptions:        s.NotificationOptions,
		NotificationsEnabled:       nagiosToBool(s.NotificationsEnabled),
		ContactGroups:              s.ContactGroups,
		Servicegroups:              s.Servicegroups,
		Notes:                      s.Notes,
		NotesURL:                   s.NotesURL,
		ActionURL:                  s.ActionURL,
		IconImage:                  s.IconImage,
		IconImageAlt:               s.IconImageAlt,
		FreeVariables:              s.FreeVariables,
	}
}
