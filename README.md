Kcinit
======

This is a command line utility to perform login on a Keycloak realm through OpenID Connect.  This tool was implemented
to provide application developers a mechanism to obtain access tokens for their command line applications.  Logins
done through this tool are persisted so that they can live between command line invocations and even console restarts.
Applications can use this tool to provide login and SSO to other command line applications.
For example, let's say you have a command line util call 'oc' that needs an access token to invoke on its backend.  You could write a wrapper script
to obtain the token from kcinit and pass it to the oc command via a flag or environment variable

     # filename: oc
     #!/bin/sh
     token=$(kcinit token oc)
     /path/to/binary/oc --token=$token $@

`kcinit` would prompt you for login information and obtain a token for the `oc` client application registered in the Keycloak realm.

Setup
-----

You will first have to set up and register a master oauth client in your keycloak realm that will be used as the master login
session for your command line console.  You can name this client anything you want and it can be a public or confidential client.
This client must have token exchange permissions for each application that you want to do SSO with on the command line console.


Next, you have to configure kcinit for your user account. Run the following to set the tool:
    
     $ kcinit install
     
This will prompt you for information about the URL of the auth server, the keycloak realm, and the client you created.
This will store configuration information with `$HOME/.keycloak/kcinit`.  If you want to store your configuration someplace else,
set the `KC_CONFIG_PATH` environment variable before running `install`.

Usage
-----

After you have installed kcinit, you can login with this command
     
     $ kcinit login

This will store a token file under `$HOME/.keycloak/kcinit` for your master client.

Invoking the `kcinit token` command will output the access token receive from a login of the master client to stdout.  If you have not
logged in yet, you will be prompted to enter in your credentials.  You can specify `kcinit token [client]` to obtain a token from
another client application registered in the realm.  The master client must have token exchange permissions to to get this token.

`kcinit token` will use any existing token that you already have queried for as it stores these tokens on disk after retrieval.
The access token timeout is checked, and if it is expired, the tool will automatically refresh the token.

The output of `kcinit token` can be captured in an environment variable.  All interactive actions are all done on stderr.

To logout, just type `kcinit logout`.  This will remove any 
     
