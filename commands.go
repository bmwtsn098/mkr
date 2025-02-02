package main

import (
	"os"
	"strings"
	"time"

	"github.com/Songmu/prompter"
	"github.com/mackerelio/mackerel-client-go"
	"github.com/mackerelio/mkr/aws_integrations"
	"github.com/mackerelio/mkr/channels"
	"github.com/mackerelio/mkr/checks"
	"github.com/mackerelio/mkr/format"
	"github.com/mackerelio/mkr/hosts"
	"github.com/mackerelio/mkr/logger"
	"github.com/mackerelio/mkr/mackerelclient"
	"github.com/mackerelio/mkr/org"
	"github.com/mackerelio/mkr/plugin"
	"github.com/mackerelio/mkr/services"
	"github.com/mackerelio/mkr/wrap"
	"github.com/urfave/cli"
)

// Commands cli.Command object list
var Commands = []cli.Command{
	commandStatus,
	hosts.CommandHosts,
	hosts.CommandCreate,
	commandUpdate,
	commandThrow,
	commandMetrics,
	commandFetch,
	commandRetire,
	services.Command,
	commandMonitors,
	channels.Command,
	commandAlerts,
	commandDashboards,
	commandAnnotations,
	org.Command,
	plugin.CommandPlugin,
	checks.Command,
	wrap.Command,
	aws_integrations.Command,
}

var commandStatus = cli.Command{
	Name:      "status",
	Usage:     "Show the host",
	ArgsUsage: "[--verbose | -v] <hostId>",
	Description: `
    Show the information of the host identified with <hostId>.
    Requests "GET /api/v0/hosts/<hostId>". See https://mackerel.io/api-docs/entry/hosts#get .
`,
	Action: doStatus,
	Flags: []cli.Flag{
		cli.BoolFlag{Name: "verbose, v", Usage: "Verbose output mode"},
	},
}

var commandUpdate = cli.Command{
	Name:      "update",
	Usage:     "Update the host",
	ArgsUsage: "[--name | -n <name>] [--displayName <displayName>] [--status | -st <status>] [--roleFullname | -R <service:role>] [--overwriteRoles | -o] [--memo <memo>] [<hostIds...>]",
	Description: `
    Update the host identified with <hostId>.
    Requests "PUT /api/v0/hosts/<hostId>". See https://mackerel.io/api-docs/entry/hosts#update-information .
`,
	Action: doUpdate,
	Flags: []cli.Flag{
		cli.StringFlag{Name: "name, n", Value: "", Usage: "Update hostname."},
		cli.StringFlag{Name: "displayName", Value: "", Usage: "Update displayName."},
		cli.StringFlag{Name: "status, st", Value: "", Usage: "Update status."},
		cli.StringSliceFlag{
			Name:  "roleFullname, R",
			Value: &cli.StringSlice{},
			Usage: "Update rolefullname.",
		},
		cli.BoolFlag{Name: "overwriteRoles, o", Usage: "Overwrite roles instead of adding specified roles."},
		cli.StringFlag{Name: "memo", Value: "", Usage: "memo for the Host"},
	},
}

var commandMetrics = cli.Command{
	Name:      "metrics",
	Usage:     "Fetch metric values",
	ArgsUsage: "[--host | -H <hostId>] [--service | -s <service>] [--name | -n <metricName>] --from int --to int",
	Description: `
    Fetch metric values of 'host metric' or 'service metric'.
    Requests "/api/v0/hosts/<hostId>/metrics" or "/api/v0/services/<serviceName>/tsdb".
    See https://mackerel.io/api-docs/entry/host-metrics#get, https://mackerel.io/api-docs/entry/service-metrics#get.
`,
	Action: doMetrics,
	Flags: []cli.Flag{
		cli.StringFlag{Name: "host, H", Value: "", Usage: "Fetch host metric values of <hostID>."},
		cli.StringFlag{Name: "service, s", Value: "", Usage: "Fetch service metric values of <service>."},
		cli.StringFlag{Name: "name, n", Value: "", Usage: "The name of the metric for which you want to obtain the metric."},
		cli.Int64Flag{Name: "from", Usage: "The first of the period for which you want to obtain the metric. (epoch seconds)"},
		cli.Int64Flag{Name: "to", Usage: "The end of the period for which you want to obtain the metric. (epoch seconds)"},
	},
}

