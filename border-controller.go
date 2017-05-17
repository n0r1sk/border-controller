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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"
)

var mainloop bool

type Message struct {
	Acode   int64
	Astring string
	Aslice  []string
}

func isprocessrunning() (running bool) {
	// check if nginx is healthy
	var run bool
	run = true

	// do not follow redirects
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}}

	// default nignx is binding to port 80 on default
	_, err := client.Get("http://localhost")

	if err != nil {
		log.Print("Error getting body from nginx. Trying to start nginx. NIL!")
		log.Print(err)
		return false
	}

	log.Print("nginx is running. All OK!")
	return run
}

func startprocess() {
	log.Print("Start Process!")
	cmd := exec.Command("nginx", "-g", "daemon off;")
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
		mainloop = false
	}

}

func reloadprocess() {
	log.Print("Reloading Process!")
	cmd := exec.Command("nginx", "-s", "reload")
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	cmd.Wait()
}

func checkconfig(be []string, domain string) (changed bool) {

	// first sort the string slice
	sort.Strings(be)

	type Backend struct {
		Node string
		Port string
	}

	var data []Backend

	for _, e := range be {
		var t Backend
		et := strings.Split(e, " ")
		t.Node = et[0] + "." + domain
		t.Port = et[1]
		data = append(data, t)
	}

	//  open template
	t, err := template.ParseFiles("/config/border-controller-config.tpl")
	if err != nil {
		log.Print(err)
		return false
	}

	// process template
	var tpl bytes.Buffer
	err = t.Execute(&tpl, data)
	if err != nil {
		log.Print(err)
		return false
	}

	// create md5 of result
	md5tpl := fmt.Sprintf("%x", md5.Sum([]byte(tpl.String())))
	log.Print("MD5 of TPL: " + md5tpl)
	log.Print("TPL: " + tpl.String())

	// open existing config, read it to memory
	exconf, errexconf := ioutil.ReadFile("/etc/nginx/nginx.conf")
	if errexconf != nil {
		log.Print("Cannot read existing conf!")
		log.Print(errexconf)
	}

	md5exconf := fmt.Sprintf("%x", md5.Sum(exconf))
	log.Print("MD5 of EXCONF: " + md5exconf)

	// comapre md5 and write config if needed
	if md5tpl == md5exconf {
		log.Print("MD5 sums equal! Nothing to do.")
		return false
	}

	log.Print("MD5 sums different writing new conf!")

	// overwrite existing conf
	err = ioutil.WriteFile("/etc/nginx/nginx.conf", []byte(tpl.String()), 0644)
	if err != nil {
		log.Print("Cannot write config file.")
		log.Print(err)
		mainloop = false
	}

	return true

}

func getstackerviceinfo(config T) (backends []string, err error) {

	var m Message

	for _, dh := range config.General.Swarm.Docker_hosts {
		log.Print(dh)

		resp, err := http.Get("http://" + dh + "." +
			config.General.Swarm.Docker_host_dns_domain + ":" +
			config.General.Swarm.Docker_controller.Exposed_port +
			"/service/inspect/" + config.General.Swarm.Ingress_service_name +
			"?api_key=" + config.General.Swarm.Docker_controller.Api_key)

		if err != nil {
			log.Print(err)
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode == 200 { // OK
			bodyBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, errors.New("Error reading response body")
			}

			err = json.Unmarshal(bodyBytes, &m)
			if err != nil {
				return nil, errors.New("Error reading during unmarshal of response body.")
			}

			if m.Acode >= 500 {
				return nil, errors.New(strconv.Itoa(int(m.Acode)) + " " + m.Astring)
			}
		}

		return m.Aslice, nil

	}

	return nil, errors.New("Cannot reach any docker host")

}

func main() {

	config, ok := ReadConfigfile()
	if !ok {
		log.Panic("Error during config parsing")
	}

	// now checkconfig, this will loop forever
	mainloop = true
	for mainloop == true {

		backends, err := getstackerviceinfo(config)

		if err != nil {
			log.Print(err)
			time.Sleep(2 * time.Second)
			continue
		}

		changed := checkconfig(backends, config.General.Swarm.Docker_host_dns_domain)

		if changed == true {
			if isprocessrunning() {
				reloadprocess()
			} else {
				startprocess()
			}
		} else {
			if !isprocessrunning() {
				startprocess()
			}
		}

		time.Sleep(5 * time.Second)
	}
}
