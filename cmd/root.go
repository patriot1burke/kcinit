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
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
    "regexp"
    "github.com/keycloak/kcinit/console"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kcinit",
	Short: "Keycloak command line login utility.",
	Long: `Keycloak command line login utility allows you to login to a Keycloak realm and manage an sso session
between your command line tools.  Command line applications can use this tool to obtain access tokens they
need to invoke on back end REST services secured by Keycloak.

For example, let's say you have a command line application called 'kubectl' that accepts a '--token' flag.  You could 
create an alias as follows:

$ alias kubectl='kubectl --token=$(kcinit token kubernetes)'

kcinit would then prompt for login credentials and create an access token targeted for the kubernetes client.

Finally, all global command line switches for kcinit can be specified instead using environment varialbes with the prefix of "KCINIT" and
a underscore "_" separators.  So, for example --realm-url could be specified by setting an environment variable KCINIT_REALM_URL.

`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.SetOutput(os.Stderr);
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var configdir string

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
    rootCmd.PersistentFlags().StringVar(&configdir, "config", "", "config directory (default is $HOME/.keycloak/kcinit).")
    rootCmd.PersistentFlags().String("realm-url", "", "realm endpoint.")
    rootCmd.PersistentFlags().String("login-client", "", "client used for login requests.")
    rootCmd.PersistentFlags().String("login-secret", "", "client secret used for login requests.")
    rootCmd.PersistentFlags().Bool(SAVE, false, "Store tokens on disk.  Defaults to true.")
    rootCmd.PersistentFlags().BoolVar(&console.NoMask, "nomask", false, "")
    rootCmd.PersistentFlags().MarkHidden("nomask")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
    //rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

const REALM_URL = "realm_url"
const LOGIN_CLIENT = "login_client"
const LOGIN_SECRET = "login_secret"
const SAVE = "save"



// initConfig reads in config file and ENV variables if set.
func initConfig() {
    viper.SetConfigFile(ConfigPath() + "/kcinit.yaml")
    viper.SetEnvPrefix("kcinit")
    viper.AutomaticEnv() // read in environment variables that match
    viper.BindEnv(REALM_URL)
    viper.BindPFlag(REALM_URL, rootCmd.Flags().Lookup("realm-url"))
    viper.BindEnv(LOGIN_CLIENT)
    viper.BindPFlag(LOGIN_CLIENT, rootCmd.Flags().Lookup("login-client"))
    viper.SetDefault(LOGIN_CLIENT, "kcinit")
    viper.BindEnv(LOGIN_SECRET)
    viper.BindPFlag(LOGIN_SECRET, rootCmd.Flags().Lookup("login-secret"))
    viper.BindEnv(SAVE)
    viper.BindPFlag(SAVE, rootCmd.Flags().Lookup(SAVE))
    viper.SetDefault(SAVE, true)




    // If a config file is found, read it in.
    if err := viper.ReadInConfig(); err == nil {
        //fmt.Println("Using config file:", viper.ConfigFileUsed())
    }
}


func ConfigPath() string {
    if (configdir != "") {
        return configdir
    }
    path := os.Getenv("KCINIT_CONFIG")
    if (path == "") {
        // Find home directory.
        home, err := homedir.Dir()
        if err != nil {
            fmt.Println(err)
            os.Exit(1)
        }
        path = home + "/.keycloak/kcinit"

    }
    return path
}

func CreateTokenDir() {
    os.MkdirAll(TokenDir(), 0700)
}

func TokenFile(client string) string {
    reg, _ := regexp.Compile("[^a-zA-Z0-9-_.]+")

    client = reg.ReplaceAllString(client, "_")
    return TokenDir() + "/" + client
}

func TokenDir() string {
    return ConfigPath() + "/tokens"
}

func CheckInstalled() {
    InitializeClient()
}
