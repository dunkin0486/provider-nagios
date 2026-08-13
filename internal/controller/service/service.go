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
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/crossplane/crossplane-runtime/v2/pkg/feature"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"

	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"

	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/errors"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/statemetrics"

	nagiosclient "github.com/dunkin0486/provider-nagios/internal/client"
	"github.com/dunkin0486/provider-nagios/internal/clients"

	v1alpha1 "github.com/dunkin0486/provider-nagios/apis/monitoring/v1alpha1"
	apisv1alpha1 "github.com/dunkin0486/provider-nagios/apis/v1alpha1"
)

const (
	errTrackPCUsage = "cannot track ProviderConfig usage"
	errGetPC        = "cannot get ProviderConfig"
	errGetCPC       = "cannot get ClusterProviderConfig"
	errGetCreds     = "cannot get credentials"
	errNewClient    = "cannot create new Nagios client"

	errGetService    = "cannot get service"
	errCreateService = "cannot create service"
	errUpdateService = "cannot update service"
	errDeleteService = "cannot delete service"

	// createRetryAttempts/createRetryBackoff tolerate Nagios's own
	// eventual-consistency window right after a write - see
	// nagiosclient.RetryUntilFound.
	createRetryAttempts = 5
	createRetryBackoff  = 2 * time.Second
)

// SetupGated adds a controller that reconciles Service managed resources with safe-start support.
func SetupGated(mgr ctrl.Manager, o controller.Options) error {
	o.Gate.Register(func() {
		if err := Setup(mgr, o); err != nil {
			panic(errors.Wrap(err, "cannot setup Service controller"))
		}
	}, v1alpha1.ServiceGroupVersionKind)
	return nil
}

func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.ServiceGroupKind)

	opts := []managed.ReconcilerOption{
		managed.WithTypedExternalConnector[*v1alpha1.Service](&connector{
			kube:         mgr.GetClient(),
			usage:        resource.NewProviderConfigUsageTracker(mgr.GetClient(), &apisv1alpha1.ProviderConfigUsage{}),
			newServiceFn: clients.NewClient}),
		managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient())),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithPollInterval(o.PollInterval),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))), //nolint:staticcheck // TODO(jbw976) Crossplane needs to update to the new events API, see https://github.com/crossplane/crossplane/issues/7152
	}

	if o.Features.Enabled(feature.EnableBetaManagementPolicies) {
		opts = append(opts, managed.WithManagementPolicies())
	}

	if o.Features.Enabled(feature.EnableAlphaChangeLogs) {
		opts = append(opts, managed.WithChangeLogger(o.ChangeLogOptions.ChangeLogger))
	}

	if o.MetricOptions != nil {
		opts = append(opts, managed.WithMetricRecorder(o.MetricOptions.MRMetrics))
	}

	if o.MetricOptions != nil && o.MetricOptions.MRStateMetrics != nil {
		stateMetricsRecorder := statemetrics.NewMRStateRecorder(
			mgr.GetClient(), o.Logger, o.MetricOptions.MRStateMetrics, &v1alpha1.ServiceList{}, o.MetricOptions.PollStateMetricInterval,
		)
		if err := mgr.Add(stateMetricsRecorder); err != nil {
			return errors.Wrap(err, "cannot register MR state metrics recorder for kind v1alpha1.ServiceList")
		}
	}

	r := managed.NewReconciler(mgr, resource.ManagedKind(v1alpha1.ServiceGroupVersionKind), opts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1alpha1.Service{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube         client.Client
	usage        *resource.ProviderConfigUsageTracker
	newServiceFn func(creds []byte) (*nagiosclient.Client, error)
}

// Connect typically produces an ExternalClient by:
// 1. Tracking that the managed resource is using a ProviderConfig.
// 2. Getting the managed resource's ProviderConfig.
// 3. Getting the credentials specified by the ProviderConfig.
// 4. Using the credentials to form a client.
func (c *connector) Connect(ctx context.Context, cr *v1alpha1.Service) (managed.TypedExternalClient[*v1alpha1.Service], error) {
	if err := c.usage.Track(ctx, cr); err != nil {
		return nil, errors.Wrap(err, errTrackPCUsage)
	}

	var cd apisv1alpha1.ProviderCredentials

	ref := cr.GetProviderConfigReference()

	switch ref.Kind {
	case "ProviderConfig":
		pc := &apisv1alpha1.ProviderConfig{}
		if err := c.kube.Get(ctx, types.NamespacedName{Name: ref.Name, Namespace: cr.GetNamespace()}, pc); err != nil {
			return nil, errors.Wrap(err, errGetPC)
		}
		cd = pc.Spec.Credentials
	case "ClusterProviderConfig":
		cpc := &apisv1alpha1.ClusterProviderConfig{}
		if err := c.kube.Get(ctx, types.NamespacedName{Name: ref.Name}, cpc); err != nil {
			return nil, errors.Wrap(err, errGetCPC)
		}
		cd = cpc.Spec.Credentials
	default:
		return nil, errors.Errorf("unsupported provider config kind: %s", ref.Kind)
	}
	data, err := resource.CommonCredentialExtractor(ctx, cd.Source, c.kube, cd.CommonCredentialSelectors)
	if err != nil {
		return nil, errors.Wrap(err, errGetCreds)
	}

	svc, err := c.newServiceFn(data)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}

	return &external{client: svc}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	client *nagiosclient.Client
}

