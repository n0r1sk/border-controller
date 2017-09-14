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
	"bufio"
	"bytes"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	log "github.com/sirupsen/logrus"
)

var mainloop bool

type Message struct {
	Acode   int64
	Astring string
	Aslice  []string
}

type Backend struct {
	Server string
	Port   string
}

func isprocessrunningps(processname string) (running bool) {

	// get all folders from proc filesystem

	files, _ := ioutil.ReadDir("/proc")
	for _, f := range files {

		// check if folder is a integer (process number)
		if _, err := strconv.Atoi(f.Name()); err == nil {
			// open status file of process
			f, err := os.Open("/proc/" + f.Name() + "/status")
			if err != nil {
				log.Println(err)
				return false
			}

			// read status line by line
			scanner := bufio.NewScanner(f)

			// check if process name in status of process
			for scanner.Scan() {

				re := regexp.MustCompile("^Name:.*" + processname + ".*")
				match := re.MatchString(scanner.Text())

				if match == true {
					return true
				}

			}

		}
	}

	return false

}

func startprocess() {
	log.Print("Start Process!")
	cmd := exec.Command("nginx", "-g", "daemon off;")

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
		mainloop = false
	}

	slurp, _ := ioutil.ReadAll(stderr)
	log.Warn(string(slurp[:]))

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

func writeconfig(data interface{}) (changed bool) {

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
	log.Debug("TPL: " + tpl.String())

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

func getstacktaskdns(task_dns string) (addrs []string, err error) {

	// resolve given service names
	servicerecords, err := net.LookupHost(task_dns)
	sort.Strings(servicerecords)

	if err != nil {
		return nil, err
	}

	log.Debug("TASK_DNS: " + task_dns + " ENTRIES: " + strings.Join(servicerecords, " "))

	return servicerecords, nil

}

func refreshconfigstruct(config T) (err error) {
	// get information on services and context configuration information

	log.Debug("Config Struct: " + fmt.Sprintf("%+v", config))
	for k, v := range config.General.Resources {
		v.Servers = nil

		// there must be a dns name given and we have to get the backend ips
		servicerecords, err := getstacktaskdns(v.Task_dns)

		if err != nil {
			log.Warn("Cannot get DNS records for config entry: " + k)
			return err
		}

		for _, s := range servicerecords {
			var b Backend
			b.Server = s
			b.Port = v.Port
			v.Servers = append(v.Servers, b)
		}

		// set the key to upstream
		v.Upstream = k

		// add values of DNS to config struct
		if v.Domain_zone == "" {
			if config.General.Domain_zone != "" {
				v.Domain_zone = config.General.Domain_zone
			} else if config.Pdns.Domain_zone != "" {
				v.Domain_zone = config.Pdns.Domain_zone
			} else {
				// leave empty
			}
		}

		if v.Domain_prefix == "" {
			if config.General.Domain_prefix != "" {
				v.Domain_prefix = config.General.Domain_prefix
			} else if config.Pdns.Domain_prefix != "" {
				v.Domain_prefix = config.Pdns.Domain_prefix
			} else {
				// leave empty
			}
		}

		// fill the dns values
		log.Debug("Config Struct: " + fmt.Sprintf("%+v", v))

	}

	return nil

}

func main() {

	// configure logrus logger
	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.FullTimestamp = true
	customFormatter.ForceColors = true
	log.SetFormatter(customFormatter)
	log.SetOutput(os.Stdout)

	config, ok := ReadConfigfile()
	if !ok {
		log.Panic("Error during config parsing")
	}

	// get debug flag from config
	if config.Debug == true {
		log.SetLevel(log.DebugLevel)
	}

	// set check intervall from config
	checkintervall := config.General.Check_intervall
	if checkintervall == 0 {
		checkintervall = 30
	}

	// check if pdns configuration is enabled
	if config.Pdns.Api_key != "" {
		updatepdns(config)
	}

	// now checkconfig, this will loop forever
	mainloop = true
	var changed bool = false

	for mainloop == true {

		// refresh config struct
		suc := refreshconfigstruct(config)
		if suc != nil {
			// on error during refresh (DNS) sleep and continue
			time.Sleep(time.Duration(checkintervall) * time.Second)
			continue
		}

		// process config
		changed = writeconfig(config.General.Resources)

		if changed == true {
			if isprocessrunningps("nginx") {
				reloadprocess()
			} else {
				startprocess()
			}
		} else {
			if !isprocessrunningps("nginx") {
				startprocess()
			}
		}

		time.Sleep(time.Duration(checkintervall) * time.Second)
	}

}
