package chartconfig

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	clusterGuestConfig, err := r.toClusterGuestConfigFunc(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	chartConfigsToUpdate, err := toChartConfigs(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(chartConfigsToUpdate) > 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "updating chartconfigs")

		clusterConfig, err := prepareClusterConfig(r.baseClusterConfig, clusterGuestConfig)
		if err != nil {
			return microerror.Mask(err)
		}

		g8sClient, err := r.guest.NewG8sClient(ctx, clusterConfig.ClusterID, clusterConfig.Domain.API)
		if err != nil {
			return microerror.Mask(err)
		}

		for _, chartConfigToUpdate := range chartConfigsToUpdate {
			_, err := g8sClient.CoreV1alpha1().ChartConfigs(metav1.NamespaceSystem).Update(chartConfigToUpdate)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "updated chartconfigs")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "no need to update chartconfigs")
	}

	return nil
}

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	create, err := r.newCreateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	update, err := r.newUpdateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	delete, err := r.newDeleteChangeForUpdatePatch(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := controller.NewPatch()
	patch.SetCreateChange(create)
	patch.SetUpdateChange(update)
	patch.SetDeleteChange(delete)

	return patch, nil
}

func (r *Resource) newUpdateChange(ctx context.Context, obj, currentState, desiredState interface{}) ([]*v1alpha1.ChartConfig, error) {
	currentChartConfigs, err := toChartConfigs(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredChartConfigs, err := toChartConfigs(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "finding out which chartconfigs have to be updated")

	chartConfigsToUpdate := make([]*v1alpha1.ChartConfig, 0)

	for _, currentChartConfig := range currentChartConfigs {
		desiredChartConfig, err := getChartConfigByName(desiredChartConfigs, currentChartConfig.Name)
		if IsNotFound(err) {
			// Ignore here. These are handled by newDeleteChangeForUpdatePatch().
			continue
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		if isChartConfigModified(desiredChartConfig, currentChartConfig) {
			chartConfigsToUpdate = append(chartConfigsToUpdate, desiredChartConfig)

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found chartconfig '%s' that has to be updated", desiredChartConfig.GetName()))
		}
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d chartconfigs which have to be updated", len(chartConfigsToUpdate)))

	return chartConfigsToUpdate, nil
}