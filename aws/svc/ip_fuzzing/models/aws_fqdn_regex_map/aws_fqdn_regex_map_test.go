package awsfqdnregexmap_test

import (
	"testing"

	awsfqdnregexmap "github.com/magneticstain/ip-2-cloudresource/aws/svc/ip_fuzzing/models/aws_fqdn_regex_map"
)

func TestGetRegexMap(t *testing.T) {
	var tests = []struct {
		cloudSvc string
	}{
		{"CLOUDFRONT"},
		{"EC2"},
	}

	for _, td := range tests {
		testName := td.cloudSvc

		t.Run(testName, func(t *testing.T) {
			svcNameRegex := awsfqdnregexmap.GetRegexMap()

			svcFound := false
			for svcName := range svcNameRegex {
				if svcName == td.cloudSvc {
					svcFound = true
					break
				}
			}

			if !svcFound {
				t.Errorf("Did not find expected service name - [ %s ] - in regex map", td.cloudSvc)
			}
		})
	}
}