var commandFetch = cli.Command{
	Name:      "fetch",
	Usage:     "Fetch latest metric values",
	ArgsUsage: "[--name | -n <metricName>] hostIds...",
	Description: `
    Fetch latest metric values about the hosts.
    Requests "GET /api/v0/tsdb/latest". See https://mackerel.io/api-docs/entry/host-metrics#get-latest .
`,
	Action: doFetch,
	Flags: []cli.Flag{
		cli.StringSliceFlag{
			Name:  "name, n",
			Value: &cli.StringSlice{},
			Usage: "Fetch metric values identified with <name>. Required. Multiple choices are allowed. ",
		},
	},
}

var commandRetire = cli.Command{
	Name:      "retire",
	Usage:     "Retire hosts",
	ArgsUsage: "[--force] hostIds...",
	Description: `
    Retire host identified by <hostId>. Be careful because this is an irreversible operation.
    Requests POST /api/v0/hosts/<hostId>/retire parallelly. See https://mackerel.io/api-docs/entry/hosts#retire .
`,
	Action: doRetire,
	Flags: []cli.Flag{
		cli.BoolFlag{Name: "force", Usage: "Force retirement without confirmation."},
	},
}

func doStatus(c *cli.Context) error {
	confFile := c.GlobalString("conf")
	argHostID := c.Args().Get(0)
	isVerbose := c.Bool("verbose")

	if argHostID == "" {
		if argHostID = mackerelclient.LoadHostIDFromConfig(confFile); argHostID == "" {
			cli.ShowCommandHelpAndExit(c, "status", 1)
		}
	}

	host, err := mackerelclient.NewFromContext(c).FindHost(argHostID)
	logger.DieIf(err)

	if isVerbose {
		err := format.PrettyPrintJSON(os.Stdout, host)
		logger.DieIf(err)
	} else {
		err := format.PrettyPrintJSON(os.Stdout, &format.Host{
			ID:            host.ID,
			Name:          host.Name,
			DisplayName:   host.DisplayName,
			Status:        host.Status,
			RoleFullnames: host.GetRoleFullnames(),
			IsRetired:     host.IsRetired,
			CreatedAt:     format.ISO8601Extended(host.DateFromCreatedAt()),
			IPAddresses:   host.IPAddresses(),
		})
		logger.DieIf(err)
	}
	return nil
}

func doUpdate(c *cli.Context) error {
	confFile := c.GlobalString("conf")
	argHostIDs := c.Args()
	optName := c.String("name")
	optDisplayName := c.String("displayName")
	optStatus := c.String("status")
	optRoleFullnames := c.StringSlice("roleFullname")
	overwriteRoles := c.Bool("overwriteRoles")
	optMemo := c.String("memo")

	if len(argHostIDs) < 1 {
		argHostIDs = make([]string, 1)
		if argHostIDs[0] = mackerelclient.LoadHostIDFromConfig(confFile); argHostIDs[0] == "" {
			cli.ShowCommandHelpAndExit(c, "update", 1)
		}
	}

	needUpdateHostStatus := optStatus != ""
	needUpdateRolesInHostUpdate := !overwriteRoles && len(optRoleFullnames) > 0
	needUpdateHost := (optName != "" || optDisplayName != "" || overwriteRoles || optMemo != "" || needUpdateRolesInHostUpdate)

	if !needUpdateHostStatus && !needUpdateHost {
		logger.Log("update", "at least one argumet is required.")
		cli.ShowCommandHelpAndExit(c, "update", 1)
	}

	client := mackerelclient.NewFromContext(c)

	for _, hostID := range argHostIDs {
		if needUpdateHostStatus {
			err := client.UpdateHostStatus(hostID, optStatus)
			logger.DieIf(err)
		}

		if overwriteRoles {
			err := client.UpdateHostRoleFullnames(hostID, optRoleFullnames)
			logger.DieIf(err)
		}

		if needUpdateHost {
			host, err := client.FindHost(hostID)
			logger.DieIf(err)
			name := ""
			if optName == "" {
				name = host.Name
			} else {
				name = optName
			}
			displayname := ""
			if optDisplayName == "" {
				displayname = host.DisplayName
			} else {
				displayname = optDisplayName
			}
			memo := ""
			if optMemo == "" {
				memo = host.Memo
			} else {
				memo = optMemo
			}
			param := &mackerel.UpdateHostParam{
				Name:        name,
				DisplayName: displayname,
				Meta:        host.Meta,
				Interfaces:  host.Interfaces,
				Memo:        memo,
			}
			if needUpdateRolesInHostUpdate {
				param.RoleFullnames = optRoleFullnames
			}
			_, err = client.UpdateHost(hostID, param)
			logger.DieIf(err)
		}

		logger.Log("updated", hostID)
	}
	return nil
}

