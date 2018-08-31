package cloud

import (
	"context"
	"os"
	"strings"

	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
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
	authorizer, err := auth.NewAuthorizerFromFile(azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		return nil, err
	}
	subscription := os.Getenv("AZURE_SUBSCRIPTION")
	service := &azureDNS{
		zonesClient:   azuredns.NewZonesClient(subscription),
		recordsClient: azuredns.NewRecordSetsClient(subscription),
		resourceGroup: os.Getenv("AZURE_RESOURCE_GROUP"),
	}
	service.zonesClient.Authorizer = authorizer
	service.recordsClient.Authorizer = authorizer
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
	_, err := g.zonesClient.CreateOrUpdate(context.TODO(), g.resourceGroup, removeTrailingDot(zone.DNSName), makeAzureZone(zone), "", "")
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
	name := removeTrailingDot(removeSuffix(newRecord.RecordName(), zone.DNSName))
	recordType := newRecord.Type()
	recordSet := azuredns.RecordSet{
		Name:                &name,
		Type:                &recordType,
		RecordSetProperties: &properties,
	}
	_, err := g.recordsClient.CreateOrUpdate(context.TODO(), g.resourceGroup, removeTrailingDot(zone.DNSName), name, azuredns.RecordType(newRecord.Type()), recordSet, "", "")
	return err
}

func (g *azureDNS) Records(zone dns.Zone) ([]dns.Record, error) {
	list, err := g.recordsClient.ListAllByDNSZone(context.TODO(), g.resourceGroup, removeTrailingDot(zone.DNSName), nil, "")
	if err != nil {
		return nil, err
	}
	items := list.Values()
	result := []dns.Record{}
	for ix := range items {
		record := makeRecordFromAzureRecord(zone, items[ix])
		if record != nil {
			result = append(result, record)
		}
	}
	return result, nil
}

func (g *azureDNS) DeleteRecord(zone dns.Zone, record dns.Record) error {
	_, err := g.recordsClient.Delete(context.TODO(), g.resourceGroup, removeTrailingDot(zone.DNSName), record.RecordName(), azuredns.RecordType(record.Type()), "")
	return err
}

func makeRecordFromAzureRecord(zone dns.Zone, record azuredns.RecordSet) dns.Record {
	name := *record.Name + "." + zone.DNSName
	switch *record.Type {
	case "A":
		return dns.AddressRecord{
			BaseRecord: dns.BaseRecord{
				Name: name,
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
				Name: name,
				Kind: "NS",
				TTL:  *record.TTL,
			},
			Nameservers: nameservers,
		}
	case "CNAME":
		return dns.CNameRecord{
			BaseRecord: dns.BaseRecord{
				Name: name,
				Kind: "CNAME",
				TTL:  *record.TTL,
			},
			CanonicalName: *(*record.RecordSetProperties.CnameRecord).Cname,
		}
	}
	return nil
}

func tagOrEmptyString(tags map[string]*string, key string) string {
	if ptr := tags[key]; ptr != nil {
		return *ptr
	}
	return ""
}

func makeZone(zone *azuredns.Zone) dns.Zone {
	return dns.Zone{
		Name:        tagOrEmptyString(zone.Tags, "name"),
		Description: tagOrEmptyString(zone.Tags, "description"),
		DNSName:     addTrailingDot(*zone.Name),
		Nameservers: *zone.ZoneProperties.NameServers,
	}
}

func strPtr(val string) *string { return &val }

func removeTrailingDot(val string) string {
	if strings.HasSuffix(val, ".") {
		return val[0 : len(val)-1]
	}
	return val
}

func addTrailingDot(val string) string {
	return val + "."
}

func makeAzureZone(zone dns.Zone) azuredns.Zone {
	return azuredns.Zone{
		Name:     strPtr(removeTrailingDot(zone.DNSName)),
		Location: strPtr("global"),
		Tags: map[string]*string{
			"name":        &zone.Name,
			"description": &zone.Description,
		},
		ZoneProperties: &azuredns.ZoneProperties{
			NameServers: &zone.Nameservers,
		},
	}
}

func removeSuffix(str string, suffix string) string {
	if strings.HasSuffix(str, suffix) {
		return str[0 : len(str)-len(suffix)]
	}
	return str
}
