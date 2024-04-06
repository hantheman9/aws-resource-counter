package main

import (
	"github.com/aws/aws-sdk-go/service/redshift"

	color "github.com/logrusorgru/aurora"
)

// RedshiftClusterCounts retrieves the count of all Redshift clusters either for all
// regions (allRegions is true) or the region associated with the session.
func RedshiftClusterCounts(sf ServiceFactory, am ActivityMonitor, allRegions bool) int {
	am.StartAction("Retrieving Redshift cluster counts")
	clusterCount := 0

	if allRegions {
		regionsSlice := GetEC2Regions(sf.GetEC2InstanceService(""), am)
		for _, regionName := range regionsSlice {
			clusterCount += redshiftClusterCountForSingleRegion(sf.GetRedshiftService(regionName), am)
		}
	} else {
		clusterCount = redshiftClusterCountForSingleRegion(sf.GetRedshiftService(""), am)
	}

	am.EndAction("OK (%d)", color.Bold(clusterCount))
	return clusterCount
}

// Get the Redshift cluster count for a single region
func redshiftClusterCountForSingleRegion(rsService *RedshiftService, am ActivityMonitor) int {
	input := &redshift.DescribeClustersInput{}
	result, err := rsService.Client.DescribeClusters(input)
	if err != nil {
		am.CheckError(err)
		return 0
	}
	return len(result.Clusters)
}
