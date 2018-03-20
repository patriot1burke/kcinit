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
	"github.com/keycloak/kcinit/console"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Interactive configuration of kcinit",
	Long: `Interactive configuration of kcinit.  Will prompt you for authentication server URL, realm, and other settings.`,
	Run: install,
}

func install(cmd *cobra.Command, args []string) {
	server := console.ReadDefault("Authentication server URL", "http://localhost:8080/auth")
	realm := console.ReadDefault("Name of realm", "master")
	client := console.ReadDefault("OAuth client id", "kcinit")
	clientSecret := console.ReadLine("Client secret [none]: ")

	viper.Set("server", server)
	viper.Set("realm", realm)
	viper.Set("client", client)
	viper.Set("clientSecret", clientSecret)

	os.MkdirAll(ConfigPath(), 0700)
	configPath := ConfigPath() + "/kcinit.yaml"
	viper.SetConfigFile(configPath)
	viper.WriteConfig()
}

func init() {
	rootCmd.AddCommand(installCmd)
}
