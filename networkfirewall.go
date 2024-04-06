package main

import (
	"github.com/aws/aws-sdk-go/service/networkfirewall"

	color "github.com/logrusorgru/aurora"
)

// CountNetworkFirewalls retrieves the count of AWS Network Firewall Firewalls
// either for a specific region or across all regions.
func CountNetworkFirewalls(sf ServiceFactory, am ActivityMonitor, allRegions bool) int {
	am.StartAction("Retrieving AWS Network Firewall counts")
	firewallCount := 0

	processRegion := func(regionName string) {
		svc := sf.GetNetworkFirewallService(regionName).Client
		input := &networkfirewall.ListFirewallsInput{}

		for {
			output, err := svc.ListFirewalls(input)
			if err != nil {
				am.CheckError(err)
				break
			}

			firewallCount += len(output.Firewalls)

			if output.NextToken == nil {
				break
			}

			input.NextToken = output.NextToken
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

	am.EndAction("OK (%d)", color.Bold(firewallCount))
	return firewallCount
}
