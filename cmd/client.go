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
	keycloak = rest.New()
    server := viper.GetString("server")
    base = keycloak.Target(server)
	if (base == nil) {
	    console.Writeln("Issues initializing client")
	    os.Exit(1)
    }
}

func ClientForm() (url.Values) {
    form := url.Values{}
    form.Set("client_id", viper.GetString("client"))
    secret := viper.GetString("clientSecret")
    if (secret != "") {
        form.Set("client_secret", secret)
    }
    return form
}


func Oidc() *rest.WebTarget {
	realm:= viper.GetString("realm")
    return  base.Path("realms").
		Path(realm).
			Path("protocol/openid-connect")
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


