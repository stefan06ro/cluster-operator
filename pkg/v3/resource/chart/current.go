package chart

import (
	"context"

	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"

	"github.com/giantswarm/cluster-operator/pkg/v3/guestcluster"
)

// GetCurrentState gets the state of the chart in the guest cluster.
func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	guestHelmClient, err := r.getGuestHelmClient(ctx, obj)
	if guestcluster.IsNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "did not get a Helm client for the guest cluster")

		// We can't continue without a Helm client. We will retry during the
		// next execution.
		resourcecanceledcontext.SetCanceled(ctx)
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource reconciliation for custom object")

		return nil, nil
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	releaseContent, err := guestHelmClient.GetReleaseContent(chartOperatorRelease)
	if helmclient.IsReleaseNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the chart-operator chart in the guest cluster")
		return nil, nil

	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	releaseHistory, err := guestHelmClient.GetReleaseHistory(chartOperatorRelease)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	chartState := &ResourceState{
		ChartName:      chartOperatorChart,
		ReleaseName:    chartOperatorRelease,
		ReleaseStatus:  releaseContent.Status,
		ReleaseVersion: releaseHistory.Version,
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "found chart-operator chart in the guest cluster")

	return chartState, nil
}