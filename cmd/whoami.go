// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"github.com/spf13/cobra"
    "github.com/keycloak/kcinit/console"
    "github.com/spf13/viper"
    "os"
    "fmt"
)

// loginCmd represents the login command
var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Who is currently logged in",
	Run: whoami,
}

func init() {
	rootCmd.AddCommand(whoamiCmd)
}

func whoami(cmd *cobra.Command, args []string) {
    realmUrl := viper.GetString(REALM_URL)
    if (realmUrl == "") {
        console.Writeln("No realm set up")
        os.Exit(1)
    }
    InitializeClient()
    masterToken, err := ReadToken(viper.GetString(LOGIN_CLIENT))
    if (err != nil || masterToken == nil) {
        console.Writeln("Not logged in")
        os.Exit(1)
    }

    res, err := Userinfo().Request().
        Header("Authorization", "bearer " + masterToken.AccessToken).
        Get()
    var info map[string]interface{}
    if (res == nil || err != nil || res.Status() != 200) {
        if (res != nil) {
            err = res.ReadJson(&info)
            if (info["error"] != "") {
                if (info["error_description"] != "") {
                    console.Writeln("Error:", info["error"], info["error_description"])
                } else {
                    console.Writeln("Error:", info["error]"])
                }
                os.Exit(1)
            }
        }
        console.Writeln("Error querying User Info service")
        os.Exit(1)
    }
    err = res.ReadJson(&info)
    if (info["preferred_username"] != "") {
        fmt.Println(info["preferred_username"])
        os.Exit(0)
    } else {
        fmt.Println("Unknown user")
        os.Exit(1)
    }







}
