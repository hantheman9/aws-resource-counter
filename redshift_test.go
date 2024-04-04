package main

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/aws/aws-sdk-go/service/redshift/redshiftiface"
	"github.com/expel-io/aws-resource-counter/mock"
)

// Define a mock struct to implement the RedshiftAPI interface
type mockRedshiftClient struct {
	redshiftiface.RedshiftAPI
	Resp redshift.DescribeClustersOutput
}

// Mock DescribeClusters function
func (m *mockRedshiftClient) DescribeClusters(input *redshift.DescribeClustersInput) (*redshift.DescribeClustersOutput, error) {
	return &m.Resp, nil
}

// Mock data for Redshift clusters
var mockRedshiftData = map[string]*redshift.DescribeClustersOutput{
	"us-east-1": {
		Clusters: []*redshift.Cluster{
			{ClusterIdentifier: aws.String("cluster-1")},
			{ClusterIdentifier: aws.String("cluster-2")},
		},
	},
	"us-west-2": {
		Clusters: []*redshift.Cluster{
			{ClusterIdentifier: aws.String("cluster-3")},
		},
	},
}

func TestRedshiftClusterCounts(t *testing.T) {
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
		sf := mock.ServiceFactory{
			MockRedshiftClient: func(regionName string) *redshift.Redshift {
				return &redshift.Redshift{Client: &mockRedshiftClient{Resp: *mockRedshiftData[regionName]}}
			},
		}
		am := &mock.ActivityMonitorImpl{}

		actualCount := RedshiftClusterCounts(sf, am, c.AllRegions)

		if c.ExpectError {
			if !am.ErrorOccured {
				t.Error("Expected an error to occur, but it did not")
			}
		} else if am.ErrorOccured {
			t.Errorf("Unexpected error occurred: %s", am.ErrorMessage)
		} else if actualCount != c.ExpectedCount {
			t.Errorf("Error: RedshiftClusterCounts returned %d; expected %d", actualCount, c.ExpectedCount)
		}
	}
}