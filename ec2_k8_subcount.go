/******************************************************************************
Cloud Resource Counter
File: spot.go

Summary: Provides a count of all Spot EC2 instances.
******************************************************************************/

package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	color "github.com/logrusorgru/aurora"
)

// SpotInstances retrieves the count of all EC2 spot instances
// either for all regions (allRegions is true) or the region
// associated with the session.
// This method gives status back to the user via the supplied
// ActivityMonitor instance.
func EC2K8SubInstances(sf ServiceFactory, am ActivityMonitor, allRegions bool) int {
	// Indicate activity
	am.StartAction("Retrieving EC2 K8 related VMs Sub-instance counts")

	// Should we get the counts for all regions?
	instanceCount := 0
	if allRegions {
		// Get the list of all enabled regions for this account
		regionsSlice := GetEC2Regions(sf.GetEC2InstanceService(""), am)

		// Loop through all of the regions
		for _, regionName := range regionsSlice {
			// Get the EC2 counts for a specific region
			instanceCount += ec2K8SubInstancesForSingleRegion(sf.GetEC2InstanceService(regionName), am)
		}
	} else {
		// Get the EC2 counts for the region selected by this session
		instanceCount = ec2K8SubInstancesForSingleRegion(sf.GetEC2InstanceService(""), am)
	}

	// Indicate end of activity
	am.EndAction("OK (%d)", color.Bold(instanceCount))

	return instanceCount
}

func ec2K8SubInstancesForSingleRegion(ec2is *EC2InstanceService, am ActivityMonitor) int {
	// Indicate activity
	am.Message(".")

	// Construct our input to find ONLY RUNNING EC2 instances that also have the "aws:eks:cluster-name" tag
	input := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag-key"),
				Values: []*string{
					aws.String("aws:eks:cluster-name"),
				},
			},
			{
				Name: aws.String("instance-state-name"),
				Values: []*string{
					aws.String("running"),
				},
			},
		},
	}

	// Invoke our service
	instanceCount := 0
	err := ec2is.InspectInstances(input, func(dio *ec2.DescribeInstancesOutput, lastPage bool) bool {
		// Loop through each reservation
		for _, reservation := range dio.Reservations {
			// We assume that the AWS Service has properly filtered the list of returned instances
			instanceCount += len(reservation.Instances)
		}

		return true
	})

	// Check for error
	am.CheckError(err)

	return instanceCount
}
