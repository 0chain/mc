/*
 * Minio Client (C) 2015 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"github.com/fatih/color"
	"github.com/minio/cli"
	"github.com/minio/mc/pkg/console"
	"github.com/minio/minio-xl/pkg/probe"
)

// command specific flags.
var (
	updateFlags = []cli.Flag{
		cli.BoolFlag{
			Name:  "help, h",
			Usage: "Help for update.",
		},
		cli.BoolFlag{
			Name:  "experimental, E",
			Usage: "Check experimental update.",
		},
	}
)

// Check for new software updates.
var updateCmd = cli.Command{
	Name:   "update",
	Usage:  "Check for a new software update.",
	Action: mainUpdate,
	Flags:  append(updateFlags, globalFlags...),
	CustomHelpTemplate: `Name:
   mc {{.Name}} - {{.Usage}}

USAGE:
   mc {{.Name}} [FLAGS]

FLAGS:
  {{range .Flags}}{{.}}
  {{end}}
EXAMPLES:
   1. Check for any new official release.
      $ mc {{.Name}}

   2. Check for any new experimental release.
      $ mc {{.Name}} --experimental
`,
}

// update URL endpoints.
const (
	mcUpdateStableURL       = "https://dl.minio.io:9000/updates/updates.json"
	mcUpdateExperimentalURL = "https://dl.minio.io:9000/updates/experimental.json"
)

// mcUpdates container to hold updates json.
type mcUpdates struct {
	BuildDate string
	Platforms map[string]string
}

// updateMessage container to hold update messages.
type updateMessage struct {
	Status   string `json:"status"`
	Update   bool   `json:"update"`
	Download string `json:"downloadURL"`
	Version  string `json:"version"`
}

// String colorized update message.
func (u updateMessage) String() string {
	if !u.Update {
		return console.Colorize("Update", "You are already running the most recent version of ‘mc’.")
	}
	var msg string
	if runtime.GOOS == "windows" {
		msg = "mc.exe cp " + u.Download + " .\\mc.exe"
	} else {
		msg = "mc cp " + u.Download + " ./mc.new; chmod 755 ./mc.new"
	}
	msg, err := colorizeUpdateMessage(msg)
	fatalIf(err.Trace(msg), "Unable to colorize experimental update notification string ‘"+msg+"’.")
	return msg
}

// JSON jsonified update message.
func (u updateMessage) JSON() string {
	u.Status = "success"
	updateMessageJSONBytes, err := json.Marshal(u)
	fatalIf(probe.NewError(err), "Unable to marshal into JSON.")

	return string(updateMessageJSONBytes)
}

// verify updates for releases.
func getReleaseUpdate(updateURL string) {
	clnt, err := url2Client(updateURL)
	fatalIf(err.Trace(updateURL), "Unable to initalize update URL.")

	data, err := clnt.Get(0, 0)
	fatalIf(err.Trace(updateURL), "Unable to read from update URL ‘"+updateURL+"’.")

	if mcVersion == "UNOFFICIAL.GOGET" {
		fatalIf(errDummy().Trace(mcVersion),
			"Update mechanism is not supported for ‘go get’ based binary builds.  Please download official releases from https://minio.io/#mc")
	}

	current, e := time.Parse(time.RFC3339, mcVersion)
	fatalIf(probe.NewError(e), "Unable to parse version string as time.")

	if current.IsZero() {
		fatalIf(errDummy().Trace(mcVersion),
			"Updates not supported for custom builds. Version field is empty. Please download official releases from https://minio.io/#mc")
	}

	var updates mcUpdates
	decoder := json.NewDecoder(data)
	e = decoder.Decode(&updates)
	fatalIf(probe.NewError(e), "Unable to decode update notification.")

	latest, e := time.Parse(time.RFC3339, updates.BuildDate)
	if e != nil {
		latest, e = time.Parse(http.TimeFormat, updates.BuildDate)
		fatalIf(probe.NewError(e), "Unable to parse BuildDate.")
	}

	if latest.IsZero() {
		fatalIf(errDummy().Trace(mcVersion),
			"Unable to validate any update available at this time. Please open an issue at https://github.com/minio/mc/issues")
	}

	updateURLParse := clnt.GetURL()
	downloadURL := updateURLParse.Scheme +
		string(updateURLParse.SchemeSeparator) +
		updateURLParse.Host + string(updateURLParse.Separator) +
		updates.Platforms[runtime.GOOS+"-"+runtime.GOARCH]

	updateMsg := updateMessage{
		Download: downloadURL,
		Version:  mcVersion,
	}
	if latest.After(current) {
		updateMsg.Update = true
	}
	printMsg(updateMsg)
}

// main entry point for update command.
func mainUpdate(ctx *cli.Context) {
	// Set global flags from context.
	setGlobalsFromContext(ctx)

	// Additional command speific theme customization.
	console.SetColor("Update", color.New(color.FgGreen, color.Bold))

	// Check for update.
	if ctx.Bool("experimental") {
		getReleaseUpdate(mcUpdateExperimentalURL)
	} else {
		getReleaseUpdate(mcUpdateStableURL)
	}
}
