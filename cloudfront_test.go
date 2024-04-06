package main

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/aws/aws-sdk-go/service/cloudfront/cloudfrontiface"
)

// Mock data for CloudFront Functions
var cloudFrontFunctions = &cloudfront.ListFunctionsOutput{
	FunctionList: &cloudfront.FunctionList{
		Items: []*cloudfront.FunctionSummary{
			{
				Name: aws.String("FunctionOne"),
				Status: aws.String("LIVE"),
			},
			{
				Name: aws.String("FunctionTwo"),
				Status: aws.String("DEVELOPMENT"),
			},
		},
	},
}

type fakeCloudFrontService struct {
	cloudfrontiface.CloudFrontAPI
}

func (f *fakeCloudFrontService) ListFunctions(input *cloudfront.ListFunctionsInput) (*cloudfront.ListFunctionsOutput, error) {
	if input.Stage == nil {
		return nil, errors.New("Stage is required")
	}
	return cloudFrontFunctions, nil
}

func TestCloudFrontFunctionCounts(t *testing.T) {
	cases := []struct {
		Stage          string
		ExpectedCount  int
		ExpectError    bool
	}{
		{
			Stage:         "LIVE",
			ExpectedCount: 1,
		},
		{
			Stage:         "DEVELOPMENT",
			ExpectedCount: 1,
		},
		{
			Stage:         "",
			ExpectError:   true,
		},
	}

	for _, c := range cases {
		sf := ServiceFactory{
			CloudFront: &fakeCloudFrontService{},
		}

		// Mock activity monitor (not shown, assume similar to EC2 test)
		mon := &mock.ActivityMonitorImpl{}

		// Adjust CloudFrontFunctionCounts to accept a stage parameter for testing
		actualCount, err := CloudFrontFunctionCounts(sf, mon, c.Stage)
		if c.ExpectError {
			if err == nil {
				t.Error("Expected an error but did not get one")
			}
		} else {
			if err != nil {
				t.Errorf("Did not expect an error but got: %v", err)
			}
			if actualCount != c.ExpectedCount {
				t.Errorf("Expected %d CloudFront Functions, got %d", c.ExpectedCount, actualCount)
			}
		}
	}
}