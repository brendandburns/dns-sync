package dns

import (
	"github.com/golang/glog"
)

func Sync(service Service, zone Zone, records []Record) error {
	glog.Info("Syncing zones.")
	if err := syncZone(service, zone); err != nil {
		return err
	}
	glog.Info("Syncing records.")
	if err := syncRecords(service, zone, records); err != nil {
		return err
	}
	return nil
}

func syncZone(service Service, zone Zone) error {
	currentZones, err := service.Zones()
	if err != nil {
		return err
	}
	glog.V(2).Infof("Current zones: %v\n", currentZones)
	var existingZone *Zone
	for ix := range currentZones {
		if currentZones[ix].Name == zone.Name {
			existingZone = &currentZones[ix]
		}
	}
	if existingZone == nil {
		glog.V(2).Info("Creating new zone.")
		return service.WriteZone(zone, true)
	}
	if !zonesEqual(zone, *existingZone) {
		glog.V(2).Info("Updating zone.")
		return service.WriteZone(zone, false)
	}
	return nil
}

func syncRecords(service Service, zone Zone, records []Record) error {
	existingRecords, err := service.Records(zone)
	if err != nil {
		return err
	}
	glog.V(2).Infof("Current records: %v", existingRecords)
	for _, record := range records {
		existingRecord := findRecord(record.RecordName(), existingRecords)
		if existingRecord != nil {
			if recordIsDifferent(record, *existingRecord) {
				glog.V(2).Infof("Updating record: %v", record)
				if err := service.WriteRecord(zone, *existingRecord, record); err != nil {
					return err
				}
			}
		} else {
			glog.V(2).Infof("Creating record: %v", record)
			if err := service.WriteRecord(zone, nil, record); err != nil {
				return err
			}
		}
	}
	for _, record := range existingRecords {
		desiredRecord := findRecord(record.RecordName(), records)
		// Maintain the apex NS record no matter what.
		if record.RecordName() == zone.DNSName {
			continue
		}

		if desiredRecord == nil {
			if err := service.DeleteRecord(zone, record); err != nil {
				return err
			}
		}
	}
	return nil
}

func findRecord(name string, records []Record) *Record {
	if len(records) == 0 {
		return nil
	}
	for ix := range records {
		if records[ix].RecordName() == name {
			return &records[ix]
		}
	}
	return nil
}

func recordIsDifferent(r1 Record, r2 Record) bool {
	if r1.RecordName() != r2.RecordName() {
		return true
	}
	if r1.TimeToLive() != r2.TimeToLive() {
		return true
	}
	if r1.Type() != r2.Type() {
		return true
	}
	rr1 := r1.RRData()
	rr2 := r2.RRData()
	if len(rr1) != len(rr2) {
		return true
	}
	for ix := range rr1 {
		if rr1[ix] != rr2[ix] {
			return true
		}
	}
	return false
}

func zonesEqual(z1 Zone, z2 Zone) bool {
	if z1.Name != z2.Name ||
		z1.DNSName != z2.DNSName ||
		z1.Description != z2.Description ||
		len(z1.Nameservers) != len(z2.Nameservers) {
		return false
	}
	for ix := range z1.Nameservers {
		if z1.Nameservers[ix] != z2.Nameservers[ix] {
			return false
		}
	}
	return true
}
