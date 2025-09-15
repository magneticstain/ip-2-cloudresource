package load_balancing

import (
	"context"
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"

	gcpcomputeapi "cloud.google.com/go/compute/apiv1"
	gcpcomputepbapi "cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"

	generalResource "github.com/magneticstain/ip-2-cloudresource/resource"
	"github.com/magneticstain/ip-2-cloudresource/utils"
)

type LoadBalancingPlugin struct {
	ProjectID string
}

func (lbp LoadBalancingPlugin) GetResources() ([]generalResource.Resource, error) {
	var gaClient *gcpcomputeapi.GlobalAddressesClient
	var lbGlobalAddrList *gcpcomputeapi.AddressIterator
	var lbResources []generalResource.Resource

	ctx := context.Background()

	gaClient, err := gcpcomputeapi.NewGlobalAddressesRESTClient(ctx)
	if err != nil {
		return lbResources, err
	}
	defer gaClient.Close() //nolint:errcheck

	req := &gcpcomputepbapi.ListGlobalAddressesRequest{
		Project: lbp.ProjectID,
	}

	lbGlobalAddrList = gaClient.List(ctx, req)

	for {
		addr, err := lbGlobalAddrList.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return lbResources, err
		}

		lbId := strconv.FormatUint(addr.GetId(), 10)
		lbName := addr.GetName()
		lbStatus := addr.GetStatus()
		lbIpAddr := addr.GetAddress()

		log.Debug("load balancer endpoint found - ID: ", lbId, ", Name: ", lbName, ", Status: ", lbStatus, ", IP: ", lbIpAddr)

		currentResource := generalResource.Resource{
			Id:             lbId,
			Name:           lbName,
			Status:         lbStatus,
			CloudSvc:       "load_balancing",
			AccountAliases: []string{lbp.ProjectID},
		}

		// for some god-awful reason, no value is being returned when calling the addr.GetIpVersion() method
		// as such, we will need to determine it ourselves :(
		ipVer, err := utils.DetermineIpAddrVersion(lbIpAddr)
		if err != nil {
			return lbResources, err
		}

		switch ipVer {
		case 4:
			currentResource.PublicIPv4Addrs = append(currentResource.PublicIPv4Addrs, lbIpAddr)
		case 6:
			currentResource.PublicIPv6Addrs = append(currentResource.PublicIPv6Addrs, lbIpAddr)
		default:
			return lbResources, fmt.Errorf("invalid IP version found for GCP LB; IP: %s, Version: IPv%d", lbIpAddr, ipVer)
		}

		lbResources = append(
			lbResources,
			currentResource,
		)
	}

	return lbResources, nil
}

func (lbp LoadBalancingPlugin) SearchResources(tgtIP string, matchingResource *generalResource.Resource) (generalResource.Resource, error) {
	log.Debug("fetching and searching load balancing resources")

	fetchedResources, err := lbp.GetResources()
	if err != nil {
		return *matchingResource, err
	}

	for _, lbResource := range fetchedResources {
		ridSlug := fmt.Sprintf("%s/%s", lbResource.Id, lbResource.Name)

		for _, ipv4Addr := range lbResource.PublicIPv4Addrs {
			if ipv4Addr == tgtIP {
				matchingResource.RID = ridSlug
				matchingResource.CloudSvc = "load_balancing"

				break
			}
		}

		for _, ipv6Addr := range lbResource.PublicIPv6Addrs {
			if ipv6Addr == tgtIP {
				matchingResource.RID = ridSlug
				matchingResource.CloudSvc = "load_balancing"

				break
			}
		}

		if matchingResource.RID != "" {
			log.Debug("IP found as Load Balancer -> ", matchingResource.RID)

			break
		}
	}

	return *matchingResource, nil
}
