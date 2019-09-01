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
    "github.com/namsral/flag"
    "gopkg.in/src-d/go-git.v4/plumbing/transport"
    "cz0.cz/czoczo/dotman/views"
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



// Program starts here
// =================================================

func main() {

    // hello
    log.Println("Starting \"dot file manager\" aka:")

    // print logo (remove colors)
    reg, _ := regexp.Compile("\\\\e\\[[0-9](;[0-9]+)?m")
    //reg, _ = regexp.Compile("\\\\e\\[")
    log.Println(reg.ReplaceAllString(getLogo(),""))

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
    var foldersMap map[string]string


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

// URL Router, methods for handling endpoints
// =================================================

    // handle fixed routes first
    // handle file serving under 'directory' variable name
    http.Handle(folder+"/"+directory+"/", http.StripPrefix(folder+"/"+directory+"/", fs))
    log.Println("Serving files under: " + folder+"/"+directory+"/")

    // handle all other HTTP requests 
    http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {

        // on each request:
        // log
		log.Println(r.RemoteAddr + ": " + r.RequestURI)
        // strip backslash at the end of request
        requestPath := strings.TrimSuffix(r.URL.Path,"/")

        // retrive repository folders mapped with alphabet characters
        foldersMap = getFoldersMap(directory, alphabet)

        // handle main request, print main menu script
        if requestPath == folder {
            views.ServeMain(w, r, baseurl, secret, getLogo())
            return
        }

        // all other futher routes require secret 
        client_secret := r.Header.Get("secret")
        if client_secret != secret {
            log.Println("Wrong secret, or not given.")
		    fmt.Fprintf(w, "echo \"Wrong secret, or not given.\"")
            return
        }

        // handle install endpointm print install menu script
        if requestPath == folder + "/install" {
            views.ServeInstall(w, r, baseurl, client_secret, getLogo(), directory, alphabet, foldersMap)
            return
        }

        // handle synchronization endpoint - pull git repo
        if requestPath == folder + "/sync" {
            gitPull(directory)
            fmt.Fprintf(w, "echo \"Repo synced\"")
            return
        }

        // handle update script endpoint
        if requestPath == folder + "/update" {
            views.ServeUpdate(w, r, baseurl, client_secret, directory, foldersMap)
            return
        }

        // handle auto update enable endpoint
        if requestPath == folder + "/autoenable" {
            views.ServeSetAuto(w, r, baseurl, client_secret, true)
            return
        }

        // handle auto update disable endpoint
        if requestPath == folder + "/autodisable" {
            views.ServeSetAuto(w, r, baseurl, client_secret, false)
            return
        }

        // if none above catched, return 404
        w.WriteHeader(http.StatusNotFound)
        fmt.Fprintf(w, "echo \"404 - Not Found\"")
	})

	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}
