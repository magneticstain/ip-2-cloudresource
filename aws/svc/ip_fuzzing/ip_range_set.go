package ipfuzzing

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"

	awsipprefix "github.com/magneticstain/ip-2-cloudresource/aws/svc/ip_fuzzing/models/aws_ip_prefix"
)

const awsIPRangeURL string = "https://ip-ranges.amazonaws.com/ip-ranges.json"

func FetchIPRanges() (awsipprefix.RawAwsIPRangeJSON, error) {
	var ipRangeData awsipprefix.RawAwsIPRangeJSON

	// fetch IP prefixes from AWS's Public IP Range API
	resp, err := http.Get(awsIPRangeURL)
	if err != nil {
		return ipRangeData, err
	} else if resp.StatusCode != http.StatusOK {
		return ipRangeData, fmt.Errorf("received HTTP status %s when fetching IP ranges from remote URL :: [ URL: %s ]", resp.Status, awsIPRangeURL)
	}
	defer resp.Body.Close() //nolint:errcheck

	// I know this isn't the most efficient way to do this, but for some reason, I could not get json.Decoder() working here
	jsonData, err := io.ReadAll(resp.Body)
	if err != nil {
		return ipRangeData, err
	}

	jsonErr := json.Unmarshal(jsonData, &ipRangeData)
	if jsonErr != nil {
		return ipRangeData, jsonErr
	}

	return ipRangeData, nil
}

func ConvertIPPrefixesToGeneric(ipv4Prefixes []awsipprefix.AwsIpv4Prefix, ipv6Prefixes []awsipprefix.AwsIpv6Prefix) ([]awsipprefix.GenericAWSPrefix, error) {
	// convert IPv4 (AwsIpv4Prefix) or IPv6 prefix (AwsIpv6Prefix) objects to GenericAWSPrefix
	var ipPrefixes []awsipprefix.GenericAWSPrefix

	if ipv4Prefixes != nil {
		for _, prefix := range ipv4Prefixes {
			ipPrefixes = append(ipPrefixes, awsipprefix.GenericAWSPrefix{
				IPRange:            prefix.IPPrefix,
				Region:             prefix.Region,
				Service:            prefix.Service,
				NetworkBorderGroup: prefix.NetworkBorderGroup,
			})
		}
	} else if ipv6Prefixes != nil {
		for _, prefix := range ipv6Prefixes {
			ipPrefixes = append(ipPrefixes, awsipprefix.GenericAWSPrefix{
				IPRange:            prefix.IPv6Prefix,
				Region:             prefix.Region,
				Service:            prefix.Service,
				NetworkBorderGroup: prefix.NetworkBorderGroup,
			})
		}
	} else {
		return ipPrefixes, fmt.Errorf("no IP prefixes defined; must send either IPv4 or IPv6 prefix set")
	}

	return ipPrefixes, nil
}

func ResolveIPAddrToCloudSvc(ipAddr string, ipPrefixSet []awsipprefix.GenericAWSPrefix) (string, error) {
	var cloudSvc string
	parsedIPAddr := net.ParseIP(ipAddr)

	for _, ipPrefix := range ipPrefixSet {
		_, cidrNet, err := net.ParseCIDR(ipPrefix.IPRange)
		if err != nil {
			return cloudSvc, err
		}

		if cidrNet.Contains(parsedIPAddr) {
			// target IP is within this IP range
			cloudSvc = ipPrefix.Service
			break
		}
	}

	return cloudSvc, nil
}
