package main

import (
	"github.com/aws/aws-sdk-go/service/apigatewayv2"

	color "github.com/logrusorgru/aurora"
)

func countAPIGatewayV2Apis(sf ServiceFactory, am ActivityMonitor, allRegions bool) int {
	am.StartAction("Retrieving API Gateway V2 API counts")
	apiCount := 0

	processRegion := func(regionName string) {
		svc := sf.GetAPIGatewayV2Service(regionName).Client
		input := &apigatewayv2.GetApisInput{}

		// Manual pagination
		for {
			output, err := svc.GetApis(input)
			if err != nil {
				am.CheckError(err)
				break
			}

			apiCount += len(output.Items)

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

	am.EndAction("OK (%d)", color.Bold(apiCount))
	return apiCount
}
