package main

import (
	"github.com/aws/aws-sdk-go/service/elbv2"

	color "github.com/logrusorgru/aurora"
)

// CountELBv2s retrieves the count of all Elastic Load Balancers V2 either for all
// regions (allRegions is true) or the region associated with the session.
func CountELBv2s(sf ServiceFactory, am ActivityMonitor, allRegions bool) int {
	am.StartAction("Retrieving ELBv2 counts")
	elbv2Count := 0

	processRegion := func(regionName string) {
		svc := sf.GetELBv2Service(regionName).Client
		input := &elbv2.DescribeLoadBalancersInput{}

		for {
			output, err := svc.DescribeLoadBalancers(input)
			if err != nil {
				am.CheckError(err)
				break
			}

			elbv2Count += len(output.LoadBalancers)

			if output.NextMarker == nil {
				break
			}

			input.Marker = output.NextMarker
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

	am.EndAction("OK (%d)", color.Bold(elbv2Count))
	return elbv2Count
}
