// Copyright 2019 The Terraformer Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aws

import (
	"context"
	"fmt"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/aws/aws-sdk-go-v2/service/eks"
)

var eksAllowEmptyValues = []string{"tags."}

type EksGenerator struct {
	AWSService
}

func (g *EksGenerator) InitResources() error {
	context := context.TODO()
	config, err := g.generateConfig()
	if err != nil {
		return err
	}
	svc := eks.NewFromConfig(config)

	clusters, err := g.loadClusters(context, svc)
	if err != nil {
		return err
	}
	for _, cluster := range clusters {
		if err := g.loadNodeGroups(context, svc, cluster); err != nil {
			return err
		}
	}
	return nil
}

func (g *EksGenerator) loadClusters(context context.Context, svc *eks.Client) ([]string, error) {
	var clusters []string
	p := eks.NewListClustersPaginator(svc, &eks.ListClustersInput{})
	for p.HasMorePages() {
		page, e := p.NextPage(context)
		if e != nil {
			return nil, e
		}
		for _, clusterName := range page.Clusters {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				clusterName,
				clusterName,
				"aws_eks_cluster",
				"aws",
				eksAllowEmptyValues,
			))
			clusters = append(clusters, clusterName)
		}
	}
	return clusters, nil
}

func (g *EksGenerator) loadNodeGroups(context context.Context, svc *eks.Client, clusterName string) error {
	p := eks.NewListNodegroupsPaginator(svc, &eks.ListNodegroupsInput{
		ClusterName: &clusterName,
	})
	for p.HasMorePages() {
		page, err := p.NextPage(context)
		if err != nil {
			return err
		}
		for _, nodeGroup := range page.Nodegroups {
			g.Resources = append(g.Resources, terraformutils.NewSimpleResource(
				fmt.Sprintf("%s:%s", clusterName, nodeGroup),
				nodeGroup,
				"aws_eks_node_group",
				"aws",
				eksAllowEmptyValues,
			))
		}
	}
	return nil
}
