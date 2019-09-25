
// Main view with menu or secret prompt screen
// =================================================

package views

import (
    "strings"
	"net/http"
    "text/template"
)


func ServeMain(w http.ResponseWriter, r *http.Request, baseurl string, secret string, logo string) {
    // check for secret presence in HTTP header
    client_secret := r.Header.Get("secret")

    // build data for template
    data := passPromptData{baseurl, strings.ReplaceAll(logo,"'","'\"'\"'"), client_secret, ""}

    // set default template 
    tmplSrc := passPromptView

    // if correct secret, print menu (index)
    if client_secret == secret {
        tmplSrc = indexView
    }

    // render pass prompt template
    tmpl, err := template.New("main").Parse(tmplSrc)
    if err != nil { panic(err) }
    err = tmpl.Execute(w, data)
    if err != nil { panic(err) }
}


var indexView = bashTemplHead + `
#!/bin/bash
tput clear

SECRET="{{ .ClientSecret }}"
SCRIPT_PATH="curl -s -H\"secret:$SECRET\" {{.BaseURL}}/update | bash -"
crontab -l 2>/dev/null | grep -q "$SCRIPT_PATH" && AUTOUPDATESTATUS="Enabled" || AUTOUPDATESTATUS="Disabled"
[ -d "$HOME/.dotman/dotfiles" ] && INSTALLMETHOD="git & sym-links" || INSTALLMETHOD="file copies"

echo -e '{{.Logo}}'
barPrint
printf "%2s%s " "" "Info: "
printf " \e[35m%s\e[0m: \e[32m%s\e[0m" "managed items" "$(cat ~/.dotman/managed 2>/dev/null | wc -l)"
printf " \e[0m | "
printf " \e[35m%s\e[0m: \e[32m%s\e[0m\n" "auto update" "$AUTOUPDATESTATUS"
[ -f "$HOME/.dotman/managed" ] && printf "          \e[35m%s\e[0m: \e[32m%s\e[0m\n" "install method used:" "$INSTALLMETHOD"
echo ""
printf "%2s%s\n\n" "" "Select action:"
printf "  \e[32m%s\e[0m)\e[35m %-15s\e[0m\n" "i" "select and install dotfiles"
printf "  \e[32m%s\e[0m)\e[35m %-15s\e[0m\n" "l" "list installed dotfiles"
printf "  \e[32m%s\e[0m)\e[35m %-15s\e[0m\n" "t" "list available packages tags"
printf "  \e[32m%s\e[0m)\e[35m %-15s\e[0m\n" "u" "update installed dotfiles"
printf "  \e[32m%s\e[0m)\e[35m %-15s\e[0m\n" "c" "change dotfiles install method (git symlinks or file copies)"
printf "  \e[32m%s\e[0m)\e[35m %-15s\e[0m\n" "s" "make dotman pull changes from repository"
printf "  \e[32m%s\e[0m)\e[35m %-15s\e[0m\n" "e" "enable auto update dotfiles (requires cron)"
printf "  \e[32m%s\e[0m)\e[35m %-15s\e[0m\n" "d" "disable auto update dotfiles"
printf "  \e[32m%s\e[0m)\e[35m %-15s\e[0m\n" "q" "exit program"
barPrint
SECRET="{{.ClientSecret}}"


exec 3<>/dev/tty
echo ""
read -u 3 -n 1 -r -p "  Choose option: " opt
echo ""
case $opt in
i)
curl -s -H"secret:$SECRET" {{.BaseURL}}/install | bash -
;;
l)
if [  -f ~/.dotman/managed ]; then
echo -e "\n  Installed packages:\n"
  while read LINE; do
    echo -e "\e[35m  $LINE\e[0m"
  done < ~/.dotman/managed
else
  echo "\n  It appears no dotfiles are managed by dotman yet."
fi
;;
t)
curl -s -H"secret:$SECRET" {{.BaseURL}}/tagslist | bash -
;;
u)
curl -s -H"secret:$SECRET" {{.BaseURL}}/update | bash -
;;
c)
curl -s -H"secret:$SECRET" {{.BaseURL}}/changeInstallMethod | bash -
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
tput clear
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
