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
	"time"
	"encoding/json"
	"github.com/keycloak/kcinit/console"
	"io/ioutil"
    "os"
    "errors"
    "fmt"
    "github.com/spf13/viper"
)

// tokenCmd represents the token command
var tokenCmd = &cobra.Command{
	Use:   "token [client]",
	Short: "Output token to stdout",
	Long: "Output token to stdout.  If not logged in, this command will interactively login, before trying to obtain a token for the client.  The login client is used if client argument not specified.",
	Args: cobra.MaximumNArgs(1),
	Run: token,
}

type ResponseType struct {
    Status int
}

type SpecType struct {
    Interactive bool
    Response ResponseType
}

type ExecCredential struct {
    Spec SpecType
}

func formatToken(token *AccessTokenResponse, tokenType string) string {
    if (tokenType == "access") {
        return token.AccessToken
    } else if (tokenType == "id") {
        return token.IdToken
    } else if (tokenType == "refresh") {
        return token.RefreshToken
    } else {
        panic("Illegal token type: " + tokenType)
    }
}

func tokenOutput(token *AccessTokenResponse, tokenType string) {
    execInfo := os.Getenv("KUBERNETES_EXEC_INFO")
    if (execInfo == "") {
        fmt.Fprint(os.Stdout, formatToken(token, tokenType))
    } else {
        var data ExecCredential
        json.Unmarshal([]byte(execInfo), &data)
        console.Writeln("KUBERNETES_EXEC_INFO", execInfo)
        console.Writeln()
        output := map[string]interface{} {
            "apiVersion": "client.authentication.k8s.io/v1alpha1",
            "kind": "ExecCredential",
            "status": map[string]string {
                "token": formatToken(token, tokenType),
                "expirationTimestamp": time.Unix(token.ExpiresIn, 0).Format(time.RFC3339),
            },
        }
        b, _ := json.Marshal(output)
        fmt.Fprint(os.Stdout, string(b))
    }
}

func token(cmd *cobra.Command, args []string) {
    tokenType, _ := cmd.Flags().GetString("token-type")
    if (tokenType != "access" && tokenType != "refresh" && tokenType != "id") {
        console.Writeln("Illegal token type requested.")
        os.Exit(1)

    }

    CheckInstalled()
    client := viper.GetString(LOGIN_CLIENT)
    masterClient := client
    if (len(args) == 1) {
        client = args[0]
    }

    if (!loginFlags.force) {
        token, err := ReadToken(client)
        if (err == nil) {
            tokenOutput(token, tokenType)
            return
        }
    }
    if (client == masterClient) {
        masterToken := DoLogin()
        tokenOutput(masterToken, tokenType)
        return
    }
    masterToken, err := ReadToken(masterClient)
    if (err != nil) {
        masterToken = DoLogin()
    }

    form := ClientForm()
    form.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
    form.Set("subject_token", masterToken.AccessToken)
    form.Set("subject_token_type", "urn:ietf:params:oauth:token-type:access_token")
    form.Set("requested_token_type", "urn:ietf:params:oauth:token-type:refresh_token")
    form.Set("audience", client)

    res, err := Token().Request().Form(form).Post()

    if (err != nil) {
        console.Writeln("Failure: connection failed")
        os.Exit(1)
    }

    if (res.Status() != 200) {
        if (res.MediaType() != "") {
            var json map[string]interface{}
            err := res.ReadJson(&json)
            if (err == nil) {
                console.Writeln("Failure: failed to exchange token:", json["error"], json["error_description"])
                os.Exit(1)

            }
        }
        console.Writeln("Failure: failed to exchange token")
        os.Exit(1)

    }
    var tokenResponse AccessTokenResponse
    res.ReadJson(&tokenResponse)
    tokenResponse.ProcessTokenResponse(client)
    tokenOutput(&tokenResponse, tokenType)
}

func (tokenResponse *AccessTokenResponse) ProcessTokenResponse(client string) {
    if (!viper.GetBool(SAVE)) {
        return
    }
	tokenResponse.ExpiresIn = tokenResponse.ExpiresIn + time.Now().Unix()
	buf, _ := json.Marshal(tokenResponse)
	tokenFile := TokenFile(client)
	console.Traceln("Writing to file: ", tokenFile)
	CreateTokenDir()
	ioutil.WriteFile(tokenFile, buf, 0600)
}

func ReadToken(client string) (*AccessTokenResponse, error) {
    tokenFile := TokenFile(client)
    if _, err := os.Stat(tokenFile); os.IsNotExist(err) {
        return nil, err
    }
    buf, err := ioutil.ReadFile(tokenFile)
    if (err != nil) {
        os.Remove(tokenFile)
        return nil, err
    }
    var tokenResponse AccessTokenResponse
    err = json.Unmarshal(buf, &tokenResponse)
    if (err != nil) {
        os.Remove(tokenFile)
        return nil, err
    }
    if (time.Now().Unix() < tokenResponse.ExpiresIn) {
        return &tokenResponse, nil
    }

    if (tokenResponse.RefreshToken == "") {
        os.Remove(tokenFile)
        return nil, errors.New("no refresh token")
    }

    form := ClientForm()
    form.Set("grant_type", "refresh_token")
    form.Set("refresh_token", tokenResponse.RefreshToken)

    res, err := Token().Request().Form(form).Post()
    if (err != nil || res.Status() != 200) {
        os.Remove(tokenFile)
        return nil, errors.New("Failed to refresh")
    }

    var responseJson AccessTokenResponse
    res.ReadJson(&responseJson)
    responseJson.ProcessTokenResponse(client)
    return &responseJson, nil

}

func init() {
	rootCmd.AddCommand(tokenCmd)
    tokenCmd.Flags().String("token-type", "access", "Token type to output.  access, id, or refresh are valid options.  access is default.")
    tokenCmd.Flags().BoolVarP(&loginFlags.force, "force", "f", false, "Forces relogin, existing session terminated.")
    tokenCmd.Flags().BoolVar(&loginFlags.browser, "browser", false, "Launch and login through a browser.")
    tokenCmd.Flags().BoolVar(&loginFlags.offline, "offline", false, "Request offline access.")
    tokenCmd.Flags().BoolVar(&loginFlags.fakeBrowser, "fake-browser", false, "Launch and login through a browser.")
    tokenCmd.Flags().MarkHidden("fake-browser")

}
