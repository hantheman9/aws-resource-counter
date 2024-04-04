package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	color "github.com/logrusorgru/aurora"
)

// RestAPICounts retrieves the count of all API Gateway RestAPIs either for all
// regions (allRegions is true) or the region associated with the session.
// This method gives status back to the user via the supplied ActivityMonitor instance.
func RestAPICounts(sf ServiceFactory, am ActivityMonitor, allRegions bool) int {
	am.StartAction("Retrieving API Gateway RestAPI counts")

	totalAPICount := 0

	if allRegions {
		regionsSlice := GetEC2Regions(sf.GetEC2InstanceService(""), am)

		for _, regionName := range regionsSlice {
			totalAPICount += apiGatewayCountForSingleRegion(sf.GetAPIGatewayService(regionName), am)
		}
	} else {
		totalAPICount = apiGatewayCountForSingleRegion(sf.GetAPIGatewayService(""), am)
	}

	am.EndAction("OK (%d)", color.Bold(totalAPICount))
	return totalAPICount
}

// apiGatewayCountForSingleRegion gets the API Gateway RestAPI count for a single region
func apiGatewayCountForSingleRegion(apiService *APIGatewayService, am ActivityMonitor) int {
	input := &apigateway.GetRestApisInput{
		Limit: aws.Int64(500), // Adjust the limit as necessary
	}

	apiCount := 0
	err := apiService.Client.GetRestApisPages(input,
		func(page *apigateway.GetRestApisOutput, lastPage bool) bool {
			apiCount += len(page.Items)
			return !lastPage
		})

	if err != nil {
		am.CheckError(err)
		return 0
	}

	return apiCount
}
