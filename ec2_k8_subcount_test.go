/******************************************************************************
Cloud Resource Counter
File: ec2_k8_subcount_test.go

Summary: The Unit Test for EC2 K8 sub-count.
******************************************************************************/

package main

import (
	"testing"

	"github.com/expel-io/aws-resource-counter/mock"
)

// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// Unit Test for EC2 K8 sub-count.
// =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

func TestEC2K8SubInstances(t *testing.T) {
	// Describe all of our test cases: 1 failure and 4 success cases
	cases := []struct {
		RegionName    string
		AllRegions    bool
		ExpectedCount int
		ExpectError   bool
	}{
		{
			RegionName:    "us-east-1",
			ExpectedCount: 1,
		}, {
			RegionName:    "us-east-2",
			ExpectedCount: 0,
		}, {
			RegionName:    "af-south-1",
			ExpectedCount: 0,
		}, {
			RegionName:  "undefined-region",
			ExpectError: true,
		}, {
			AllRegions:    true,
			ExpectedCount: 1,
		},
	}

	// Loop through each test case
	for _, c := range cases {
		// Create our fake service factory
		sf := fakeEC2ServiceFactory{
			RegionName: c.RegionName,
			DRResponse: ec2Regions,
		}

		// Create a mock activity monitor
		mon := &mock.ActivityMonitorImpl{}

		// Invoke our EC K8 Subcount Instances function
		actualCount := EC2K8SubInstances(sf, mon, c.AllRegions)

		// Did we expect an error?
		if c.ExpectError {
			// Did it fail to arrive?
			if !mon.ErrorOccured {
				t.Error("Expected an error to occur, but it did not... :^(")
			}
		} else if mon.ErrorOccured {
			t.Errorf("Unexpected error occurred: %s", mon.ErrorMessage)
		} else if actualCount != c.ExpectedCount {
			t.Errorf("Error: EC K8 SubcountInstances returned %d; expected %d", actualCount, c.ExpectedCount)
		} else if mon.ProgramExited {
			t.Errorf("Unexpected Exit: The program unexpected exited with status code=%d", mon.ExitCode)
		}
	}
}
