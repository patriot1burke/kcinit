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

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kcinit",
	Short: "Keycloak command line login utility.",
	Long: `Keycloak command line login utility allows you to login to a Keycloak realm and manage an sso session
between your command line tools.  Applications can use this tool in wrapper scripts to their command line applications to
retrieve and manage access tokens to back end REST services secured by Keycloak.

For example, let's say you have a command line application called 'oc'.  Your would write a wrapper script like the following:

#!/bin/sh
oc --token=$(kcinit token oc) $@

kcinit would then prompt for login credentials and create an access token for the oc client.

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

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.keycloak/kcinit/kcinit.yaml)")
    rootCmd.PersistentFlags().BoolVar(&console.NoMask, "nomask", false, "")
    rootCmd.PersistentFlags().MarkHidden("nomask")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
    //rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
        path := ConfigPath()
        viper.SetConfigFile(path + "/kcinit.yaml")
    }
    viper.AutomaticEnv() // read in environment variables that match

    // If a config file is found, read it in.
    if err := viper.ReadInConfig(); err == nil {
        //fmt.Println("Using config file:", viper.ConfigFileUsed())
        InitializeClient()
    }
}


func ConfigPath() string {
    path := os.Getenv("KC_CONFIG_PATH")
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
    client := viper.GetString("client")
    if (client == "") {
        console.Writeln("Not configured.  Please run the `install` command")
        os.Exit(1)
    }
}
