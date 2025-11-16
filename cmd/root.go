package cmd

import (
	"fmt"
	"io"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/magneticstain/ip-2-cloudresource/app"
)

var (
	// Common flags
	silentOutput   bool
	jsonOutput     bool
	verboseOutput  bool
	platform       string
	ipAddr         string
	cloudSvc       string
	tenantID       string
	ipFuzzing      bool
	advIPFuzzing   bool
	orgSearch      bool
	networkMapping bool

	// AWS Organization specific flags
	orgSearchXaccountRoleARN string
	orgSearchRoleName        string
	orgSearchOrgUnitID       string
)

var rootCmd = &cobra.Command{
	Use:     "ip-2-cloudresource",
	Short:   "A tool for searching cloud resources by IP address",
	Version: app.APP_VER,
	RunE: func(cmd *cobra.Command, args []string) error {
		if ipAddr == "" {
			return fmt.Errorf("IP address is required")
		}

		if jsonOutput {
			silentOutput = true
		}
		if silentOutput {
			log.SetOutput(io.Discard)
		}
		if verboseOutput {
			log.SetLevel(log.DebugLevel)
		}

		// if the service(s) are specified, then we don't need to spend our time fuzzing the IP
		if cloudSvc != "all" {
			ipFuzzing = false
			advIPFuzzing = false
		}

		// modify flags based on platform's supported feature set
		switch {
		case platform != "aws":
			ipFuzzing = false
			advIPFuzzing = false
			orgSearch = false
			networkMapping = false
		case platform == "gcp", platform == "azure":
			if tenantID == "" {
				return fmt.Errorf("tenant ID is required for searching %s", strings.ToUpper(platform))
			}
		}

		log.Info("starting IP-2-CloudResource")

		app.InitRollbar()
		app.WrapAndWait(app.RunCloudSearch,
			platform,
			tenantID,
			ipAddr,
			cloudSvc,
			orgSearchXaccountRoleARN,
			orgSearchRoleName,
			orgSearchOrgUnitID,
			ipFuzzing,
			advIPFuzzing,
			orgSearch,
			networkMapping,
			silentOutput,
			jsonOutput,
		)
		app.CloseRollbar()

		return nil
	},
}

func init() {
	// Output flags
	rootCmd.Flags().BoolVar(&silentOutput, "silent", false, "If enabled, only output the results")
	rootCmd.Flags().BoolVar(&jsonOutput, "json", false, "Outputs results in JSON format; implies usage of --silent flag")
	rootCmd.Flags().BoolVar(&verboseOutput, "verbose", false, "Outputs all logs, from debug level to critical")

	// Base flags
	// TODO: change to separate subcommands per platform
	rootCmd.Flags().StringVar(&platform, "platform", "aws", "Platform to target for IP search (supported values: aws, gcp, azure)")
	rootCmd.Flags().StringVar(&ipAddr, "ipaddr", "", "IP address to search for")
	// TODO: change to separate subcommands per service
	rootCmd.Flags().StringVar(&cloudSvc, "svc", "all", "Specific cloud service(s) to search. Multiple services can be listed in CSV format, e.g. elbv1,elbv2. Available services are: [all, cloudfront , ec2 , elbv1 , elbv2]")
	rootCmd.Flags().StringVar(&tenantID, "tenant-id", "", "For cloud platforms that require or support it, set this to the ID of the target tenant (e.g. project, account, subscription, etc) ID to search")

	// Feature flags
	rootCmd.Flags().BoolVar(&ipFuzzing, "ip-fuzzing", true, "Toggle the IP fuzzing feature to evaluate the IP and help optimize search (not recommended for small accounts due to overhead outweighing value)")
	rootCmd.Flags().BoolVar(&advIPFuzzing, "adv-ip-fuzzing", true, "Toggle the advanced IP fuzzing feature to perform a more intensive heuristics evaluation to fuzz the service (not recommended for IPv6 addresses)")
	rootCmd.Flags().BoolVar(&orgSearch, "org-search", false, "Search through all child accounts of the organization for resources, as well as target account (target account should be parent account)")
	rootCmd.Flags().StringVar(&orgSearchXaccountRoleARN, "org-search-xaccount-role-arn", "", "The ARN of the role to assume for gathering AWS Organizations information for search, e.g. the role to assume with R/O access to your AWS Organizations account")
	rootCmd.Flags().StringVar(&orgSearchRoleName, "org-search-role-name", "ip2cr", "The name of the role in each child account of an AWS Organization to assume when performing a search")
	rootCmd.Flags().StringVar(&orgSearchOrgUnitID, "org-search-ou-id", "", "The ID of the AWS Organizations Organizational Unit to target when performing a search")
	rootCmd.Flags().BoolVar(&networkMapping, "network-mapping", false, "If enabled, generate a network map associated with the identified resource if it's found")

	if err := rootCmd.MarkFlagRequired("ipaddr"); err != nil {
		panic(err)
	}
}

// Execute is exported so main.go can call from cmd package
func Execute() error {
	return rootCmd.Execute()
}
