package action

import (
	"context"
	"fmt"
	"time"

	cf "github.com/aws/aws-sdk-go/service/cloudformation"
)

func (a *action) cleanCfStacks(ctx context.Context, input *CleanupScope) error {
	client := cf.New(input.Session)

	stacksToDelete := []*string{}
	pageFunc := func(page *cf.DescribeStacksOutput, _ bool) bool {
		for _, stack := range page.Stacks {
			var ignore bool
			for _, tag := range stack.Tags {
				if *tag.Key == input.IgnoreTag {
					ignore = true
					break
				}
			}

			if ignore {
				LogDebug("cloudformation stack %s has ignore tag, skipping cleanup", *stack.StackName)
				continue
			}

			maxAge := stack.CreationTime.Add(input.TTL)

			if time.Now().Before(maxAge) {
				LogDebug("cloudformation stack %s has max age greater than now, skipping cleanup", *stack.StackName)
				continue
			}

			LogDebug("adding cloudformation stack %s to delete list", *stack.StackName)
			stacksToDelete = append(stacksToDelete, stack.StackName)
		}

		return true
	}

	if err := client.DescribeStacksPagesWithContext(ctx, &cf.DescribeStacksInput{}, pageFunc); err != nil {
		return fmt.Errorf("failed getting list of cloudformation stacks: %w", err)
	}

	if len(stacksToDelete) == 0 {
		Log("no eks clusters to delete")
		return nil
	}

	for _, stackName := range stacksToDelete {
		if !a.commit {
			LogDebug("skipping deletion of cloudformation stack %s as running in dry-mode", *stackName)
			continue
		}

		//if err := a.deleteCfStack(ctx, *stackName, client); err != nil {
		//	LogError("failed to delete cloudformation stack %s: %s", *stackName, err.Error())
		//}
	}

	return nil
}

func (a *action) deleteCfStack(ctx context.Context, stackName string, client *cf.CloudFormation) error {
	return nil
}
