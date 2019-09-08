package views

type passPromptData struct {
    BaseURL string
    Logo string
    ClientSecret string
    EndPoint string
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
curl -s -H"secret:$SECRET" {{.BaseURL}}{{.EndPoint}} | bash -
`
