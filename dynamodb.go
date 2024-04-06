package main

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"

	color "github.com/logrusorgru/aurora"
)

// CountDynamoDBTables retrieves the count of all DynamoDB tables either for all
// regions (allRegions is true) or the region associated with the session.
func CountDynamoDBTables(sf ServiceFactory, am ActivityMonitor, allRegions bool) int {
	am.StartAction("Retrieving DynamoDB table counts")
	tableCount := 0

	processRegion := func(regionName string) {
		svc := sf.GetDynamoDBService(regionName).Client
		input := &dynamodb.ListTablesInput{}

		for {
			output, err := svc.ListTables(input)
			if err != nil {
				am.CheckError(err)
				break
			}

			tableCount += len(output.TableNames)

			if output.LastEvaluatedTableName == nil {
				break
			}

			input.ExclusiveStartTableName = output.LastEvaluatedTableName
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

	am.EndAction("OK (%d)", color.Bold(tableCount))
	return tableCount
}
