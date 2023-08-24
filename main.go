package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

var fail = flag.Bool("fail", true, "fail pvcs")

func main() {
	flag.Parse()
	remote, err := url.Parse("https://localhost:6443")
	if err != nil {
		log.Fatal(err)
	}

	cert, err := tls.LoadX509KeyPair("/var/lib/kubelet/pki/kubelet-client-current.pem", "/var/lib/kubelet/pki/kubelet-client-current.pem")
	if err != nil {
		log.Fatal(err)
	}

	caCert, err := ioutil.ReadFile("/etc/kubernetes/pki/ca.crt")
	if err != nil {
		log.Fatal(err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caCertPool,
		},
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v1/persistentvolumes/pvc-") && *fail {
			log.Println("failing PVC call to", r.URL.Path)
			http.Error(w, "go away", http.StatusInternalServerError)
			return
		}
		proxy.ServeHTTP(w, r)
	})

	if err := http.ListenAndServeTLS(":8443", "/etc/kubernetes/pki/apiserver.crt", "/etc/kubernetes/pki/apiserver.key", nil); err != nil {
		log.Fatal(err)
	}
}
