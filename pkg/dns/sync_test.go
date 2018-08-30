package dns

import (
	"testing"
)

func TestSyncZoneAndRecordCreate(t *testing.T) {
	svc := &FakeDNSService{}
	zone := Zone{
		Name:    "test",
		DNSName: "example.com.",
		Nameservers: []string{
			"ns1.hoster.com",
			"ns2.hoster.com",
		},
	}
	records := []Record{
		AddressRecord{
			BaseRecord: BaseRecord{
				Name: "example.com.",
				TTL:  25,
			},
			Addresses: []string{
				"1.2.3.4",
				"2.3.4.5",
			},
		},
		CNameRecord{
			BaseRecord: BaseRecord{
				Name: "cname.example.com.",
				TTL:  125,
			},
			CanonicalName: "somewhere.else.com",
		},
		NSRecord{
			BaseRecord: BaseRecord{
				Name: "www.example.com",
				TTL:  525,
			},
			Nameservers: []string{
				"ns1.company.com",
				"ns2.company.com",
			},
		},
	}

	if err := Sync(svc, zone, records); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if _, exists := svc.ZoneMap[zone.Name]; !exists {
		t.Errorf("expected zone '%s' to exist in %v", zone.Name, svc.ZoneMap)
	}

	recordsOut, err := svc.Records(zone.Name)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	expectRecordSetsEqual(records, recordsOut, t)
}

func TestSyncZoneUpdate(t *testing.T) {
	svc := &FakeDNSService{}
	svc.ZoneMap = map[string]Zone{
		"test": Zone{
			Name:    "test",
			DNSName: "example.com.",
			Nameservers: []string{
				"ns1.other.com",
				"ns2.other.com",
			},
		},
	}
	svc.RecordMap = map[string]FakeRecords{}

	zone := Zone{
		Name:    "test",
		DNSName: "example.com.",
		Nameservers: []string{
			"ns1.hoster.com",
			"ns2.hoster.com",
		},
	}

	if err := syncZone(svc, zone); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !zonesEqual(zone, svc.ZoneMap[zone.Name]) {
		t.Errorf("unexpected inequality: %v vs %v", zone, svc.ZoneMap[zone.Name])
	}
}

func TestRecordUpdate(t *testing.T) {
	svc := &FakeDNSService{}
	zone := Zone{
		Name:    "test",
		DNSName: "example.com.",
		Nameservers: []string{
			"ns1.hoster.com",
			"ns2.hoster.com",
		},
	}
	records := []Record{
		AddressRecord{
			BaseRecord: BaseRecord{
				Name: "example.com.",
				TTL:  25,
			},
			Addresses: []string{
				"1.2.3.4",
				"2.3.4.5",
			},
		},
		CNameRecord{
			BaseRecord: BaseRecord{
				Name: "cname.example.com.",
				TTL:  125,
			},
			CanonicalName: "somewhere.else.com",
		},
		NSRecord{
			BaseRecord: BaseRecord{
				Name: "www.example.com",
				TTL:  525,
			},
			Nameservers: []string{
				"ns1.company.com",
				"ns2.company.com",
			},
		},
	}

	if err := Sync(svc, zone, records); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	records[1] = CNameRecord{
		BaseRecord: BaseRecord{
			Name: "cname.example.com",
			TTL:  725,
		},
		CanonicalName: "alternative.else.com",
	}

	if err := Sync(svc, zone, records); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	recordsOut, err := svc.Records(zone.Name)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	expectRecordSetsEqual(records, recordsOut, t)
}

func TestRecordDelete(t *testing.T) {
	svc := &FakeDNSService{}
	zone := Zone{
		Name:    "test",
		DNSName: "example.com.",
		Nameservers: []string{
			"ns1.hoster.com",
			"ns2.hoster.com",
		},
	}
	records := []Record{
		AddressRecord{
			BaseRecord: BaseRecord{
				Name: "example.com.",
				TTL:  25,
			},
			Addresses: []string{
				"1.2.3.4",
				"2.3.4.5",
			},
		},
		CNameRecord{
			BaseRecord: BaseRecord{
				Name: "cname.example.com.",
				TTL:  125,
			},
			CanonicalName: "somewhere.else.com",
		},
		NSRecord{
			BaseRecord: BaseRecord{
				Name: "www.example.com",
				TTL:  525,
			},
			Nameservers: []string{
				"ns1.company.com",
				"ns2.company.com",
			},
		},
	}

	if err := Sync(svc, zone, records); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	records = []Record{
		records[0],
		records[2],
	}

	if err := Sync(svc, zone, records); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	recordsOut, err := svc.Records(zone.Name)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(recordsOut) != 2 {
		t.Errorf("expected only two records.")
	}
	expectRecordSetsEqual(records, recordsOut, t)
}

func expectRecordSetsEqual(r1 []Record, r2 []Record, t *testing.T) {
	if len(r1) != len(r2) {
		t.Errorf("unexpected record set: %v vs %v", r1, r2)
		t.FailNow()
	}
	recordMap := map[string]Record{}
	for _, record := range r1 {
		recordMap[record.RecordName()] = record
	}

	for _, record := range r2 {
		if recordIsDifferent(record, recordMap[record.RecordName()]) {
			t.Errorf("unexpected record set difference: %v vs %v", record, recordMap[record.RecordName()])
		}
	}
}
