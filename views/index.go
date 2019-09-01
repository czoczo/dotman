
// Main view with menu or secret prompt screen
// =================================================

package views

import (
    "strings"
	"net/http"
    "text/template"
)

type IndexData struct {
    BaseURL string
    Logo string
    ClientSecret string
}

func ServeMain(w http.ResponseWriter, r *http.Request, baseurl string, secret string, logo string) {
    // check for secret presence in HTTP header
    client_secret := r.Header.Get("secret")

    // build data for template
    data := IndexData{baseurl, strings.ReplaceAll(logo,"'","'\"'\"'"), client_secret}

    // set default template 
    tmplSrc := passPromptView

    // if correct secret, print menu (index)
    if client_secret == secret {
        tmplSrc = indexView
    }

    // render template
    tmpl, err := template.New("main").Parse(tmplSrc)
    if err != nil { panic(err) }
    err = tmpl.Execute(w, data)
    if err != nil { panic(err) }
}

var passPromptView = `
#!/bin/bash
tput clear
echo -e '{{.Logo}}'
echo -e "\e[97m-========================================================-\n\e[0;37m"
exec 3<>/dev/tty
printf "secret: "
read -u 3 -s SECRET
curl -s -H"secret:$SECRET" {{.BaseURL}} | bash -
`

var indexView = `
#!/bin/bash
tput clear
echo -e '{{.Logo}}'
echo -e "\e[97m-========================================================-\n\e[0;37m"
printf "%2s%s\n\n" "" "Select action:"
printf "  \e[32m%s\e[0m)\e[35m %-15s\e[0m\n" "i" "install selected dotfiles"
printf "  \e[32m%s\e[0m)\e[35m %-15s\e[0m\n" "l" "list installed dotfiles"
printf "  \e[32m%s\e[0m)\e[35m %-15s\e[0m\n" "u" "update installed dotfiles"
printf "  \e[32m%s\e[0m)\e[35m %-15s\e[0m\n" "s" "make dotman pull changes from repository"
printf "  \e[32m%s\e[0m)\e[35m %-15s\e[0m\n\n" "q" "exit program"
echo -e "\e[97m-========================================================-\n\e[0m"
SECRET="{{.ClientSecret}}"

exec 3<>/dev/tty
echo ""
read -u 3 -p "  Chosen option: " opt
echo "$opt"
echo ""
case $opt in
i)
curl -s -H"secret:$SECRET" {{.BaseURL}}/install | bash -
;;
l)
[  -f ~/.dotman/managed ] && cat ~/.dotman/managed || echo "It appears no dotfiles are managed by dotman yet."
;;
u)
curl -s -H"secret:$SECRET" {{.BaseURL}}/update | bash -
;;
s)
curl -s -H"secret:$SECRET" {{.BaseURL}}/sync | bash -
;;
q)
echo "Quiting"; exit 0
;;
*)
echo "Invalid option, quiting"; exit 1
;;
esac
`
