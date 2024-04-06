package main

import (
	"github.com/aws/aws-sdk-go/service/elb"

	color "github.com/logrusorgru/aurora"
)

// CountELBs retrieves the count of all Elastic Load Balancers either for all
// regions (allRegions is true) or the region associated with the session.
func CountELBs(sf ServiceFactory, am ActivityMonitor, allRegions bool) int {
	am.StartAction("Retrieving ELB counts")
	elbCount := 0

	processRegion := func(regionName string) {
		svc := sf.GetELBService(regionName).Client
		input := &elb.DescribeLoadBalancersInput{}

		for {
			output, err := svc.DescribeLoadBalancers(input)
			if err != nil {
				am.CheckError(err)
				break
			}

			elbCount += len(output.LoadBalancerDescriptions)

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

	am.EndAction("OK (%d)", color.Bold(elbCount))
	return elbCount
}
