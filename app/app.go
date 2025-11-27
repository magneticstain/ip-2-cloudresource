package app

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/rollbar/rollbar-go"
	log "github.com/sirupsen/logrus"

	"github.com/magneticstain/ip-2-cloudresource/resource"
	platformsearch "github.com/magneticstain/ip-2-cloudresource/search"
	"github.com/magneticstain/ip-2-cloudresource/utils"
)

const APP_ENV = "development"
const APP_VER = "v3.1.6"

func GetSupportedPlatforms() []string {
	return []string{"aws", "gcp", "azure"}
}

func OutputResults(matchedResource resource.Resource, networkMapping bool, silent bool, jsonOutput bool) {
	acctAliasFmted := strings.Join(matchedResource.AccountAliases, ", ")

	if !silent {
		if matchedResource.RID != "" {
			var acctStr string
			if matchedResource.AccountID == "current" {
				acctStr = "current account"
			} else {
				acctStr = fmt.Sprintf("account [ %s ( %s ) ]", matchedResource.AccountID, acctAliasFmted)
			}

			log.Info("resource found -> [ ", matchedResource.RID, " ] within ", matchedResource.CloudSvc, " service running in ", acctStr)

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

func RunCloudSearch(platform, tenantID, ipAddr, cloudSvc, orgSearchXaccountRoleARN, orgSearchRoleName, orgSearchOrgUnitID string, ipFuzzing, advIPFuzzing, orgSearch, networkMapping, silent, jsonOutput bool) {
	var err error

	platform = strings.ToLower(platform)
	supportedPlatforms := GetSupportedPlatforms()
	if !slices.Contains(supportedPlatforms, platform) {
		log.Fatal("'", platform, "' is not a supported platform")
		return
	}

	// search
	log.Info("searching for IP ", ipAddr, " in ", cloudSvc, " ", strings.ToUpper(platform), " service(s)")

	searchCtlr := platformsearch.Search{
		Platform: platform,
		TenantID: tenantID,
		IpAddr:   ipAddr,
	}

	_, err = searchCtlr.StartSearch(
		cloudSvc,
		ipFuzzing,
		advIPFuzzing,
		orgSearch,
		orgSearchXaccountRoleARN,
		orgSearchRoleName,
		orgSearchOrgUnitID,
		networkMapping,
	)
	if err != nil {
		log.Fatal(err)
		return
	}

	OutputResults(searchCtlr.MatchedResource, networkMapping, silent, jsonOutput)
}

func InitRollbar() {
	utils.InitRollbar(APP_ENV, APP_VER)
}

func WrapAndWait(fn interface{}, args ...interface{}) {
	rollbar.WrapAndWait(fn, args...)
}

func CloseRollbar() {
	rollbar.Close()
}
