package main

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/aws/aws-sdk-go/service/docdb/docdbiface"
)

// Define a mock struct to implement the DocDBAPI interface
type mockDocDBClient struct {
	docdbiface.DocDBAPI
	Resp docdb.DescribeDBInstancesOutput
}

// Mock DescribeDBInstances function
func (m *mockDocDBClient) DescribeDBInstances(input *docdb.DescribeDBInstancesInput) (*docdb.DescribeDBInstancesOutput, error) {
	return &m.Resp, nil
}

// Mock data for DocDB instances
var mockDocDBData = map[string]*docdb.DescribeDBInstancesOutput{
	"us-east-1": {
		DBInstances: []*docdb.DBInstance{
			{DBInstanceIdentifier: aws.String("db-instance-1")},
			{DBInstanceIdentifier: aws.String("db-instance-2")},
		},
	},
	"us-west-2": {
		DBInstances: []*docdb.DBInstance{
			{DBInstanceIdentifier: aws.String("db-instance-3")},
		},
	},
}

func TestDocDBInstanceCounts(t *testing.T) {
	cases := []struct {
		RegionName    string
		AllRegions    bool
		ExpectedCount int
		ExpectError   bool
	}{
		{RegionName: "us-east-1", ExpectedCount: 2},
		{RegionName: "us-west-2", ExpectedCount: 1},
		{AllRegions: true, ExpectedCount: 3},
	}

	for _, c := range cases {
		sf := mockServiceFactory{
			MockDocDBClient: func(regionName string) docdbiface.DocDBAPI {
				return &mockDocDBClient{Resp: *mockDocDBData[regionName]}
			},
		}
		am := &mockActivityMonitor{}

		actualCount := DocDBInstanceCounts(&sf, am, c.AllRegions)

		if c.ExpectError {
			if !am.ErrorOccurred {
				t.Error("Expected an error to occur, but it did not")
			}
		} else if am.ErrorOccurred {
			t.Errorf("Unexpected error occurred: %s", am.ErrorMessage)
		} else if actualCount != c.ExpectedCount {
			t.Errorf("Error: DocDBInstanceCounts returned %d; expected %d", actualCount, c.ExpectedCount)
		}
	}
}