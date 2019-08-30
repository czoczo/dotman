#!/bin/bash
tput clear
echo -e '{{.Logo}}'
echo -e "\e[97m-========================================================-\n\e[0;37m"
printf "%2s%s\n\n" "" "Select action:"
printf "  \e[32m%s\e[0m)\e[35m %-15s\e[0m\n" "i" "install selected dotfiles"
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
