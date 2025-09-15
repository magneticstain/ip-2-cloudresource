package compute

import (
	"context"
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"

	gcpcomputeapi "cloud.google.com/go/compute/apiv1"
	gcpcomputepbapi "cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"

	generalResource "github.com/magneticstain/ip-2-cloudresource/resource"
)

type ComputePlugin struct {
	ProjectID string
}

func CheckComputeIP(computeResource, matchingResource *generalResource.Resource, tgtIp string, ipVer int) (*generalResource.Resource, bool) {
	var ipSet []string
	var found bool

	log.Debug("checking target IP [ ", tgtIp, "] against IPv", ipVer, " resource address set")

	switch ipVer {
	case 4:
		ipSet = computeResource.PublicIPv4Addrs
	case 6:
		ipSet = computeResource.PublicIPv6Addrs
	}

	for _, ipAddr := range ipSet {
		if ipAddr == tgtIp {
			matchingResource.RID = fmt.Sprintf("%s/%s/%s", computeResource.AccountID, computeResource.Id, computeResource.Name)
			matchingResource.CloudSvc = "compute"

			found = true

			break
		}
	}

	return matchingResource, found
}

func GetPublicIPAddrsFromInstance(computeInstance *gcpcomputepbapi.Instance) ([]string, []string) {
	var publicIPv4Addrs, publicIPv6Addrs []string

	for _, networkIface := range computeInstance.GetNetworkInterfaces() {
		for _, accessConfig := range networkIface.AccessConfigs {
			if accessConfig.NatIP != nil {
				publicIPv4Addrs = append(publicIPv4Addrs, *accessConfig.NatIP)
			}

			if accessConfig.ExternalIpv6 != nil {
				publicIPv6Addrs = append(publicIPv6Addrs, *accessConfig.ExternalIpv6)
			}
		}
	}

	return publicIPv4Addrs, publicIPv6Addrs
}

func (comp ComputePlugin) GetResources() ([]generalResource.Resource, error) {
	var computeClient *gcpcomputeapi.InstancesClient
	var instanceList *gcpcomputeapi.InstancesScopedListPairIterator
	var computeResources []generalResource.Resource

	// REF: https://cloud.google.com/compute/docs/samples/compute-instances-list-all#compute_instances_list_all-go
	ctx := context.Background()

	computeClient, err := gcpcomputeapi.NewInstancesRESTClient(ctx)
	if err != nil {
		return computeResources, err
	}
	defer computeClient.Close() //nolint:errcheck

	req := &gcpcomputepbapi.AggregatedListInstancesRequest{
		Project: comp.ProjectID,
	}

	// normally, we would just return the iterator and allow the search function to iterate through each
	// however, for some reason, passing the iterator back to the search function results in a memory error when trying to read it
	// if someone in the future figures out why, feel free to refactor/patch it
	instanceList = computeClient.AggregatedList(ctx, req)

	for {
		instanceListPair, err := instanceList.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return computeResources, err
		}

		instances := instanceListPair.Value.Instances
		for _, instance := range instances {
			instanceId := strconv.FormatUint(instance.GetId(), 10)
			instanceName := instance.GetName()
			instanceStatus := instance.GetStatus()
			publicIPv4Addrs, publicIPv6Addrs := GetPublicIPAddrsFromInstance(instance)

			log.Debug("compute instance found - ID: ", instanceId, ", Name: ", instanceName, ", Status: ", instanceStatus)

			currentResource := generalResource.Resource{
				Id:              instanceId,
				AccountID:       comp.ProjectID,
				Name:            instanceName,
				Status:          instanceStatus,
				CloudSvc:        "compute",
				PublicIPv4Addrs: publicIPv4Addrs,
				PublicIPv6Addrs: publicIPv6Addrs,
			}

			computeResources = append(
				computeResources,
				currentResource,
			)
		}
	}

	return computeResources, nil
}

func (comp ComputePlugin) SearchResources(tgtIP string, matchingResource *generalResource.Resource) (generalResource.Resource, error) {
	log.Debug("fetching and searching compute resources")

	fetchedResources, err := comp.GetResources()
	if err != nil {
		return *matchingResource, err
	}

	var found bool
	for _, computeResource := range fetchedResources {
		// IPv4 is checked first
		matchingResource, found = CheckComputeIP(&computeResource, matchingResource, tgtIP, 4)

		if !found {
			// check IPv6 next
			matchingResource, found = CheckComputeIP(&computeResource, matchingResource, tgtIP, 6)
		}

		if found {
			log.Debug("IP found as Compute VM -> ", matchingResource.RID)

			break
		}
	}

	return *matchingResource, nil
}
