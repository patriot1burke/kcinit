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
	"github.com/spf13/viper"
    "github.com/keycloak/kcinit/console"
    "os"
    "strings"
    "unicode"
    "net/url"
    "github.com/keycloak/kcinit/rest"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Interactive login",
	Long: `Interactive login to the currently configured realm.`,
	Run: login,
}

type AccessTokenResponse struct {
    AccessToken string `json:"access_token"`
    IdToken string `json:"id_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn int64 `json:"expires_in"`
    RefreshExpiresIn int64 `json:"refresh_expires_in"`
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().BoolP("force", "f", false, "Forces relogin terminating existing session")
}

type LoginParams struct {
	ResponseType string `url:"response_type,omitempty"`
	ClientId string `url:"client_id,omitempty"`
	RedirectUri string `url:"redirect_uri,omitempty"`
	Display string `url:"display,omitempty"`
	Scope string `url:"scope,omitempty"`
}

type param struct {
    name string
    label string
    mask bool
}

func login(cmd *cobra.Command, args []string) {
    CheckInstalled()
    forceLogin, _ := cmd.Flags().GetBool("force")
    if (!forceLogin) {
        token, err := ReadToken(viper.GetString("client"))
        if (token != nil && err == nil) {
            console.Writeln("Already logged in...")
            return
        }
    }
    DoLogin()
    console.Writeln()
    console.Writeln("Login successful!")
}

func DoLogin() *AccessTokenResponse {
    console.Traceln("login....")
    code := loginPrompt()
    console.Traceln("Got code!", code)
    form := ClientForm()
    form.Set("grant_type", "authorization_code")
    form.Set("code", code)
    form.Set("redirect_uri", "urn:ietf:wg:oauth:2.0:oob")
    console.Traceln("code2token params:", form)
    res, err := Token().Request().Form(form).Post()
    if (err != nil) {
        console.Writeln("Failure: failed to turn code into token")
        os.Exit(1)
    }
    if (res.Status() != 200) {
        if (res.MediaType() != "") {
            var json map[string]interface{}
            err := res.ReadJson(&json)
            if (err == nil) {
                console.Writeln("Failure: failed to turn code into token:", json["error"], json["error_description"])
                os.Exit(1)

            }
        }
        console.Writeln("Failure: failed to turn code into token")
        os.Exit(1)

    }
    var tokenResponse AccessTokenResponse
    res.ReadJson(&tokenResponse)
    tokenResponse.ProcessTokenResponse(viper.GetString("client"))
    return &tokenResponse
}


func loginPrompt() string {
    console.Traceln("invoke initial request")
	res, err := Authorization().
		QueryParam("response_type", "code").
        QueryParam("client_id", viper.GetString("client")).
        QueryParam("redirect_uri", "urn:ietf:wg:oauth:2.0:oob").
            QueryParam("scope", "openid").
                QueryParam("display", "console").
                    Request().Get()

    console.Traceln("Finished initial request")

    for {
        console.Traceln("Looping")
        if (err != nil) {
            console.Writeln("Failure: Could not invoke request", err)
            os.Exit(1)
        }

        if (res.Status() == 403) {
            console.Traceln("403")
            if (strings.EqualFold("text/plain", res.MediaType())) {
                text, _ := res.ReadText()
                console.Writeln(text)
            } else {
                console.Writeln("Failure: Forbidden to login")
            }
            os.Exit(1)
        } else if (res.Status() == 302) {
            console.Traceln("302")
            for i :=0 ; res.Status() == 302 && i < 5; i++ {
                console.Traceln("looping 302")
                location := res.Location()

                if (location == "") {
                    break
                }

                console.Traceln("Location:", location)

                url, _ := url.Parse(location)
                if (url == nil) {
                    console.Writeln("Failure: Could not parse redirect")
                    os.Exit(1)
                }
                q := url.Query()
                if (q.Get("code") != "") {
                    return q.Get("code")
                }

                res, err = rest.New().Target(location).Request().Get()
                if (err != nil) {
                    console.Writeln("Failure: Request execution failed")
                    os.Exit(1)
                }
            }
            if (res.Status() == 302) {
                console.Writeln("Failure:  Too many redirects")
                os.Exit(1)
            }

        } else if (res.Status() == 401) {
            console.Traceln("401")
            authenticationHeader := res.Header("WWW-Authenticate")
            if (authenticationHeader == "") {
                console.Writeln("Failure:  Invalid protocol.  No WWW-Authenticate header")
                os.Exit(1)
            }
            if (!strings.Contains(authenticationHeader, "X-Text-Form-Challenge")) {
                console.Writeln("Failure:  Invalid WWW-Authenticate header:", authenticationHeader)
                os.Exit(1)
            }
            if (res.MediaType() != "") {
                text, err := res.ReadText()
                if (err == nil) {
                    console.Writeln(text)

                }
            }
            lastQuote := rune(0)
            f := func(c rune) bool {
                switch {
                case c == lastQuote:
                    lastQuote = rune(0)
                    return false
                case lastQuote != rune(0):
                    return false
                case unicode.In(c, unicode.Quotation_Mark):
                    lastQuote = c
                    return false
                default:
                    return unicode.IsSpace(c)

                }
            }

            // splitting string by space but considering quoted section
            items := strings.FieldsFunc(authenticationHeader, f)
            console.Traceln("params len", len(items))

            var callback = ""
            var currparam *param
            var params = make([]*param, 0)
            for _, item := range items {
                console.Traceln("item:",item)
                var name,value string
                idx := strings.Index(item, "=")
                if (idx != -1) {
                    name = item[0:idx]
                    value= strings.Trim(item[idx + 1:], "\"")
                } else {
                    name = item
                    value = ""
                }

                if (name == "callback") {
                    callback = value
                } else if (name == "param"){
                    currparam = &param{}
                    params = append(params, currparam)
                    currparam.name = value
                    console.Traceln("Added param:", currparam.name)
                } else if (name == "label"){
                    currparam.label = value
                } else if (name == "mask") {
                    currparam.mask = value == "true"
                } else if (name == "browserRequired") {
                    if (res.MediaType() == "") {
                        console.Writeln("A browser is required to login.  Please login via --browser mode.")
                    }
                    os.Exit(1)
                }
            }

            console.Traceln("callback:",callback)

            form := url.Values{}
            for i, param := range params {
                console.Traceln("reading param:", i, param.name)
                val := ""
                if (param.mask) {
                    val = console.Password(param.label)
                } else {
                    val = console.ReadLine(param.label)
                }
                form.Add(param.name, val)
            }
            res, err = rest.New().Target(callback).Request().Form(form).Post()
        } else {
            if (strings.EqualFold("text/plain", res.MediaType())) {
                console.Writeln(res.ReadText())
            } else {
                console.Writeln("Unknown response from server:", res.Status())

            }
            os.Exit(1)
        }


    }


}
