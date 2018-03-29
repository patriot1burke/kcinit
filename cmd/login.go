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
    "net"
    "fmt"
    "net/http"
    "runtime"
    "os/exec"
    "time"
    "golang.org/x/net/context"
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
    loginCmd.Flags().BoolP("force", "f", false, "Forces relogin, existing session terminated.")
    loginCmd.Flags().Bool("browser", false, "Launch and login through a browser.")
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

    browser, _ := cmd.Flags().GetBool("browser")
    if (browser) {
        Browser()
    } else {
        DoLogin()
    }

    console.Writeln()
    console.Writeln("Login successful!")
}

func DoLogin() *AccessTokenResponse {
    console.Traceln("login....")
    code, redirect := loginPrompt()
    console.Traceln("Got code!", code)
    return codeToToken(code, redirect)
}

func codeToToken(code string, redirect string) *AccessTokenResponse {
    form := ClientForm()
    form.Set("grant_type", "authorization_code")
    form.Set("code", code)
    form.Set("redirect_uri", redirect)
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


func loginPrompt() (string, string) {
    freePort, err := GetFreePort()
    var redirect string
    if (err != nil) {
        freePort = -1
        redirect = "http://localhost:666"
    } else {
        redirect = fmt.Sprintf("http://localhost:%d", freePort)
    }
    console.Traceln("invoke initial request")
	res, err := Authorization().
		QueryParam("response_type", "code").
        QueryParam("client_id", viper.GetString("client")).
        QueryParam("redirect_uri", redirect).
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
                    return q.Get("code"), redirect
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
            var wasOutput bool
            if (res.MediaType() != "") {
                text, err := res.ReadText()
                if (err == nil) {
                    wasOutput = true
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
            var promptBrowser bool
            var browserPrompt string
            var browserPromptAnswer string
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
                } else if (name == "browserContinue") {
                    if (freePort == -1) {
                        if (!wasOutput) {
                            console.Writeln("A browser is required to login.  Please login via --browser mode.")
                        }
                        os.Exit(1)
                    }
                    promptBrowser = true
                    browserPrompt = value
                } else if (name == "answer") {
                    browserPromptAnswer = value
                }
            }

            if (promptBrowser) {
                answer := console.ReadLine(browserPrompt)
                if (answer == browserPromptAnswer) {
                    listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", freePort))
                    if err != nil {
                        console.Writeln("Cannot start local http server to handle login redirect.")
                        os.Exit(1)
                    }
                    return launch(callback, listener), redirect
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

func Browser() *AccessTokenResponse {
    listener, err := net.Listen("tcp", "localhost:")
    if err != nil {
        console.Writeln("Cannot start local http server to handle login redirect.")
        os.Exit(1)
    }
    port := listener.Addr().(*net.TCPAddr).Port

    redirect := fmt.Sprintf("http://localhost:%d", port)
    url := Authorization().
        QueryParam("response_type", "code").
        QueryParam("client_id", viper.GetString("client")).
        QueryParam("redirect_uri", redirect).
        QueryParam("scope", "openid").Url()

    code := launch(url.String(), listener)
    if (code != "") {
        return codeToToken(code, redirect)
    } else {
        console.Writeln("Login failed")
        os.Exit(1)
    }
    return nil
}

func openBrowser(url string) bool {
    var args []string
    switch runtime.GOOS {
    case "darwin":
        args = []string{"open"}
    case "windows":
        args = []string{"cmd", "/c", "start"}
    default:
        args = []string{"xdg-open"}
    }
    cmd := exec.Command(args[0], append(args[1:], url)...)
    return cmd.Start() == nil
}

func launch(url string, listener net.Listener) string {
    c := make(chan string)

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        url := r.URL
        q := url.Query()
        w.Header().Set("Content-Type", "text/html")

        if (q.Get("code") != "") {
            fmt.Fprintln(w, "<html><h1>Login completed.</h1><div>");
            fmt.Fprintln(w, "This browser will remain logged in until you close it, logout, or the session expires.");
            fmt.Fprintln(w, "</div></html>");
            c <- q.Get("code")
        } else {
            fmt.Fprintln(w,"<html><h1>Login attempt failed.</h1><div>");
            fmt.Fprintln(w,"</div></html>");

            c <- ""
        }

    })

    srv := &http.Server{}
    ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
    defer srv.Shutdown(ctx)



    go func() {
        if err := srv.Serve(listener); err != nil {
            // cannot panic, because this probably is an intentional close
        }
    }()

    var code string
    if (openBrowser(url)) {
        code = <-c
    }

    return code

}

func GetFreePort() (int, error) {
    addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
    if err != nil {
        return 0, err
    }

    l, err := net.ListenTCP("tcp", addr)
    if err != nil {
        return 0, err
    }
    defer l.Close()
    return l.Addr().(*net.TCPAddr).Port, nil
}
