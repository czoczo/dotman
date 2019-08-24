package main

// go get -u gopkg.in/src-d/go-git.v4/
// go get github.com/gliderlabs/ssh
// go get github.com/namsral/flag

import (
	"fmt"
    "strconv"
    "log"
    "regexp"
	"os"
	"net/http"
    "strings"
    "io/ioutil"
    "path/filepath"
    "github.com/namsral/flag"
    "gopkg.in/src-d/go-git.v4/plumbing/transport"
)


// Global variables
// =================================================

// define ssh or http connection string
var url string
var username string
var password string
var directory string
var secret string
var port int

// server config
// URL (e.g. https://exmaple.org:1338/dotfiles) under to create links to resources
var baseurl string

// set of characters to assign options to
var alphabet string
var ssh_known_hosts string

// authorization object for git connection
var auth transport.AuthMethod




// print bash CASE operator body with available dotfiles folders as options.
// This CASE operator is used to eihter print options for menu, or output command for installing dotfiles
func repoOptsCasePrint(w http.ResponseWriter, foldersMap map[string]string, byName bool)  {

    // itertate over available option
    for key, val := range foldersMap {
            index := key

            // map either by alphabet char, or folder name
            if byName { index = val }

            // print bash condition. If $2 passed, print menu option and return
            fmt.Fprint(w,"        "+index+")\n            if [ \"$2\" ]; then\n                printf \"\\e[0;32m*\\e[0m)\\e[0;35m %s\\e[0m\" \""+val+"\"\n                return\n            fi\n")

            // else print download commands
            fmt.Fprint(w,"             echo -e \"Installing \\e[0;35m"+val+"\\e[0m\"\n")

            // search folders and add mkdir and download commands
            // recursive walk thorough the dir
            err := filepath.Walk(directory+"/"+val,
                func(path string, info os.FileInfo, err error) error {
                if err != nil {
                    return err
                }

                // skip if not a dir
                if info.IsDir() {
                    return nil
                }

                // if not, strip absolute path and filename. Print mkdir commands
                output := strings.TrimPrefix(path, directory+"/"+val+"/")
                if filepath.Dir(output) != "." {
                    fmt.Fprint(w,"             mkdir -p $HOME/" + filepath.Dir(output)+"\n")
                }

                // print download commands
                fmt.Fprint(w,"             echo -n \"downloading file - "+output+" : \"\n")
                fmt.Fprint(w,"             curl -sH\"secret:$SECRET\" \"" + baseurl + "/" + path + "\" > \"$HOME/" + output+"\"\n")
                fmt.Fprint(w,"             RESULT=$?; [ $RESULT -eq 0 ] && echo -e \"\\e[0;32mok\\e[0m\" || echo -e \"\\e[0;31merror\\e[0m\"\n")

                // if not present, add option to managed dotfiles list
                fmt.Fprint(w,"             cat \"$HOME/.dotman/managed\" | grep -q \""+val+"\" || echo \""+val+"\" >> \"$HOME/.dotman/managed\" \n")
                return nil
            })

            if err != nil {
                log.Println(err)
            }
            fmt.Fprint(w,"             ;;\n")
    }
}




// Program starts here
// =================================================

