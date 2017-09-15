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
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

type Backendcfg struct {
	Upstream      string
	Context       string
	Servers       []Backend
	Port          string
	Task_dns      string
	Domain_prefix string
	Domain_zone   string
}

type T struct {
	Debug   bool
	General struct {
		Check_intervall int64
		Domain_prefix   string
		Domain_zone     string
		Resources       map[string]*Backendcfg
	}
	Pdns struct {
		Api_url       string
		Api_key       string
		Ip_address    string
		Domain_prefix string
		Domain_zone   string
	}
}

func ReadConfigfile() (ok bool, config T) {
	cfgdata, err := ioutil.ReadFile("/config/border-controller.yml")

	if err != nil {
		log.Panic("Cannot open config file from /config/border-controller.yml")
	}

	t := T{}

	err = yaml.Unmarshal([]byte(cfgdata), &t)
	if err != nil {
		log.Panic("Cannot map yml config file to interface, possible syntax error")
		log.Panic(err)
	}

	return true, t
}
