package virtual_machines_test

import (
	"reflect"
	"testing"

	plugin "github.com/magneticstain/ip-2-cloudresource/azure/plugin/virtual_machines"
	generalResource "github.com/magneticstain/ip-2-cloudresource/resource"
)

func azvmPlugFactory() plugin.AzVirtualMachinePlugin {
	azvmPlug := plugin.AzVirtualMachinePlugin{}

	return azvmPlug
}

func TestGetResources(t *testing.T) {
	azvmPlug := azvmPlugFactory()

	vmResources, _ := azvmPlug.GetResources()

	expectedType := "Resource"
	for _, resource := range vmResources {
		resourceType := reflect.TypeOf(resource)
		if resourceType.Name() != expectedType {
			t.Errorf("Fetching resources via Azure Virtual Machines Plugin failed; wanted %s type, received %s", expectedType, resourceType.Name())
		}
	}
}

func TestSearchResources(t *testing.T) {
	azvmPlug := azvmPlugFactory()

	var tests = []struct {
		ipAddr, expectedType string
	}{
		{"1.1.1.1", "Resource"},
		{"1234.45.9666.1", "Resource"},
		{"18.161.22.61", "Resource"},
		{"2600:9000:24eb:dc00:1:3b80:4f00:21", "Resource"},
		{"x2600:9000:24eb:XYZ1:1:3b80:4f00:21", "Resource"},
	}

	var matchingResource generalResource.Resource
	for _, td := range tests {
		testName := td.ipAddr

		t.Run(testName, func(t *testing.T) {
			matchedInstance, _ := azvmPlug.SearchResources(td.ipAddr, &matchingResource)
			matchedInstanceType := reflect.TypeOf(*matchedInstance)

			if matchedInstanceType.Name() != td.expectedType {
				t.Errorf("Azure Virtual Machines search failed; expected %s after search, received %s", td.expectedType, matchedInstanceType.Name())
			}
		})
	}
}
