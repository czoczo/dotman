#!/bin/bash
tput clear
echo -e '{{.Logo}}'
echo -e "\e[97m-========================================================-\n\e[0;37m"
exec 3<>/dev/tty
printf "secret: "
read -u 3 -s SECRET
curl -s -H"secret:$SECRET" {{.BaseURL}} | bash -

