package main

import (
    "github.com/aws/aws-sdk-go/service/docdb"

	color "github.com/logrusorgru/aurora"
)

// DocDBInstanceCounts retrieves the count of all DocumentDB instances either for all
// regions (allRegions is true) or the region associated with the session.
func DocDBInstanceCounts(sf ServiceFactory, am ActivityMonitor, allRegions bool) int {
    am.StartAction("Retrieving DocumentDB instance counts")
    instanceCount := 0

    if allRegions {
        regionsSlice := GetEC2Regions(sf.GetEC2InstanceService(""), am)
        for _, regionName := range regionsSlice {
            instanceCount += docDBInstanceCountForSingleRegion(sf.GetDocDBService(regionName), am)
        }
    } else {
        instanceCount = docDBInstanceCountForSingleRegion(sf.GetDocDBService(""), am)
    }

	am.EndAction("OK (%d)", color.Bold(instanceCount))
    return instanceCount
}

// Get the DocumentDB instance count for a single region
func docDBInstanceCountForSingleRegion(dbService *DocDBService, am ActivityMonitor) int {
    input := &docdb.DescribeDBInstancesInput{}
    result, err := dbService.Client.DescribeDBInstances(input)
    if err != nil {
        am.CheckError(err)
        return 0
    }
    return len(result.DBInstances)
}