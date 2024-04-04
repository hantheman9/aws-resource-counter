package main

import (
    "github.com/aws/aws-sdk-go/service/opensearchservice"

	color "github.com/logrusorgru/aurora"
)

// OpenSearchDomainCounts retrieves the count of all OpenSearch domains either for all
// regions (allRegions is true) or the region associated with the session.
func OpenSearchDomainCounts(sf ServiceFactory, am ActivityMonitor, allRegions bool) int {
    am.StartAction("Retrieving OpenSearch domain counts")
    domainCount := 0

    if allRegions {
        regionsSlice := GetEC2Regions(sf.GetEC2InstanceService(""), am)
        for _, regionName := range regionsSlice {
            domainCount += openSearchDomainCountForSingleRegion(sf.GetOpenSearchService(regionName), am)
        }
    } else {
        domainCount = openSearchDomainCountForSingleRegion(sf.GetOpenSearchService(""), am)
    }

	am.EndAction("OK (%d)", color.Bold(domainCount))
    return domainCount
}

// Get the OpenSearch domain count for a single region
func openSearchDomainCountForSingleRegion(osService *OpenSearchService, am ActivityMonitor) int {
    input := &opensearchservice.ListDomainNamesInput{}
    result, err := osService.Client.ListDomainNames(input)
    if err != nil {
        am.CheckError(err)
        return 0
    }
    return len(result.DomainNames)
}