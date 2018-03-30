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
    "os"
    "github.com/spf13/viper"
)

// logoutCmd represents the logout command
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Run: logout,
}

func logout(cmd *cobra.Command, args []string) {
    realmUrl := viper.GetString(REALM_URL)
    masterClient := viper.GetString(LOGIN_CLIENT)
    if (masterClient != "" && realmUrl != "") {
        InitializeClient()
        token, err := ReadToken(masterClient)
        if (token != nil && err == nil) {
            form := ClientForm()
            form.Set("refresh_token", token.RefreshToken)
            // don't care about response
            Logout().Request().Form(form).Post()
        }
    }
    os.RemoveAll(TokenDir())
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
