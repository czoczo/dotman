
// Update view - updates already installed dotfiles
// =================================================

package views

import (
	"net/http"
    "text/template"
)

type AutoSetData struct {
    BaseURL string
    ClientSecret string
    Enable bool
}

func ServeSetAuto(w http.ResponseWriter, r *http.Request, baseurl string, client_secret string, enable bool) {

    // build data for template
    data := AutoSetData{baseurl, client_secret, enable}

    // render template
    tmpl, err := template.New("autoset").Parse(tmplAutoSet)
    if err != nil { panic(err) }
    err = tmpl.Execute(w, data)
    if err != nil { panic(err) }
}

var tmplAutoSet = `
exec 3<>/dev/tty
SECRET="{{ .ClientSecret }}"
SCRIPT_PATH="curl -s -H\"secret:$SECRET\" {{.BaseURL}}/update | bash -"

enableCron() {
      # check if crontab present
      command -v crontab >/dev/null 2>&1 || ( echo "  Error: couldn't find crontab. Auto update feature unsupported.")
    
      # prompt about adding updates to crontab
      echo -e "\n  This will add curl request \"{{.BaseURL}}/update | bash -\" to crontab. Proceed? [y/N]"
      read -u 3 -n 1 -r
      echo ""
      if [[ ! $REPLY =~ ^[Yy]$ ]]
      then
      echo "  Aborted."
      exit 0
      fi
     
      # Random point in a hour (0-59)
      MINS=$((RANDOM % 60))

      # Check if script is already in crontab, if not add itself
      CRON_ENTRY="$MINS * * * * $SCRIPT_PATH"
      if crontab -l 2>/dev/null | grep -q "$SCRIPT_PATH"; then
        echo "  Auto update already enabled"
      else
        ( crontab -l 2>/dev/null; echo "  $CRON_ENTRY" ) | crontab -
        echo -e "  Auto update enabled."
      fi 
      crontab -l 2>/dev/null | grep "$SCRIPT_PATH"
      exit 0
}

disableCron() {
      # prompt about adding updates to crontab
      echo -e "\n  This will delete curl auto update request from crontab. Proceed? [y/N]"
      read -u 3 -n 1 -r
      echo ""
      if [[ ! $REPLY =~ ^[Yy]$ ]]
      then
      echo "  Aborted."
      exit 0
      fi

      if crontab -l 2>/dev/null | grep -q "$SCRIPT_PATH"; then
          ( crontab -l 2>/dev/null | grep -v "$SCRIPT_PATH"; ) | crontab -
          echo -e "  Removed dotman auto update from crontab."
      else
        echo -e "  There is no dotman auto update entry in crontab. Aborting."
      fi
      exit 0
}
{{ if .Enable }}
enableCron
{{ else }}
disableCron
{{ end }}

`