func main() {

    // define var for program ascii logo
    logo := `
\e[2;35m                $$.  ,$$    $$    $$,   $$
                $$$,.$$$  ,$$$$.  $$$.  $$
\e[0;35m          .$$$. \e[2;35m$$'$$'$$ ,$$.,$$. $$'$$.$$
\e[0;35m          $$$$$ \e[2;35m$$    $$ $$$$$$$$ $$  '$$$
\e[0;35m          '$$$' \e[2;35m$$    $$ $$    $$ $$   '$$\e[0m
`
    // hello
    log.Println("Starting \"dot file manager\" aka:")

    // print logo (remove colors)
    reg, _ := regexp.Compile("\\\\e\\[[0-9](;[0-9]+)?m")
    //reg, _ = regexp.Compile("\\\\e\\[")
    log.Println(reg.ReplaceAllString(logo,""))

    // parse arguments/environment configuration
    var sshkey string
    var sshAccept bool
    flag.StringVar(&url, "url", "", "git repository to connect to")
    flag.StringVar(&password, "password", "", "used to connect to git repository, when using http protocol")
    flag.StringVar(&directory, "directory", "dotfiles", "endpoint under which to serve files.")
    flag.StringVar(&secret, "secret", "", "used to protect files served by server")
    flag.StringVar(&sshkey, "sshkey", "ssh_data/id_rsa", "path to key used to connect git repository when using ssh protocol.")
    flag.BoolVar(&sshAccept, "sshaccept", false, "whether to add ssh remote servers key to known hosts file")
    flag.IntVar(&port, "port", 1338, "servers listening port")
    flag.StringVar(&baseurl, "baseurl", "http://127.0.0.1:1338", "URL for generating download commands.")
	flag.Parse()

    // validate configuration
    isSsh, _ := regexp.Compile("ssh://")

    // check url
    // if url variable not set
    if url == "" {
        log.Println("For server to start you must provide git repository URL containing your dotfiles in folders.")
        log.Println("Use -url switch or URL environment variable.")
        os.Exit(1)
    }

    // if url not valid
    re := regexp.MustCompile("(ssh|https?)://(.+)@([^/$]+)")
    if re.MatchString(url) == false {
        flag.PrintDefaults()
        fmt.Println()
        log.Println("Provided repository URL: " + url + " not supported. Provide either ssh or http(s) protocol URL with username. Exiting.")
        os.Exit(1)
    }

    // extract username
    urlMatch := re.FindStringSubmatch(url)
    username = urlMatch[2]

    // extracte remote host address with port
    remoteHost := urlMatch[3]
    if ! strings.Contains(remoteHost, ":") {
        remoteHost = remoteHost + ":22"
    }

    // check baseurl
    match, _ := regexp.MatchString("https?://.+",baseurl)
    if match == false {
        log.Println("Unsupported base URL given. Use http or https protocol based URL. Exiting.")
        os.Exit(2)
    }

    // server config
    alphabet = "0123456789abcdefghijklmnopqrstuwxyzABCDEFGHIJKLMNOPQRSTUWXYZ"
    ssh_known_hosts = "ssh_data/known_hosts"

    // available options list
    foldersMap := make(map[string]string)


    // check if ssh protocol
    if isSsh.MatchString(url) {
        // create ssh_data dir
        os.MkdirAll("ssh_data", os.ModePerm)

        if !fileExists(sshkey) {
            log.Println("SSH Key " + sshkey + " not found. Falling back to generating key pair")
            err := MakeSSHKeyPair("ssh_data/id_rsa.pub","ssh_data/id_rsa")
            CheckIfError(err)
            log.Println("SSH Key pair generated successfully")
        }

        if !fileExists(ssh_known_hosts) {
            log.Println("SSH " + ssh_known_hosts + " file not found.")
            emptyFile, err := os.Create(ssh_known_hosts)
            CheckIfError(err)
            emptyFile.Close()
            log.Println("SSH known_hosts file successfully")
        }
    }

    // print hello and server configuration
    log.Println("Starting dotman - dot file manager.")
    log.Println("Repository URL: " + url)
    log.Println("GIT username: " + username)
    log.Println("Listening port: " + strconv.Itoa(port))
    log.Println("Download URLs prefix: " + baseurl+"/"+directory)

    // if using generated key pair print public key
    if sshkey == "ssh_data/id_rsa" && fileExists("ssh_data/id_rsa.pub") {
        log.Println("Using generate ssh key pair. Public key is:\n")
        pubkey, err := ioutil.ReadFile("ssh_data/id_rsa.pub")
        CheckIfError(err)
        fmt.Println(string(pubkey))
    }


    // sync file server directory with remote git repository

    // obatin auth object
    auth = getAuth(url, sshkey, remoteHost, sshAccept)

    // do actual sync
    gitSync(auth, url, directory)

    // serve locally cloned repo with dotfiles through HTTP
    // secured with secret and file blacklisting
    fs := checkSecretThenServe(http.FileServer(fileHidingFileSystem{http.Dir(directory)}))

    // handle being served as a subfolder 
    basere := regexp.MustCompile("^https?://[^/]+(.+)?")
    basematch := basere.FindStringSubmatch(baseurl)
    folder := strings.TrimSuffix(basematch[1],"/")

// URL Router, methods handling different endpoints
// =================================================

    // handle fixed routes first
    // handle file serving another 'directory' variable name
    http.Handle(folder+"/"+directory+"/", http.StripPrefix(folder+"/"+directory+"/", fs))
    log.Println("Serving files under: " + folder+"/"+directory+"/")

    // handle all other HTTP requests with dynamic URLs and headers
    http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {

        // on each request:
        // strip backslash at the end of request
        requestPath := strings.TrimSuffix(r.URL.Path,"/")
        // log
		log.Println(requestPath)

        // regex for options as URL, not used as for now
        //commaListRegex, _ := regexp.Compile("^/([0-9a-zA-Z]+,?)+$")
        //listRegex, _ := regexp.Compile("^/[0-9a-zA-Z]+$")


        // get folder list made of 1st level of folders in reposiotory contining dotfiles
        files, err := ioutil.ReadDir(directory)
        if err != nil {
            log.Fatal(err)
        }

            // iterate over above list, and assign characters from alphabet to folders in map
        charCounter := 0
        for _, f := range files {

                // skip .git and README.md - case insensitive
                match, _ := regexp.MatchString("(?i)(.git|README.md)", f.Name())
                if match {
                    continue
                }

                // print next alphabet character and mapped folder name
                foldersMap[string(alphabet[charCounter])] = f.Name()
                charCounter = charCounter+1

        }

        // handle main request, print main menu script
        if requestPath == folder {

            // start with shebang
            fmt.Fprint(w,`#!/bin/bash
tput clear
`)
            // print ASCII logo
            fmt.Fprint(w,"echo -e '"+strings.ReplaceAll(logo,"'","'\"'\"'")+"'\n")
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
            if len(files) > len(alphabet) {
                fmt.Fprint(w,"printf \"%2s%s\\n%4s%s\\n%4s%s\\n%4s%s\\n\\n%2s%s\\n\\n\" \"\" \"Congratz - you reached the limit of number of supported folders. Either:\" \"\" \"a) Wait for me to have the same problem someday.\" \"\" \"b) Increase number of unique characters in 'alphabet' variable.\" \"\" \"c) Implement other solution yourself.\" \"\" \"Decide which is the fastes option on your own ;)\"\n")
            }
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


        // all other following routes require secret 
        client_secret := r.Header.Get("secret")
        if client_secret != secret {
                    log.Println("client secret "+client_secret+ " != " + secret + " header secret")
		    fmt.Fprintf(w, "echo \"Secret not given.\"")
            return
        }

        // handle install endpointm print install menu script
        if requestPath == folder + "/install" {

            // start with shebang
            fmt.Fprint(w,`#!/bin/bash
tput clear
`)
            // print ASCII logo
            fmt.Fprint(w,"echo -e '"+strings.ReplaceAll(logo,"'","'\"'\"'")+"'\n")
            fmt.Fprint(w,"echo -e \"\\e[97m-========================================================-\\n\\e[0;37m\"\n")

            // print menu
            fmt.Fprint(w,"printf \"%2s%s\\n%2s%s\\e[32m%s\\e[0;37m%s\\n\\n\" \"\" \"Choose dotfiles to be installed.\" \"\" \"Select by typing keys (\" \"green\" \") and confirm with enter.\"\n")
            fmt.Fprint(w,"echo -e \"\\e[97m-========================================================-\\n\\e[0m\"\n")

            // vars for iterating available options
            count := 0
            nl := ""

            // iterate over options
            for i := 0; i < len(alphabet); i++ {

                    // exit if no keys in map anymore
                    key := string(alphabet[i])
                    if _, ok := foldersMap[key]; !ok {
                        continue
                    }

                    // handle deviding menu in 3 columns
                    count += 1
                    if count % 3 == 0 {
                        nl = "\\n"
                    }

                    // print menu option
                    fmt.Fprint(w,"printf \"  \\e[32m%s\\e[0m)\\e[35m %-15s\\e[0m"+nl+"\" \""+key+"\" \""+foldersMap[key]+"\" \n")
                    nl = ""
            }

            // touch list of dotfiles to be update
            fmt.Fprint(w,"mkdir -p \"$HOME/.dotman\"; touch \"$HOME/.dotman/managed\"\n")

            // print case function
            fmt.Fprint(w,"SECRET=\""+client_secret+"\"\n")
            fmt.Fprint(w,"selectOption() {\n    case \"$1\" in\n")
            repoOptsCasePrint(w, foldersMap, false)

            // close case 
            fmt.Fprint(w,"    esac\n}\n")

            // print variable with available options for cross checking input
            keys := make([]string, 0)
            for key := range foldersMap {
                keys = append(keys, key)
            }
            fmt.Fprint(w,"OPTS=\""+strings.Join(keys, "")+"\"\n")

            // print rest of script
            fmt.Fprint(w,`
exec 3<>/dev/tty
echo ""
read -u 3 -p "  Chosen options: " words
echo ""
if [ -z $words ]; then
    echo -e "\e[0;37mNothing to do... exiting."
    exit 0
fi
echo -e "\e[97m-========================================================-\e[0;37m"
printf "%2s\n" "" "Follwing dotfiles will be installed in order:"
COMMA=""
for CHAR in $(echo "$words" | fold -w1); do
    test "${OPTS#*$CHAR}" != "$OPTS" || continue
    echo -en "$COMMA" 
    selectOption $CHAR False
    COMMA=", "
done

if [ "$COMMA" == "" ]; then
    echo "Nothing to do... exiting."
    exit 0
fi

printf "\n%2s" "" "Proceed? [y/N]"
read -u 3 -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]
then
    exit 0
fi

echo -e "\\n\e[97m-========================================================-\e[0m\n"
echo "Installing dotfiles:"

for CHAR in $(echo "$words" | fold -w1); do
    test "${OPTS#*$CHAR}" != "$OPTS" || continue
    selectOption $CHAR 
done
            `)

            return
        }

//        if commaListRegex.MatchString(requestPath) {
//            slice := strings.Split(strings.Replace(requestPath,"/","",-1), ",")
//            var response string
//            for _, num := range slice {
//                response = response +" | "+ num
//            }
//		    fmt.Fprintf(w, response)
//            return
//        }

//        // if URI with chosen options print download script
//        if listRegex.MatchString(requestPath) {
//            choice := strings.Replace(requestPath,"/","",-1)
//            var response string
//            for _, char :=  range choice {
//                response = response +" | "+ string(char)
//            }
//		    fmt.Fprintf(w, response)
//            return
//        }

        // handle synchronization endpoint - pull git repo
        if requestPath == folder + "/sync" {
            gitPull(directory)
            fmt.Fprintf(w, "echo \"Repo synced\"")
            return
        }

        // handle update script endpoint
        if requestPath == folder + "/update" {

            // print case function
//            fmt.Fprint(w,"tput clear\n")
            fmt.Fprint(w,"SECRET=\""+client_secret+"\"\n")
            fmt.Fprint(w,"selectOption() {\n    case \"$1\" in\n")
            repoOptsCasePrint(w, foldersMap, true)

            // close case 
            fmt.Fprint(w,"    esac\n}\n")
            fmt.Fprint(w,`
if [ ! -f "$HOME/.dotman/managed" ]; then 
    echo "It appears, you don't manage any dotfiles using dotman. Exiting."
    exit 1
fi

for NAME in $(cat "$HOME/.dotman/managed"); do
    selectOption $NAME 
done
            `)
            return
        }

        // if none above catched, return 404
        w.WriteHeader(http.StatusNotFound)
        fmt.Fprintf(w, "echo \"404 - Not Found\"")
	})

	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}
