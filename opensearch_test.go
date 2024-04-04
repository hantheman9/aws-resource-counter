func TestOpenSearchDomainCounts(t *testing.T) {
	cases := []struct {
		RegionName    string
		AllRegions    bool
		ExpectedCount int
		ExpectError   bool
	}{
		{
			RegionName:    "us-east-1",
			ExpectedCount: 2,
		},
		{
			RegionName:    "us-east-2",
			ExpectedCount: 1,
		},
		{
			RegionName:    "af-south-1",
			ExpectedCount: 0,
		},
		{
			AllRegions:    true,
			ExpectedCount: 3,
		},
	}

	for _, c := range cases {
		sf := fakeEC2ServiceFactory{ // Reusing the EC2 service factory for simplicity
			RegionName: c.RegionName,
		}
		mon := &mock.ActivityMonitorImpl{}

		// Mock the OpenSearch service within the factory
		sf.GetOpenSearchService = func(regionName string) *OpenSearchService {
			return &OpenSearchService{
				Client: &fakeOpenSearchService{
					ListDomainNamesOutput: openSearchDomainsPerRegion[regionName],
				},
			}
		}

		actualCount := OpenSearchDomainCounts(sf, mon, c.AllRegions)

		if c.ExpectError {
			if !mon.ErrorOccured {
				t.Error("Expected an error to occur, but it did not")
			}
		} else if mon.ErrorOccured {
			t.Errorf("Unexpected error occurred: %s", mon.ErrorMessage)
		} else if actualCount != c.ExpectedCount {
			t.Errorf("Error: OpenSearchDomainCounts returned %d; expected %d", actualCount, c.ExpectedCount)
		}
	}
}