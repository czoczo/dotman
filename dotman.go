package main

// go get -u gopkg.in/src-d/go-git.v4/
// go get github.com/gliderlabs/ssh
// go get github.com/namsral/flag

import (
	"fmt"
	"net/http"
    "strconv"
    "log"
    "regexp"
	"os"
    "strings"
    "io/ioutil"
    "crypto/rsa"
    "crypto/rand"
    "encoding/pem"
    "crypto/x509"
    "path/filepath"
    "golang.org/x/crypto/ssh"
    "golang.org/x/crypto/ssh/knownhosts"
    "github.com/namsral/flag"
    "gopkg.in/src-d/go-git.v4/plumbing/transport"
    gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	git "gopkg.in/src-d/go-git.v4"
    githttp "gopkg.in/src-d/go-git.v4/plumbing/transport/http"
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



// Utilities
// =================================================

// authorization object for git connection
var auth transport.AuthMethod

// CheckIfError should be used to naively panics if an error is not nil.
func CheckIfError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}

// Info should be used to describe the example commands that are about to run.
func Info(format string, args ...interface{}) {
	fmt.Printf("\x1b[35;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

// Warning should be used to display a warning
func Warning(format string, args ...interface{}) {
	fmt.Printf("\x1b[36;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

// print bash CASE operator body with available dotfiles folders as options.
// This CASE operator is used to eihter print options for menu, or output command for installing dotfiles
func repoOptsCasePrint(w http.ResponseWriter, foldersMap map[string]string, byName bool)  {

    // itertate over available option
    for key, val := range foldersMap {
            index := key

            // map either by alphabet char, or folder name
            if byName { index = val }

            // print bash condition. If $2 passed, print menu option and return
            fmt.Fprint(w,"        "+index+")\n            if [ $2 ]; then\n                printf \"\\e[0;32m*\\e[0m)\\e[0;35m %s\\e[0m\" \""+val+"\"\n                return\n            fi\n")

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
                    fmt.Fprint(w,"             #mkdir -p $HOME/" + filepath.Dir(output)+"\n")
                }

                // print download commands
                fmt.Fprint(w,"             #curl -H\"secret:$SECRET\"" + baseurl + "/" + path + "\" > $HOME/" + output+"\n")

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

// git pull method used to sync shared folder
func gitPull(directory string) {

    // we instance\iate a new repository targeting the given path (the .git folder)
    gitr, err := git.PlainOpen(directory)
    CheckIfError(err)

    // get the working directory for the repository
    gitw, err := gitr.Worktree()
    CheckIfError(err)

    // pull the latest changes from the origin remote and merge into the current branch
    Info("git pull origin")
    err = gitw.Pull(&git.PullOptions{
        Auth: auth,
        RemoteName: "origin",
    })

    // ... retrieving the branch being pointed by HEAD
    ref, err := gitr.Head()
    CheckIfError(err)
    // ... retrieving the commit object
    commit, err := gitr.CommitObject(ref.Hash())
    CheckIfError(err)

    fmt.Println(commit)
}

// MakeSSHKeyPair make a pair of public and private keys for SSH access.
// Public key is encoded in the format for inclusion in an OpenSSH authorized_keys file.
// Private Key generated is PEM encoded
func MakeSSHKeyPair(pubKeyPath, privateKeyPath string) error {
    privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
    if err != nil {
        return err
    }

    // generate and write private key as PEM
    privateKeyFile, err := os.Create(privateKeyPath)
    defer privateKeyFile.Close()
    if err != nil {
        return err
    }
    privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
    if err := pem.Encode(privateKeyFile, privateKeyPEM); err != nil {
        return err
    }

    // generate and write public key
    pub, err := ssh.NewPublicKey(&privateKey.PublicKey)
    if err != nil {
        return err
    }
    return ioutil.WriteFile(pubKeyPath, ssh.MarshalAuthorizedKey(pub), 0655)
}

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filename string) bool {
    info, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return !info.IsDir()
}

// Functions for serving cloned repository files on HTTP
// =================================================

// blacklisting some files from being served

// check if path starts with dot or a README.md
func containsDotFile(name string) bool {
        if strings.HasPrefix(name, "/.") || name=="/README.md" {
            return true
        }
    return false
}

// fileHidingFile is the http.File use in fileHidingFileSystem.
// it is used to wrap the Readdir method of http.File so that we can
// remove files and directories that are blacklisted from its output.
type fileHidingFile struct {
    http.File
}

// readdir is a wrapper around the Readdir method of the embedded File
// that filters out all files that start with a period in their name.
func (f fileHidingFile) Readdir(n int) (fis []os.FileInfo, err error) {
    files, err := f.File.Readdir(n)
    for _, file := range files { // filters out the files
        if !strings.HasPrefix(file.Name(), ".") && file.Name()!="README.md" {
            fis = append(fis, file)
        }
    }
    return
}

// fileHidingFileSystem is an http.FileSystem that hides
// hidden "dot files" from being served.
type fileHidingFileSystem struct {
    http.FileSystem
}

// open is a wrapper around the Open method of the embedded FileSystem
// that serves a 403 permission error when name has a file or directory
// with whose name starts with a period in its path.
func (fs fileHidingFileSystem) Open(name string) (http.File, error) {
    if containsDotFile(name) { // If dot file, return 403 response
        return nil, os.ErrPermission
    }

    file, err := fs.FileSystem.Open(name)
    if err != nil {
        return nil, err
    }
    return fileHidingFile{file}, err
}

// Check if secret passed to protect shared dotfiles
func checkSecretThenServe(h http.Handler) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        client_secret := r.Header.Get("secret")
        if client_secret != secret {
            fmt.Fprint(w,"Ups! No secret given.\n")
            return
        }
        // Serve with the actual handler.
        h.ServeHTTP(w, r)
    }
}




// Program starts here
// =================================================

func main() {

    // define var for program ascii logo
    logo := `
\e[2;35m                $$.  ,$$    $$    $$,   $$
                $$$,.$$$  ,$/\$.  $$\.  $$
\e[0;35m          .$$$. \e[2;35m$$'$$'$$ ,$$__$$. $$'$$.$$
\e[0;35m          $$$$$ \e[2;35m$$    $$ $$&&&&$$ $$  '\$$
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
    flag.StringVar(&url, "url", "", "git repository to connect to")
    flag.StringVar(&password, "password", "", "used to connect to git repository, when using http protocol")
    flag.StringVar(&directory, "directory", "dotfiles", "endpoint under which to serve files.")
    flag.StringVar(&secret, "secret", "", "used to protect files served by server")
    flag.StringVar(&sshkey, "sshkey", "ssh_data/id_rsa", "path to key used to connect git repository when using ssh protocol.")
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
    re := regexp.MustCompile("(ssh|https?)://(.+)@.+")
    if re.MatchString(url) == false {
        flag.PrintDefaults()
        fmt.Println()
        log.Println("Provided repository URL: " + url + " not supported. Provide either ssh or http(s) protocol URL with username. Exiting.")
        os.Exit(1)
    }

    // extract username
    unamematch := re.FindStringSubmatch(url)
    username = unamematch[2]

    // check baseurl
    match, _ := regexp.MatchString("https?://.+",baseurl)
    if match == false {
        log.Println("Unsupported base URL given. Use http or https protocol based URL. Exiting.")
        os.Exit(2)
    }

    // server config
    alphabet = "01234567890abcdefghijklmnopqrstuwxyzABCDEFGHIJKLMNOPQRSTUWXYZ"
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

    // define ssh or http connection string

    // switch available protocols, and generate auth object
    if strings.HasPrefix(url, "http") {
        // create auth object for git  connection
        auth = &githttp.BasicAuth {
                Username: username,
                Password: password,
        }
    }
    if strings.HasPrefix(url, "ssh") {
        // read private key
        sshKey, _ := ioutil.ReadFile(sshkey)
        signer, _ := ssh.ParsePrivateKey([]byte(sshKey))

        // create known_hosts file
        hostKeyCallback, err := knownhosts.New(ssh_known_hosts)
        if err != nil {
            log.Fatal(err)
        }
        hostKeyCallbackHelper := gitssh.HostKeyCallbackHelper{HostKeyCallback: hostKeyCallback}
//        hostKetCallbakHelper.HostKeyCallback = hostKeyCallback

        // create auth object for git  connection
        auth = &gitssh.PublicKeys{User: "git", Signer: signer, HostKeyCallbackHelper: hostKeyCallbackHelper}
    }

    // clone repo locally if not already here
    if _, err := os.Stat(directory); os.IsNotExist(err) {
        // Clone the given repository to the given directory
        Info("git clone %s %s", url, directory)

        gitr, err := git.PlainClone(directory, false, &git.CloneOptions{
            Auth: auth,
            URL:      url,
            Progress: os.Stdout,
        })

        CheckIfError(err)

        // ... retrieving the branch being pointed by HEAD
        ref, err := gitr.Head()
        CheckIfError(err)
        // ... retrieving the commit object
        commit, err := gitr.CommitObject(ref.Hash())
        CheckIfError(err)

        fmt.Println(commit)

    } else {
        gitPull(directory)
    }

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
            fmt.Fprint(w,`
#!/bin/sh)
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
curl -s -H"secret:$SECRET" ` + baseurl + " | sh -")
                return
            }

            // print error if more folders than available alphabet
            if len(files) > len(alphabet) {
                fmt.Fprint(w,"printf \"%2s%s\\n%4s%s\\n%4s%s\\n%4s%s\\n\\n%2s%s\\n\\n\" \"\" \"Congratz - you reached the limit of number of supported folders. Either:\" \"\" \"a) Wait for me to have the same problem someday.\" \"\" \"b) Increase number of unique characters in 'alphabet' variable.\" \"\" \"c) Implement other solution yourself.\" \"\" \"Decide which is the fastes option on your own ;)\"\n")
            }

            // print menu
            fmt.Fprint(w,"printf \"%2s%s\\n%2s%s\\e[32m%s\\e[0;37m%s\\n\\n\" \"\" \"Choose dotfiles to be installed.\" \"\" \"Select by typing keys (\" \"green\" \") and confirm with enter.\"\n")
            fmt.Fprint(w,"echo -e \"\\e[97m-========================================================-\\n\\e[0m\"\n")

            // vars for iterating available options
            count := 0
            nl := ""

            // iterate over options
            for char := range alphabet {

                    // exit if no keys in map anymore
                    key := string(char)
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

        // all other following routes require secret 
        client_secret := r.Header.Get("secret")
        if client_secret != secret {
                    log.Println("client secret "+client_secret+ " != " + secret + " header secret")
		    fmt.Fprintf(w, "echo \"Secret not given.\"")
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
            fmt.Fprint(w,"tput clear\n")
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
