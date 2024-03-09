package plugin

import (
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"

	// "github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	generalResource "github.com/magneticstain/ip-2-cloudresource/resource"
)

type AzureController struct {
	AzureConn azidentity.DefaultAzureCredential
}

func New() (AzureController, error) {
	azConn, err := ConnectToAzure()
	if err != nil {
		return AzureController{}, err
	}

	azc := AzureController{AzureConn: azConn}

	return azc, err
}

func ConnectToAzure() (azidentity.DefaultAzureCredential, error) {
	log.Debug("connecting to Azure using default credentials")
	dac, err := azidentity.NewDefaultAzureCredential(nil)

	return *dac, err
}

func GetSupportedSvcs() []string {
	return []string{
		"virtual_machines",
	}
}

func (azctrlr *AzureController) SearchAzureSvc(subscriptionID, ipAddr, cloudSvc string, matchingResource *generalResource.Resource) (generalResource.Resource, error) {
	// var err error

	log.Debug("searching ", cloudSvc, " in subscription ", subscriptionID, "using Azure controller")

	switch cloudSvc {
	case "virtual_machines":
		// comp := compute.ComputePlugin{
		// 	ProjectID: projectID,
		// }
		// _, err = comp.SearchResources(ipAddr, matchingResource)
		// if err != nil {
		// 	return *matchingResource, err
		// }
	default:
		msg := fmt.Sprintf("unknown Azure service provided: '%s'", cloudSvc)

		return *matchingResource, errors.New(msg)
	}

	return *matchingResource, nil
}
