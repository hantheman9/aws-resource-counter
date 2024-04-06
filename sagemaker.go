package main

import (
	"github.com/aws/aws-sdk-go/service/sagemaker"

	color "github.com/logrusorgru/aurora"
)

// CountSageMakerNotebookInstances retrieves the count of AWS SageMaker Notebook Instances
// either for a specific region or across all regions.
func CountSageMakerNotebookInstances(sf ServiceFactory, am ActivityMonitor, allRegions bool) int {
	am.StartAction("Retrieving AWS SageMaker Notebook Instance counts")
	notebookInstanceCount := 0

	processRegion := func(regionName string) {
		svc := sf.GetSageMakerService(regionName).Client
		input := &sagemaker.ListNotebookInstancesInput{}

		for {
			output, err := svc.ListNotebookInstances(input)
			if err != nil {
				am.CheckError(err)
				break
			}

			notebookInstanceCount += len(output.NotebookInstances)

			if output.NextToken == nil {
				break
			}

			input.NextToken = output.NextToken
		}
	}

	if allRegions {
		regionsSlice := GetEC2Regions(sf.GetEC2InstanceService(""), am)
		for _, regionName := range regionsSlice {
			processRegion(regionName)
		}
	} else {
		processRegion(sf.GetCurrentRegion())
	}

	am.EndAction("OK (%d)", color.Bold(notebookInstanceCount))
	return notebookInstanceCount
}
