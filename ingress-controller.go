/*
Copyright 2017 Mario Kleinsasser and Bernhard Rausch

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"text/template"
	"time"
)

var issd string
var issp string
var issthp string
var mainloop bool

type Backend struct {
	Hostname string
	Port     string
}

func isprocessrunning() (running bool) {
	// check if traefik is healthy
	var run bool
	run = true

	resp, err := http.Get("http://localhost:" + issthp + "/health")

	if err != nil {
		log.Print("Error getting traefik status. Trying to start traefik. NIL!")
		log.Print(err)
		return false
	}

	if resp.StatusCode != 200 {
		log.Print("Error getting traefik status. Trying to start traefik.")
		log.Print(string(resp.StatusCode))
		return false
	}

	log.Print("Traefik is running. All OK!")
	return run
}

func startprocess() {
	log.Print("Start Process!")
	cmd := exec.Command("/data/traefik", "-c", "/config/traefik.toml")
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
		mainloop = false
	}

}

func checkconfig(backends []Backend) {

	//  open template
	t, err := template.ParseFiles("/config/ingress-controller-config.tpl")
	if err != nil {
		log.Print(err)
		return
	}

	// process template
	var tpl bytes.Buffer
	err = t.Execute(&tpl, backends)
	if err != nil {
		log.Print(err)
		return
	}

	// create md5 of result
	md5tpl := fmt.Sprintf("%x", md5.Sum([]byte(tpl.String())))
	log.Print("MD5 of TPL: " + md5tpl)
	log.Print("TPL: " + tpl.String())

	// open existing config, read it to memory
	exconf, errexconf := ioutil.ReadFile("/config/traefik.toml")
	if errexconf != nil {
		log.Print("Cannot read existing conf!")
		log.Print(errexconf)
	}

	md5exconf := fmt.Sprintf("%x", md5.Sum(exconf))
	log.Print("MD5 of EXCONF: " + md5exconf)

	// comapre md5 and write config if needed
	if md5tpl == md5exconf {
		log.Print("MD5 sums equal! Nothing to do.")
		return
	}

	log.Print("MD5 sums different writing new conf!")

	// overwrite existing conf
	err = ioutil.WriteFile("/config/traefik.toml", []byte(tpl.String()), 0644)
	if err != nil {
		log.Print("Cannot write config file.")
		log.Print(err)
		mainloop = false
	}

}

func querydns() (bends []Backend) {

	// we dont kill the mainloop if dns not working because it could come online later
	var data []string
	var backends []Backend
	var b Backend

	// lookup dns name
	be, errr := net.LookupIP(issd)

	if errr != nil {
		log.Print(errr)
	}

	// save to simple string slice
	for _, ip := range be {
		data = append(data, ip.String())
	}

	sort.Strings(data)

	// convert it to struct
	for _, ip := range data {
		b.Hostname = ip
		b.Port = issp
		backends = append(backends, b)
	}

	return backends
}

func checkenvironment() {

	var errb bool

	// The following environments must be present
	issd, errb = os.LookupEnv("INGRESS_STACK_SERVICE_DNS")
	if (!errb) || (issd == "") {
		log.Panic("No environment variable INGRESS_STACK_SERVICE_DNS or empty value!")
	}

	issp, errb = os.LookupEnv("INGRESS_STACK_SERVICE_PORT")
	if (!errb) || (issp == "") {
		log.Panic("No environment variable INGRESS_STACK_SERVICE_PORT or empty value!")
	}

	issthp, errb = os.LookupEnv("INGRESS_STACK_TRAEFIKHEALTH_PORT")
	if (!errb) || (issp == "") {
		log.Panic("No environment variable INGRESS_STACK_TRAEFIKHEALTH_PORT or empty value!")
	}

}

func main() {

	checkenvironment()

	// now checkconfig, this will loop forever
	mainloop = true
	for mainloop == true {

		backends := querydns()

		checkconfig(backends)

		if !isprocessrunning() {
			startprocess()
		}

		time.Sleep(10 * time.Second)
	}
}
