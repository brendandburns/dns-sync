package cloud

import (
	"context"

	azuredns "github.com/azure/azure-sdk-for-go/services/preview/dns/mgmt/2018-03-01-preview/dns"
	"github.com/brendandburns/dns-sync/pkg/dns"
)

type azureDNS struct {
	recordsClient azuredns.RecordSetsClient
	zonesClient   azuredns.ZonesClient
	resourceGroup string
}

var _ = dns.Service(&azureDNS{})

func NewAzureDNSService() (dns.Service, error) {
	subscription := "foo"
	service := &azureDNS{
		zonesClient:   azuredns.NewZonesClient(subscription),
		recordsClient: azuredns.NewRecordSetsClient(subscription),
	}
	return service, nil
}

func (g *azureDNS) Zones() ([]dns.Zone, error) {
	list, err := g.zonesClient.List(context.TODO(), nil)
	if err != nil {
		return nil, err
	}
	result := []dns.Zone{}
	for ix := range list.Values() {
		result = append(result, makeZone(&list.Values()[ix]))
	}
	return result, nil
}

func (g *azureDNS) WriteZone(zone dns.Zone, create bool) error {
	_, err := g.zonesClient.CreateOrUpdate(context.TODO(), g.resourceGroup, zone.DNSName, makeAzureZone(zone), "", "")
	return err
}

func (g *azureDNS) DeleteZone(zone dns.Zone) error {
	_, err := g.zonesClient.Delete(context.TODO(), g.resourceGroup, zone.DNSName, "")
	return err
}

func (g *azureDNS) WriteRecord(zone dns.Zone, oldRecord, newRecord dns.Record) error {
	ttl := newRecord.TimeToLive()
	properties := azuredns.RecordSetProperties{
		TTL: &ttl,
	}
	switch newRecord.Type() {
	case "A":
		properties.ARecords = &[]azuredns.ARecord{
			azuredns.ARecord{
				Ipv4Address: &newRecord.RRData()[0],
			},
		}
	case "NS":
		rrdata := newRecord.RRData()
		arr := make([]azuredns.NsRecord, len(rrdata))
		for ix := range rrdata {
			arr[ix] = azuredns.NsRecord{
				Nsdname: &newRecord.RRData()[0],
			}
		}
		properties.NsRecords = &arr
	case "CNAME":
		properties.CnameRecord = &azuredns.CnameRecord{
			Cname: &newRecord.RRData()[0],
		}
	}
	name := newRecord.RecordName()
	recordType := newRecord.Type()
	recordSet := azuredns.RecordSet{
		Name:                &name,
		Type:                &recordType,
		RecordSetProperties: &properties,
	}
	_, err := g.recordsClient.CreateOrUpdate(context.TODO(), g.resourceGroup, zone.DNSName, newRecord.RecordName(), azuredns.RecordType(newRecord.Type()), recordSet, "", "")
	return err
}

func (g *azureDNS) Records(zone dns.Zone) ([]dns.Record, error) {
	list, err := g.recordsClient.ListAllByDNSZone(context.TODO(), g.resourceGroup, zone.DNSName, nil, "")
	if err != nil {
		return nil, err
	}
	items := list.Values()
	result := make([]dns.Record, len(items))
	for ix := range items {
		result[ix] = makeRecordFromAzureRecord(items[ix])
	}
	return result, nil
}

func (g *azureDNS) DeleteRecord(zone dns.Zone, record dns.Record) error {
	_, err := g.recordsClient.Delete(context.TODO(), g.resourceGroup, zone.DNSName, record.RecordName(), azuredns.RecordType(record.Type()), "")
	return err
}

func makeRecordFromAzureRecord(record azuredns.RecordSet) dns.Record {
	switch *record.Type {
	case "A":
		return dns.AddressRecord{
			BaseRecord: dns.BaseRecord{
				Name: *record.Name,
				Kind: "A",
				TTL:  *record.TTL,
			},
			Addresses: []string{*(*record.RecordSetProperties.ARecords)[0].Ipv4Address},
		}
	case "NS":
		nameservers := []string{}
		for _, record := range *record.RecordSetProperties.NsRecords {
			nameservers = append(nameservers, *record.Nsdname)
		}
		return dns.NSRecord{
			BaseRecord: dns.BaseRecord{
				Name: *record.Name,
				Kind: "NS",
				TTL:  *record.TTL,
			},
			Nameservers: nameservers,
		}
	case "CNAME":
		return dns.CNameRecord{
			BaseRecord: dns.BaseRecord{
				Name: *record.Name,
				Kind: "CNAME",
				TTL:  *record.TTL,
			},
			CanonicalName: *(*record.RecordSetProperties.CnameRecord).Cname,
		}
	}
	return nil
}

func makeZone(zone *azuredns.Zone) dns.Zone {
	return dns.Zone{
		Name:        *zone.Tags["name"],
		Description: *zone.Tags["description"],
		DNSName:     *zone.Name,
		Nameservers: *zone.ZoneProperties.NameServers,
	}
}

func makeAzureZone(zone dns.Zone) azuredns.Zone {
	return azuredns.Zone{
		Name: &zone.DNSName,
		Tags: map[string]*string{
			"name":        &zone.Name,
			"description": &zone.Description,
		},
		ZoneProperties: &azuredns.ZoneProperties{
			NameServers: &zone.Nameservers,
		},
	}
}
