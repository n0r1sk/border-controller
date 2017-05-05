package main

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"sort"
	"text/template"
	"time"
)

var issd string
var issp string

type Backend struct {
	Hostname string
	Port     string
}

func checkprocess() {
	// TODO check if process running and restart
}

func startprocess() {
	log.Print("Start Process!")
	cmd := exec.Command("nginx", "-g", "daemon off;")
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
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

func checkconfig(tplpath string, confpath string, backends []Backend) (reload bool) {

	//  open template
	t, err := template.ParseFiles(tplpath)
	if err != nil {
		log.Print(err)
		return
	}

	// process template
	var tpl bytes.Buffer
	err = t.Execute(&tpl, backends)
	if err != nil {
		log.Print("execute: ", err)
		return
	}

	// create md5 of result
	md5tpl := fmt.Sprintf("%x", md5.Sum([]byte(tpl.String())))
	log.Print("MD5 of TPL: " + md5tpl)
	log.Print("TPL: " + tpl.String())

	// open existing config, read it to memory
	exconf, errexconf := ioutil.ReadFile(confpath)
	if errexconf != nil {
		panic("Cannot read existing conf!")
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
	err = ioutil.WriteFile(confpath, []byte(tpl.String()), 0644)

	return true

}

func querydns() (bends []Backend) {
	var data []string
	var backends []Backend
	var b Backend

	// lookup dns name
	be, errr := net.LookupIP(issd)

	if errr != nil {
		fmt.Println(errr)
	}

	// save to simple string slice
	for _, ip := range be {
		data = append(data, ip.String())
	}

	sort.Strings(data)

	// convert it to struct slice/etc/ngingx/nginx.conf
	for _, ip := range data {
		b.Hostname = ip
		b.Port = issp
		backends = append(backends, b)
	}

	return backends
}

func main() {

	var errb bool
	issd, errb = os.LookupEnv("INGRESS_STACK_SERVICE_DNS")

	if !errb {
		panic("No environment variable INGRESS_STACK_SERVICE_DNS!")
	}

	if issd == "" {
		panic("Environment variable INGRESS_STACK_SERVICE_DNS is empty!")
	}

	issp, errb = os.LookupEnv("INGRESS_STACK_SERVICE_PORT")

	if !errb {
		panic("No environment variable INGRESS_STACK_SERVICE_PORT!")
	}

	if issp == "" {
		panic("Environment variable INGRESS_STACK_SERVICE_PORT is empty!")
	}

	// start loadbalancer
	startprocess()

	// now checkconfig, this will loop forever
	exit := false
	for exit == false {
		backends := querydns()
		reload := checkconfig("/data/ingress-controller-nginx.tpl", "/etc/nginx/nginx.conf", backends)
		if reload == true {
			reloadprocess()
		}
		time.Sleep(5 * time.Second)
	}
}
