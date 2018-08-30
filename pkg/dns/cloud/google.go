package cloud

import (
	"fmt"
	"os"

	"github.com/brendandburns/dns-sync/pkg/dns"
	cloud_dns "google.golang.org/api/dns/v1"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type googleDNS struct {
	client  *cloud_dns.Service
	project string
}

var _ = dns.Service(&googleDNS{})

func NewGoogleCloudDNSService() (dns.Service, error) {
	project := os.Getenv("GOOGLE_PROJECT")
	client, err := google.DefaultClient(oauth2.NoContext,
		"https://www.googleapis.com/auth/ndev.clouddns.readwrite")
	if err != nil {
		return nil, err
	}

	svc, err := cloud_dns.New(client)
	if err != nil {
		return nil, err
	}

	return &googleDNS{client: svc, project: project}, nil
}

func (g *googleDNS) Zones() ([]dns.Zone, error) {
	list, err := g.client.ManagedZones.List(g.project).Do()
	if err != nil {
		return nil, err
	}
	result := make([]dns.Zone, len(list.ManagedZones))
	for ix, zone := range list.ManagedZones {
		result[ix] = dns.Zone{
			Name:        zone.Name,
			DNSName:     zone.DnsName,
			Nameservers: zone.NameServers,
		}
	}
	return result, nil
}

func (g *googleDNS) WriteZone(zone dns.Zone, create bool) error {
	cloudZone := cloud_dns.ManagedZone{
		Name:        zone.Name,
		DnsName:     zone.DNSName,
		NameServers: zone.Nameservers,
		Description: zone.Description,
	}
	if create {
		_, err := g.client.ManagedZones.Create(g.project, &cloudZone).Do()
		return err
	}
	currentZone, err := g.client.ManagedZones.Get(g.project, zone.Name).Do()
	if err != nil {
		return err
	}
	currentZone.Name = zone.Name
	currentZone.DnsName = zone.DNSName
	currentZone.Description = zone.Description
	if len(zone.Nameservers) > 0 {
		currentZone.NameServers = zone.Nameservers
	}
	_, err = g.client.ManagedZones.Update(g.project, zone.Name, currentZone).Do()
	return err
}

func (g *googleDNS) DeleteZone(zone dns.Zone) error {
	return g.client.ManagedZones.Delete(g.project, zone.Name).Do()
}

func (g *googleDNS) WriteRecord(zone string, oldRecord, newRecord dns.Record) error {
	recordSet := makeRecordSet(newRecord)
	change := cloud_dns.Change{
		Additions: []*cloud_dns.ResourceRecordSet{recordSet},
	}
	if oldRecord != nil {
		deleteSet := makeRecordSet(oldRecord)
		change.Deletions = []*cloud_dns.ResourceRecordSet{deleteSet}
	}
	_, err := g.client.Changes.Create(g.project, zone, &change).Do()
	return err
}

func (g *googleDNS) Records(zone string) ([]dns.Record, error) {
	list, err := g.client.ResourceRecordSets.List(g.project, zone).Do()
	if err != nil {
		return nil, err
	}
	result := []dns.Record{}
	for _, record := range list.Rrsets {
		if cloudRecord, err := makeRecord(record); err == nil {
			result = append(result, cloudRecord)
		}
	}
	return result, nil
}

func (g *googleDNS) DeleteRecord(zone string, record dns.Record) error {
	change := cloud_dns.Change{
		Deletions: []*cloud_dns.ResourceRecordSet{makeRecordSet(record)},
	}
	_, err := g.client.Changes.Create(g.project, zone, &change).Do()
	return err
}

func makeRecordSet(record dns.Record) *cloud_dns.ResourceRecordSet {
	return &cloud_dns.ResourceRecordSet{
		Type:    record.Type(),
		Name:    record.RecordName(),
		Ttl:     record.TimeToLive(),
		Rrdatas: record.RRData(),
	}
}

func makeRecord(recordSet *cloud_dns.ResourceRecordSet) (dns.Record, error) {
	baseRecord := dns.BaseRecord{
		Name: recordSet.Name,
		TTL:  recordSet.Ttl,
		Kind: recordSet.Type,
	}
	if recordSet.Type == "A" {
		return dns.AddressRecord{
			BaseRecord: baseRecord,
			Addresses:  recordSet.Rrdatas,
		}, nil
	}
	if recordSet.Type == "CNAME" {
		return dns.CNameRecord{
			BaseRecord:    baseRecord,
			CanonicalName: recordSet.Rrdatas[0],
		}, nil
	}
	if recordSet.Type == "NS" {
		return dns.NSRecord{
			BaseRecord:  baseRecord,
			Nameservers: recordSet.Rrdatas,
		}, nil
	}
	return nil, fmt.Errorf("Unsupported record type: %s", recordSet.Type)
}
