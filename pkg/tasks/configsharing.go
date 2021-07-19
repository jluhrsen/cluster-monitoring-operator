// Copyright 2018 The Cluster Monitoring Operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tasks

import (
	"context"
	"net/url"

	"github.com/openshift/cluster-monitoring-operator/pkg/client"
	"github.com/openshift/cluster-monitoring-operator/pkg/manifests"
	"github.com/pkg/errors"
)

type ConfigSharingTask struct {
	client  *client.Client
	factory *manifests.Factory
	config  *manifests.Config
}

func NewConfigSharingTask(client *client.Client, factory *manifests.Factory, config *manifests.Config) *ConfigSharingTask {
	return &ConfigSharingTask{
		client:  client,
		factory: factory,
		config:  config,
	}
}

func (t *ConfigSharingTask) Run(ctx context.Context) error {
	promRoute, err := t.factory.PrometheusK8sRoute()
	if err != nil {
		return errors.Wrap(err, "initializing Prometheus Route failed")
	}

	promURL, err := t.client.GetRouteURL(ctx, promRoute)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve Prometheus host")
	}

	amRoute, err := t.factory.AlertmanagerRoute()
	if err != nil {
		return errors.Wrap(err, "initializing Alertmanager Route failed")
	}

	amURL, err := t.client.GetRouteURL(ctx, amRoute)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve Alertmanager host")
	}

	var grafanaURL *url.URL

	if t.config.ClusterMonitoringConfiguration.GrafanaConfig.IsEnabled() {
		grafanaRoute, err := t.factory.GrafanaRoute()
		if err != nil {
			return errors.Wrap(err, "initializing Grafana Route failed")
		}

		grafanaURL, err = t.client.GetRouteURL(ctx, grafanaRoute)
		if err != nil {
			return errors.Wrap(err, "failed to retrieve Grafana host")
		}
	}

	thanosRoute, err := t.factory.ThanosQuerierRoute()
	if err != nil {
		return errors.Wrap(err, "initializing Thanos Querier Route failed")
	}

	thanosURL, err := t.client.GetRouteURL(ctx, thanosRoute)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve Thanos Querier host")
	}

	cm := t.factory.SharingConfig(promURL, amURL, grafanaURL, thanosURL)
	err = t.client.CreateOrUpdateConfigMap(ctx, cm)
	if err != nil {
		return errors.Wrapf(err, "reconciling %s/%s Config ConfigMap failed", cm.Namespace, cm.Name)
	}

	return nil
}
