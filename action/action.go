package action

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/eks"
)

type AwsJanitorAction interface {
	Cleanup(ctx context.Context, input *Input) error
}

func New(commit bool) AwsJanitorAction {
	return &action{
		commit: commit,
	}
}

type action struct {
	commit bool
}

func (a *action) Cleanup(ctx context.Context, input *Input) error {

	//NOTE: ordering matters here!
	cleanupFuncs := map[string]CleanupFunc{
		eks.ServiceName:         a.cleanEKSClusters,
		autoscaling.ServiceName: a.cleanASGs,
	}
	inputRegions := strings.Split(input.Regions, ",")

	for service, cleanupFunc := range cleanupFuncs {
		regions := getServiceRegions(service, inputRegions)

		for _, region := range regions {
			sess, err := session.NewSession(&aws.Config{
				Region: aws.String(region)},
			)
			if err != nil {
				return fmt.Errorf("failed to create aws session for region %s: %w", region, err)
			}

			scope := &CleanupScope{
				TTL:     input.TTL,
				Session: sess,
				Commit:  input.Commit,
			}

			Log("Cleaning up resources for service %s in region %s", service, region)
			if err := cleanupFunc(ctx, scope); err != nil {
				return fmt.Errorf("failed running cleanup for service %s: %w", service, err)
			}
		}
	}

	return nil
}

func getServiceRegions(service string, inputRegions []string) []string {
	regions := []string{}
	allRegions := inputRegions[0] == "*"

	sr, exists := endpoints.RegionsForService(endpoints.DefaultPartitions(), endpoints.AwsPartitionID, service)
	if exists {
		for _, region := range sr {
			if allRegions {
				regions = append(regions, region.ID())
			} else {
				for _, r := range inputRegions {
					if r == region.ID() {
						regions = append(regions, region.ID())
					}
				}
			}
		}
	}

	return regions
}
