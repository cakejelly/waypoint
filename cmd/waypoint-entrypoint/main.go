package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"

	"github.com/hashicorp/go-hclog"

	"github.com/hashicorp/waypoint/internal/ceb"
	"github.com/hashicorp/waypoint/internal/pkg/signalcontext"
)

func main() {
	os.Exit(realMain())
}

const (
	DefaultPort                = 5000
	DefaultWaypointControlAddr = "control.alpha.waypoint.run"
)

func realMain() int {
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		usage()
		return 1
	}

	// TODO(mitchellh): proper log setup
	log := hclog.L()
	hclog.L().SetLevel(hclog.Trace)

	// Create a context that is cancelled on interrupt
	ctx, closer := signalcontext.WithInterrupt(context.Background(), log)
	defer closer()

	options := []ceb.Option{
		ceb.WithEnvDefaults(),
		ceb.WithExec(args),
	}

	labels := os.Getenv("WAYPOINT_URL_LABELS")
	if labels != "" {
		var port int

		portStr := os.Getenv("PORT")
		if portStr == "" {
			port = DefaultPort
			os.Setenv("PORT", strconv.Itoa(DefaultPort))
		} else {
			i, err := strconv.Atoi(portStr)
			if err != nil {
				fmt.Fprintf(flag.CommandLine.Output(), "Invalid value of PORT: %s\n", err)
				return 1
			}

			port = i
		}

		controlAddr := os.Getenv("WAYPOINT_CONTROL_ADDR")
		if controlAddr == "" {
			controlAddr = DefaultWaypointControlAddr
		}

		token := os.Getenv("WAYPOINT_TOKEN")
		if token == "" {
			fmt.Fprintf(flag.CommandLine.Output(), "No token provided via WAYPOINT_TOKEN.\n")
			return 1
		}

		options = append(options, ceb.WithURLService(controlAddr, token, port, labels))
	}

	// Run our core logic
	err := ceb.Run(ctx, options...)
	if err != nil {
		fmt.Fprintf(flag.CommandLine.Output(),
			"Error initializing Waypoint entrypoint: %s\n", err)
		return 1
	}

	return 0
}

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(),
		strings.TrimLeftFunc(usageText, unicode.IsSpace),
		os.Args[0])
	flag.PrintDefaults()
}

const usageText = `
Usage: %[1]s [cmd] [args...]

    This the custom entrypoint to support Waypoint. It will re-execute any
    command given after configuring the environment for usage with Waypoint.

`
