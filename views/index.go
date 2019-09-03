
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
case '{{.BaseURL}}' in
  "http://"*) echo -e "Plain HTTP used! Please configure TLS before using this script.\n"
esac
exec 3<>/dev/tty
printf "secret: "
read -u 3 -s SECRET
curl -s -H"secret:$SECRET" {{.BaseURL}} | bash -
`

var indexView = `
#!/bin/bash
tput clear

SECRET="{{ .ClientSecret }}"
SCRIPT_PATH="curl -s -H\"secret:$SECRET\" {{.BaseURL}}/update | bash -"
crontab -l 2>/dev/null | grep -q "$SCRIPT_PATH" && AUTOUPDATESTATUS="Enabled" || AUTOUPDATESTATUS="Disabled"

echo -e '{{.Logo}}'
echo -e "\e[97m-========================================================-\n"
printf "\e[0;37m%2s%s " "" "Info: "
printf " \e[35m%s\e[0m: \e[32m%s\e[0m" "managed items" "$(cat ~/.dotman/managed 2>/dev/null | wc -l)"
printf " \e[0m | "
printf " \e[35m%s\e[0m: \e[32m%s\e[0m\n\n" "auto update" "$AUTOUPDATESTATUS"
printf "\e[0;37m%2s%s\n\n" "" "Select action:"
printf "  \e[32m%s\e[0m)\e[35m %-15s\e[0m\n" "i" "select and install dotfiles"
printf "  \e[32m%s\e[0m)\e[35m %-15s\e[0m\n" "l" "list installed dotfiles"
printf "  \e[32m%s\e[0m)\e[35m %-15s\e[0m\n" "u" "update installed dotfiles"
printf "  \e[32m%s\e[0m)\e[35m %-15s\e[0m\n" "s" "make dotman pull changes from repository"
printf "  \e[32m%s\e[0m)\e[35m %-15s\e[0m\n" "e" "enable auto update dotfiles (requires cron)"
printf "  \e[32m%s\e[0m)\e[35m %-15s\e[0m\n" "d" "disable auto update dotfiles"
printf "  \e[32m%s\e[0m)\e[35m %-15s\e[0m\n\n" "q" "exit program"
echo -e "\e[97m-========================================================-\n\e[0m"
SECRET="{{.ClientSecret}}"


exec 3<>/dev/tty
echo ""
read -u 3 -n 1 -r -p "  Chosen option: " opt
echo ""
case $opt in
i)
curl -s -H"secret:$SECRET" {{.BaseURL}}/install | bash -
;;
l)
if [  -f ~/.dotman/managed ]; then
  echo -e "\e[35m"
  cat ~/.dotman/managed
  echo -e "\e[0m"
else
  echo "\n  It appears no dotfiles are managed by dotman yet."
fi
;;
u)
curl -s -H"secret:$SECRET" {{.BaseURL}}/update | bash -
;;
s)
curl -s -H"secret:$SECRET" {{.BaseURL}}/sync | bash -
;;
e)
curl -s -H"secret:$SECRET" {{.BaseURL}}/autoenable | bash -
;;
d)
curl -s -H"secret:$SECRET" {{.BaseURL}}/autodisable | bash -
;;
q)
echo -e "\n  Quiting"; exit 0
;;
*)
echo -e "\n  Invalid option. Try again."
;;
esac
echo ''
read -u 3 -n 1 -s -r -p "  Press any key to continue"
curl -s -H"secret:$SECRET" {{.BaseURL}} | bash -
`
