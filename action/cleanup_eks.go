package action

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/eks"
)

func (a *action) cleanEKSClusters(ctx context.Context, input *CleanupScope) error {
	client := eks.New(input.Session)

	clustersToDelete := []*string{}
	pageFunc := func(page *eks.ListClustersOutput, _ bool) bool {
		for _, name := range page.Clusters {
			cluster, err := client.DescribeClusterWithContext(ctx, &eks.DescribeClusterInput{
				Name: name,
			})
			if err != nil {
				LogWarning("failed getting cluster %s: %s", *name, err.Error())
				continue
			}

			if _, ok := cluster.Cluster.Tags[input.IgnoreTag]; ok {
				LogDebug("eks cluster %s has ignore tag, skipping cleanup", *name)
				continue
			}

			maxAge := cluster.Cluster.CreatedAt.Add(input.TTL)

			if time.Now().Before(maxAge) {
				LogDebug("eks cluster %s has max age greater than now, skipping cleanup", *name)
				continue
			}

			LogDebug("adding eks cluster %s to delete list", *name)
			clustersToDelete = append(clustersToDelete, name)
		}

		return true
	}

	if err := client.ListClustersPagesWithContext(ctx, &eks.ListClustersInput{}, pageFunc); err != nil {
		return fmt.Errorf("failed getting list of eks clusters: %w", err)
	}

	if len(clustersToDelete) == 0 {
		Log("no eks clusters to delete")
		return nil
	}

	for _, clusterName := range clustersToDelete {
		if !a.commit {
			LogDebug("skipping deletion of eks cluster %s as running in dry-mode", *clusterName)
			continue
		}

		if err := a.deleteEKSCluster(ctx, *clusterName, client); err != nil {
			LogError("failed to delete cluster %s: %s", *clusterName, err.Error())
		}
	}

	return nil
}

func (a *action) deleteEKSCluster(ctx context.Context, clusterName string, client *eks.EKS) error {
	Log("Deleting EKS cluster %s", clusterName)

	LogDebug("Deleting nodegroups for cluster %s", clusterName)

	listErr := client.ListNodegroupsPagesWithContext(ctx, &eks.ListNodegroupsInput{ClusterName: &clusterName}, func(page *eks.ListNodegroupsOutput, b bool) bool {
		for _, ngName := range page.Nodegroups {
			Log("Deleting nodegroup %s in cluster %s", *ngName, clusterName)
			if _, err := client.DeleteNodegroupWithContext(ctx, &eks.DeleteNodegroupInput{ClusterName: &clusterName, NodegroupName: ngName}); err != nil {
				LogError("failed to deleted nodegroups %s for cluster %s: %s", *ngName, clusterName, err.Error())
			}

			if err := client.WaitUntilNodegroupDeletedWithContext(ctx, &eks.DescribeNodegroupInput{ClusterName: &clusterName, NodegroupName: ngName}); err != nil {
				LogError("failed to wait for nodegroups %s in cluster %s to be deleted: %s", *ngName, clusterName, err.Error())
			}
		}

		return true
	})
	if listErr != nil {
		return fmt.Errorf("failed to list nodegroups for cluster %s: %w", clusterName, listErr)
	}

	if _, err := client.DeleteClusterWithContext(ctx, &eks.DeleteClusterInput{Name: &clusterName}); err != nil {
		return fmt.Errorf("failed to delete cluster %s: %w", clusterName, err)
	}

	if err := client.WaitUntilClusterDeletedWithContext(ctx, &eks.DescribeClusterInput{Name: &clusterName}); err != nil {
		return fmt.Errorf("failed to wait for cluster %s to be delete: %w", clusterName, err)
	}

	return nil
}
