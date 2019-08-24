// SSH Protocol key handling
// =================================================

package main

import (
    "encoding/base64"
    "net"
    "crypto/rsa"
    "crypto/rand"
    "encoding/pem"
    "crypto/x509"
    "golang.org/x/crypto/ssh"
    "io/ioutil"
	"os"
    "log"
	"fmt"
    "strings"
)

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

// print server key
func KeyPrint(dialAddr string, addr net.Addr, key ssh.PublicKey) error {
    log.Printf("Remote server key: %s %s %s\n", dialAddr, key.Type(), base64.StdEncoding.EncodeToString(key.Marshal()))
    return nil
}

// format server key
func TrustKey(dialAddr string, addr net.Addr, key ssh.PublicKey) error {
    // format host record line for known_hosts file
    line := fmt.Sprintf("%s %s %s\n", dialAddr, key.Type(), base64.StdEncoding.EncodeToString(key.Marshal()))

    if fileExists(ssh_known_hosts) {
        b, err := ioutil.ReadFile(ssh_known_hosts)
        if err != nil {
            panic(err)
        }
        // check if file contains line
        s := string(b)
        // //check whether s contains substring text
        fmt.Println(strings.Contains(s, line))
        if strings.Contains(s, line) {
            log.Println("Host " + url + " already added to known_hosts file. Remove -accept flag/environment variable and run again. Exiting.")
            os.Exit(0)
        }
    }

    // add line to file
    f, err := os.OpenFile(ssh_known_hosts, os.O_APPEND|os.O_WRONLY, 0600)
    if err != nil {
        panic(err)
    }

    defer f.Close()

    if _, err = f.WriteString(line); err != nil {
        panic(err)
    }
    log.Println("Remote host key added to known_hosts file. Exiting.")
    os.Exit(0)
    return nil
}

