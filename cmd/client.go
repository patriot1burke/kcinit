package cmd



import (
	"github.com/spf13/viper"
	"github.com/keycloak/kcinit/rest"
    "github.com/keycloak/kcinit/console"
    "os"
    "net/url"
)


type KeycloakClient struct {
	client *rest.RestClient
}

var keycloak *rest.RestClient
var base *rest.WebTarget

func InitializeClient() {
    realmUrl := viper.GetString(REALM_URL)
    if (realmUrl == "") {
        answer := console.ReadLine("Realm URL not set, do you want to install a config file? [y/n]")
        if (answer == "n") {
            os.Exit(1)
        }
        runInstall()
        realmUrl = viper.GetString(REALM_URL)
    }
	keycloak = rest.New()
    base = keycloak.Target(realmUrl)
	if (base == nil) {
	    console.Writeln("Issues initializing client")
	    os.Exit(1)
    }
}

func ClientForm() (url.Values) {
    form := url.Values{}
    clientId := viper.GetString(LOGIN_CLIENT)
    form.Set("client_id", clientId)
    secret := viper.GetString(LOGIN_SECRET)
    if (secret != "") {
        form.Set("client_secret", secret)
    }
    return form
}


func Oidc() *rest.WebTarget {
    return  base.Path("protocol/openid-connect")
}

func Authorization() *rest.WebTarget {
	return Oidc().Path("auth")
}

func Token() *rest.WebTarget {
	return Oidc().Path("token")
}


func Logout() *rest.WebTarget {
    return Oidc().Path("logout")
}

func Userinfo() *rest.WebTarget {
    return Oidc().Path("userinfo")
}




