package main

import (
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/service/cloudfront"

    color "github.com/logrusorgru/aurora"
)

func CloudFrontFunctionCounts(sf ServiceFactory, am ActivityMonitor) int {
    am.StartAction("Retrieving CloudFront Function counts")
    cfService := sf.GetCloudFrontService()

    // Initialize count
    totalFunctionCount := 0

    // Stages to check
    stages := []string{"LIVE", "DEVELOPMENT"}

    for _, stage := range stages {
        input := &cloudfront.ListFunctionsInput{
            Stage: aws.String(stage),
        }

        result, err := cfService.Client.ListFunctions(input)
        if err != nil {
            am.CheckError(err)
            return 0
        }

        count := len(result.FunctionList.Items)
        totalFunctionCount += count
    }

    am.EndAction("OK (%d)", color.Bold(totalFunctionCount))
    return totalFunctionCount
}