func (c *external) Observe(ctx context.Context, cr *v1alpha1.Service) (managed.ExternalObservation, error) {
	name := meta.GetExternalName(cr)
	if name == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	s, err := c.client.GetService(ctx, name)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetService)
	}
	if s == nil {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	cr.Status.AtProvider = observationFromService(s)
	cr.SetConditions(xpv2.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: isUpToDate(serviceFromParameters(name, cr.Spec.ForProvider), s),
	}, nil
}

// isUpToDate compares the service we'd create from spec against the service
// Nagios actually reports. Register is excluded because it is a
// Nagios-managed field we never set, not a drift signal - comparing it
// would put every Service in a permanent Update loop.
func isUpToDate(desired, current *nagiosclient.Service) bool {
	return cmp.Equal(desired, current,
		cmpopts.IgnoreFields(nagiosclient.Service{}, "Register"),
		cmpopts.EquateEmpty(),
		cmpopts.SortSlices(func(a, b string) bool { return a < b }),
	)
}

func (c *external) Create(ctx context.Context, cr *v1alpha1.Service) (managed.ExternalCreation, error) {
	name := meta.GetExternalName(cr)
	s := serviceFromParameters(name, cr.Spec.ForProvider)

	if err := c.client.NewService(ctx, s); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateService)
	}

	if _, err := nagiosclient.RetryUntilFound(ctx, createRetryAttempts, createRetryBackoff, func() (*nagiosclient.Service, error) {
		return c.client.GetService(ctx, name)
	}); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateService)
	}

	return managed.ExternalCreation{}, nil
}

// Update addresses the service being updated by (oldServiceName,
// oldDescription) - Nagios's PUT endpoint is compound-keyed, unlike Host's
// name-only PUT (see internal/client.Service's doc comment). oldDescription
// comes from cr.Status.AtProvider, which Observe populated earlier in this
// same reconcile - not from cr.Spec.ForProvider.Description, which may
// already hold the caller's newly desired value.
func (c *external) Update(ctx context.Context, cr *v1alpha1.Service) (managed.ExternalUpdate, error) {
	name := meta.GetExternalName(cr)
	s := serviceFromParameters(name, cr.Spec.ForProvider)

	if err := c.client.UpdateService(ctx, s, name, cr.Status.AtProvider.Description); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateService)
	}

	return managed.ExternalUpdate{}, nil
}

// Delete addresses the service being deleted by (host_name, description),
// distinct from the (config_name, description) key Update uses (see
// internal/client.Service's doc comment). Both values are read from
// cr.Status.AtProvider - populated by the Observe call that immediately
// precedes Delete in the same reconcile - rather than cr.Spec.ForProvider,
// since it's the currently-live association in Nagios that must be
// addressed, not whatever the (possibly already-edited) desired spec says.
func (c *external) Delete(ctx context.Context, cr *v1alpha1.Service) (managed.ExternalDelete, error) {
	if err := c.client.DeleteService(ctx, cr.Status.AtProvider.HostName, cr.Status.AtProvider.Description); err != nil {
		return managed.ExternalDelete{}, errors.Wrap(err, errDeleteService)
	}

	return managed.ExternalDelete{}, nil
}

func (c *external) Disconnect(_ context.Context) error {
	return nil
}
