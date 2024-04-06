package main

import (
	"github.com/aws/aws-sdk-go/service/elasticache"
	color "github.com/logrusorgru/aurora"
)

// ElastiCacheClusterCounts retrieves the count of all ElastiCache CacheClusters either for all
// regions (allRegions is true) or the region associated with the session.
func ElastiCacheClusterCounts(sf ServiceFactory, am ActivityMonitor, allRegions bool) int {
	am.StartAction("Retrieving ElastiCache CacheCluster counts")
	clusterCount := 0

	if allRegions {
		regionsSlice := GetEC2Regions(sf.GetEC2InstanceService(""), am)
		for _, regionName := range regionsSlice {
			clusterCount += elasticacheClusterCountForSingleRegion(sf.GetElastiCacheService(regionName), am)
		}
	} else {
		clusterCount = elasticacheClusterCountForSingleRegion(sf.GetElastiCacheService(""), am)
	}

	am.EndAction("OK (%d)", color.Bold(clusterCount))
	return clusterCount
}

// Get the ElastiCache CacheCluster count for a single region
func elasticacheClusterCountForSingleRegion(ecService *ElastiCacheService, am ActivityMonitor) int {
	input := &elasticache.DescribeCacheClustersInput{}
	result, err := ecService.Client.DescribeCacheClusters(input)
	if err != nil {
		am.CheckError(err)
		return 0
	}
	return len(result.CacheClusters)
}
