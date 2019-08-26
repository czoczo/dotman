
// Main view with menu or secret prompt screen
// =================================================

package main

import (
	"fmt"
    "strings"
	"net/http"
)

func serveMain(w http.ResponseWriter, r *http.Request) {
            // start with shebang
            fmt.Fprint(w,`#!/bin/bash
tput clear
`)
            // print ASCII logo
            fmt.Fprint(w,"echo -e '"+strings.ReplaceAll(getLogo(),"'","'\"'\"'")+"'\n")
            fmt.Fprint(w,"echo -e \"\\e[97m-========================================================-\\n\\e[0;37m\"\n")

            // check for secret presence in HTTP header
            client_secret := r.Header.Get("secret")

            // if bad secret, print secret prompt
            if client_secret != secret {
                fmt.Fprint(w,`
exec 3<>/dev/tty
printf "secret: "
read -u 3 -s SECRET
curl -s -H"secret:$SECRET" ` + baseurl + " | bash -")
                return
            }

            // print error if more folders than available alphabet
            fmt.Fprint(w,"printf \"%2s%s\\n\\n\" \"\" \"Select action:\"\n")
            fmt.Fprint(w,"printf \"  \\e[32m%s\\e[0m)\\e[35m %-15s\\e[0m\\n\" \"i\" \"install selected dotfiles\" \n")
            fmt.Fprint(w,"printf \"  \\e[32m%s\\e[0m)\\e[35m %-15s\\e[0m\\n\" \"u\" \"update installed dotfiles\" \n")
            fmt.Fprint(w,"printf \"  \\e[32m%s\\e[0m)\\e[35m %-15s\\e[0m\\n\" \"s\" \"make dotman pull changes from repository\" \n")
            fmt.Fprint(w,"printf \"  \\e[32m%s\\e[0m)\\e[35m %-15s\\e[0m\\n\\n\" \"q\" \"exit program\" \n")
            fmt.Fprint(w,"echo -e \"\\e[97m-========================================================-\\n\\e[0m\"\n")
            fmt.Fprint(w,"SECRET=\""+client_secret+"\"\n")
            fmt.Fprint(w,`
exec 3<>/dev/tty
echo ""
read -u 3 -p "  Chosen option: " opt
echo "$opt"
echo ""
case $opt in
i)
    curl -s -H"secret:$SECRET" ` + baseurl + `/install | bash -
    ;;
u)
    curl -s -H"secret:$SECRET" ` + baseurl + `/update | bash -
    ;;
s)
    curl -s -H"secret:$SECRET" ` + baseurl + `/sync | bash -
    ;;
q)
    echo "Quiting"; exit 0
    ;;
*)
    echo "Invalid option, quiting"; exit 1
    ;;
esac
`)
            return
        }