func split(ids []string, count int) [][]string {
	xs := make([][]string, 0, (len(ids)+count-1)/count)
	for i, name := range ids {
		if i%count == 0 {
			xs = append(xs, []string{})
		}
		xs[len(xs)-1] = append(xs[len(xs)-1], name)
	}
	return xs
}

func doMetrics(c *cli.Context) error {
	optHostID := c.String("host")
	optService := c.String("service")
	optMetricName := c.String("name")

	from := c.Int64("from")
	to := c.Int64("to")
	if to == 0 {
		to = time.Now().Unix()
	}

	client := mackerelclient.NewFromContext(c)

	if optHostID != "" {
		metricValue, err := client.FetchHostMetricValues(optHostID, optMetricName, from, to)
		logger.DieIf(err)

		err = format.PrettyPrintJSON(os.Stdout, metricValue)
		logger.DieIf(err)
	} else if optService != "" {
		metricValue, err := client.FetchServiceMetricValues(optService, optMetricName, from, to)
		logger.DieIf(err)

		err = format.PrettyPrintJSON(os.Stdout, metricValue)
		logger.DieIf(err)
	} else {
		cli.ShowCommandHelpAndExit(c, "metrics", 1)
	}
	return nil
}

func doFetch(c *cli.Context) error {
	argHostIDs := c.Args()
	optMetricNames := c.StringSlice("name")

	if len(argHostIDs) < 1 || len(optMetricNames) < 1 {
		cli.ShowCommandHelpAndExit(c, "fetch", 1)
	}

	allMetricValues := make(mackerel.LatestMetricValues)
	// Fetches 100 hosts per one request (to avoid URL maximum length).
	for _, hostIds := range split(argHostIDs, 100) {
		metricValues, err := mackerelclient.NewFromContext(c).FetchLatestMetricValues(hostIds, optMetricNames)
		logger.DieIf(err)
		for key := range metricValues {
			allMetricValues[key] = metricValues[key]
		}
	}

	err := format.PrettyPrintJSON(os.Stdout, allMetricValues)
	logger.DieIf(err)
	return nil
}

func doRetire(c *cli.Context) error {
	confFile := c.GlobalString("conf")
	force := c.Bool("force")
	argHostIDs := c.Args()

	if len(argHostIDs) < 1 {
		argHostIDs = make([]string, 1)
		if argHostIDs[0] = mackerelclient.LoadHostIDFromConfig(confFile); argHostIDs[0] == "" {
			cli.ShowCommandHelpAndExit(c, "retire", 1)
		}
	}

	if !force && !prompter.YN("Retire following hosts.\n  "+strings.Join(argHostIDs, "\n  ")+"\nAre you sure?", true) {
		logger.Log("", "retirement is canceled.")
		return nil
	}

	client := mackerelclient.NewFromContext(c)

	for _, hostID := range argHostIDs {
		err := client.RetireHost(hostID)
		logger.DieIf(err)

		logger.Log("retired", hostID)
	}
	return nil
}
