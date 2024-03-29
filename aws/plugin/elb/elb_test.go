package plugin_test

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	awsconnector "github.com/magneticstain/ip-2-cloudresource/aws/aws_connector"
	plugin "github.com/magneticstain/ip-2-cloudresource/aws/plugin/elb"
)

func elbpFactory() plugin.ELBPlugin {
	ac, _ := awsconnector.New()

	elbp := plugin.ELBPlugin{AwsConn: ac}

	return elbp
}

func TestGetElbListeners(t *testing.T) {
	elbp := elbpFactory()

	var tests = []struct {
		testName, elbArn, expectedType string
	}{
		{"validElbArn", "arn:aws:elasticloadbalancing:us-east-1:509915386432:loadbalancer/app/IP2CR-Testing-ALB/e482622e74065ea", "Listener"},
	}

	var elbListeners []types.Listener
	var elbListenersType string
	for _, td := range tests {
		testName := td.testName

		t.Run(testName, func(t *testing.T) {
			elbListeners, _ = elbp.GetElbListeners(td.elbArn)

			for _, listener := range elbListeners {
				elbListenersType = reflect.TypeOf(listener).Name()

				if elbListenersType != td.expectedType {
					t.Errorf("ELB listener fetch failed; expected %s after search, received %s", td.expectedType, elbListenersType)
				}
			}
		})
	}
}

func TestGetElbTargets(t *testing.T) {
	elbp := elbpFactory()

	var tests = []struct {
		testName, expectedType string
	}{
		{"validElbTgtArn", "ELBTarget"},
	}

	var elbListeners []types.Listener
	for _, td := range tests {
		testName := td.testName

		t.Run(testName, func(t *testing.T) {
			elbTargets, _ := elbp.GetElbTgts(elbListeners)

			for _, tgt := range elbTargets {
				elbTgtType := reflect.TypeOf(tgt).Name()

				if elbTgtType != td.expectedType {
					t.Errorf("ELB target fetch failed; expected %s after search, received %s", td.expectedType, elbTgtType)
				}
			}
		})
	}
}

func TestGetResources(t *testing.T) {
	elbp := elbpFactory()

	elbResources, _ := elbp.GetResources()

	expectedType := "LoadBalancer"
	for _, elb := range elbResources {
		elbType := reflect.TypeOf(elb)
		if elbType.Name() != expectedType {
			t.Errorf("Fetching resources via ELB Plugin failed; wanted %s type, received %s", expectedType, elbType.Name())
		}
	}
}

func TestSearchResources(t *testing.T) {
	elbp := elbpFactory()

	var tests = []struct {
		ipAddr, expectedType string
	}{
		{"1.1.1.1", "Resource"},
		{"1234.45.9666.1", "Resource"},
		{"18.161.22.61", "Resource"},
		{"2600:9000:24eb:dc00:1:3b80:4f00:21", "Resource"},
		{"x2600:9000:24eb:XYZ1:1:3b80:4f00:21", "Resource"},
	}

	for _, td := range tests {
		testName := td.ipAddr

		t.Run(testName, func(t *testing.T) {
			matchedELB, _ := elbp.SearchResources(td.ipAddr)
			matchedELBType := reflect.TypeOf(matchedELB)

			if matchedELBType.Name() != td.expectedType {
				t.Errorf("ELB search failed; expected %s after search, received %s", td.expectedType, matchedELBType.Name())
			}
		})
	}
}

func elbv1pFactory() plugin.ELBv1Plugin {
	ac, _ := awsconnector.New()

	elbv1p := plugin.ELBv1Plugin{AwsConn: ac}

	return elbv1p
}

func TestGetResources_Elbv1(t *testing.T) {
	elbv1p := elbv1pFactory()

	elbResources, _ := elbv1p.GetResources()

	expectedType := "LoadBalancerDescription"
	for _, elb := range elbResources {
		elbType := reflect.TypeOf(elb)
		if elbType.Name() != expectedType {
			t.Errorf("Fetching resources via ELBv1 Plugin failed; wanted %s type, received %s", expectedType, elbType.Name())
		}
	}
}

func TestSearchResources_Elbv1(t *testing.T) {
	elbv1p := elbv1pFactory()

	var tests = []struct {
		ipAddr, expectedType string
	}{
		{"1.1.1.1", "Resource"},
		{"1234.45.9666.1", "Resource"},
		{"18.161.22.61", "Resource"},
		{"2600:9000:24eb:dc00:1:3b80:4f00:21", "Resource"},
		{"x2600:9000:24eb:XYZ1:1:3b80:4f00:21", "Resource"},
	}

	for _, td := range tests {
		testName := td.ipAddr

		t.Run(testName, func(t *testing.T) {
			matchedELB, _ := elbv1p.SearchResources(td.ipAddr)
			matchedELBType := reflect.TypeOf(matchedELB)

			if matchedELBType.Name() != td.expectedType {
				t.Errorf("ELBv1 search failed; expected %s after search, received %s", td.expectedType, matchedELBType.Name())
			}
		})
	}
}
