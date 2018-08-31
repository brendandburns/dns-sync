package main

import (
	"flag"
	"io/ioutil"
	"log"

	"github.com/brendandburns/dns-sync/pkg/dns"
	"github.com/brendandburns/dns-sync/pkg/dns/cloud"
	"github.com/ghodss/yaml"
	"github.com/golang/glog"
)

var (
	configFile = flag.String("config", "", "Path to config file")
	cloudDNS   = flag.String("cloud", "", "Which cloud DNS provider to use, currently 'google' or 'azure'")
)

func main() {
	flag.Parse()

	if len(*configFile) == 0 {
		log.Fatal("--config is required.")
	}
	config := dns.Config{}
	data, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatal(err.Error())
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Fatal(err.Error())
	}

	glog.V(4).Infof("LoadedConfig: %v\n", config)

	var svc dns.Service
	if len(*cloudDNS) == 0 || *cloudDNS == "google" {
		svc, err = cloud.NewGoogleCloudDNSService()
	} else if *cloudDNS == "azure" {
		svc, err = cloud.NewAzureDNSService()
	}
	if err != nil {
		log.Fatal(err.Error())
	}

	if err := dns.Sync(svc, config.Zone, config.Records); err != nil {
		log.Fatal(err.Error())
	}
	log.Println("Synchronized.")
}
