package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rollbar/rollbar-go"
	log "github.com/sirupsen/logrus"

	awsconnector "github.com/magneticstain/ip-2-cloudresource/aws_connector"
	"github.com/magneticstain/ip-2-cloudresource/resource"
	"github.com/magneticstain/ip-2-cloudresource/search"
	"github.com/magneticstain/ip-2-cloudresource/utils"
)

const APP_VER = "v1.2.0"

func outputResults(matchedResource resource.Resource, networkMapping bool, silent bool, jsonOutput bool) {
	acctAliasFmted := strings.Join(matchedResource.AccountAliases, ", ")

	if !silent {
		if matchedResource.RID != "" {
			var acctStr string
			if matchedResource.AccountID == "current" {
				acctStr = "current account"
			} else {
				acctStr = fmt.Sprintf("account [ %s ( %s ) ]", matchedResource.AccountID, acctAliasFmted)
			}

			log.Info("resource found -> [ ", matchedResource.RID, " ] in ", acctStr)

			if networkMapping {
				var networkMapGraph string

				var networkResourceElmnt string
				networkMapResourceCnt := len(matchedResource.NetworkMap)
				for i, networkResource := range matchedResource.NetworkMap {
					networkResourceElmnt = "%s"
					if i != networkMapResourceCnt-1 {
						networkResourceElmnt += " -> "
					}

					networkMapGraph += fmt.Sprintf(networkResourceElmnt, networkResource)
				}

				log.Info("network map: [ ", networkMapGraph, " ]")
			}
		} else {
			log.Info("resource not found :( better luck next time!")
		}
	} else {
		if jsonOutput {
			output, err := json.Marshal(matchedResource)
			if err != nil {
				errMap := map[string]error{"error": err}
				errMapJSON, _ := json.Marshal(errMap)

				fmt.Printf("%s\n", errMapJSON)
			} else {
				fmt.Printf("%s\n", output)
			}
		} else {
			// plaintext
			if matchedResource.RID != "" {
				fmt.Println(matchedResource.RID)
				fmt.Printf("%s (%s)", matchedResource.AccountID, acctAliasFmted)
			} else {
				fmt.Println("not found")
			}
		}
	}
}

func runCloudSearch(ipAddr string, cloudSvc string, ipFuzzing bool, advIPFuzzing bool, orgSearch bool, orgSearchXaccountRoleARN string, orgSearchRoleName string, orgSearchOrgUnitID string, networkMapping bool, silent bool, jsonOutput bool) {
	// cloud connections
	log.Debug("generating AWS connection")
	ac, err := awsconnector.New()
	if err != nil {
		log.Fatal(err)
	}

	// search
	log.Info("searching for IP ", ipAddr, " in ", cloudSvc, " service(s)")
	searchCtlr := search.Search{AWSConn: ac, IpAddr: ipAddr}
	matchingResource, err := searchCtlr.InitSearch(cloudSvc, ipFuzzing, advIPFuzzing, orgSearch, orgSearchXaccountRoleARN, orgSearchRoleName, orgSearchOrgUnitID, networkMapping)
	if err != nil {
		log.Fatal("failed to run search :: [ ERR: ", err, " ]")
	}

	// output
	outputResults(matchingResource, networkMapping, silent, jsonOutput)
}

func main() {
	// CLI param parsing
	version := flag.Bool("version", false, "Outputs the version of IP2CR in use and exits")

	// output
	silentOutput := flag.Bool("silent", false, "If enabled, only output the results (default: false)")
	jsonOutput := flag.Bool("json", false, "Outputs results in JSON format; implies usage of --silent flag (default: false)")
	verboseOutput := flag.Bool("verbose", false, "Outputs all logs, from debug level to critical (default: false)")

	// base
	ipAddr := flag.String("ipaddr", "", "IP address to search for (default: 127.0.0.1)")
	cloudSvc := flag.String("svc", "all", "Specific cloud service(s) to search. Multiple services can be listed in CSV format, e.g. elbv1,elbv2. Available services are: [all, cloudfront , ec2 , elbv1 , elbv2]  (default: all)")

	// FEATURE FLAGS
	// IP fuzzing
	ipFuzzing := flag.Bool("ip-fuzzing", true, "Toggle the IP fuzzing feature to evaluate the IP and help optimize search (not recommended for small accounts due to overhead outweighing value)  (default: true)")
	advIPFuzzing := flag.Bool("adv-ip-fuzzing", true, "Toggle the advanced IP fuzzing feature to perform a more intensive heuristics evaluation to fuzz the service (not recommended for IPv6 addresses) (default: true)")

	// org search
	orgSearch := flag.Bool("org-search", false, "Search through all child accounts of the organization for resources, as well as target account (target account should be parent account) (default: false)")
	orgSearchXaccountRoleARN := flag.String("org-search-xaccount-role-arn", "", "The ARN of the role to assume for gathering AWS Organizations information for search, e.g. the role to assume with R/O access to your AWS Organizations account (default: '')")
	orgSearchRoleName := flag.String("org-search-role-name", "ip2cr", "The name of the role in each child account of an AWS Organization to assume when performing a search (default: ip2cr)")
	orgSearchOrgUnitID := flag.String("org-search-ou-id", "", "The ID of the AWS Organizations Organizational Unit to target when performing a search (default: '')")

	// network mapping
	networkMapping := flag.Bool("network-mapping", false, "If enabled, generate a network map associated with the identified resource if it's found (default: false)")

	flag.Parse()

	if *ipAddr == "" {
		log.Error("IP address is required")
		os.Exit(1)
	}

	if *version {
		fmt.Println("ip-2-cloudresource", APP_VER)
		return
	}

	if *jsonOutput {
		*silentOutput = true
	}
	if *silentOutput {
		log.SetOutput(io.Discard)
	}
	if *verboseOutput {
		log.SetLevel(log.DebugLevel)
	}

	// if the service(s) are specified, then we don't need to spend our time fuzzing the IP
	if *cloudSvc != "all" {
		*ipFuzzing = false
		*advIPFuzzing = false
	}

	log.Info("starting IP-2-CloudResource")

	utils.InitRollbar(APP_VER)

	rollbar.WrapAndWait(runCloudSearch, *ipAddr, *cloudSvc, *ipFuzzing, *advIPFuzzing, *orgSearch, *orgSearchXaccountRoleARN, *orgSearchRoleName, *orgSearchOrgUnitID, *networkMapping, *silentOutput, *jsonOutput)

	rollbar.Close()
}
