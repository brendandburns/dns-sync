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

	svc, err := cloud.NewGoogleCloudDNSService()
	if err != nil {
		log.Fatal(err.Error())
	}

	if err := dns.Sync(svc, config.Zone, config.Records); err != nil {
		log.Fatal(err.Error())
	}
	log.Println("Synchronized.")
}
