package dns

type Service interface {
	Zones() ([]Zone, error)
	WriteZone(zone Zone, create bool) error
	DeleteZone(zone Zone) error

	Records(zone string) ([]Record, error)
	WriteRecord(zone string, oldRecord, record Record) error
	DeleteRecord(zone string, record Record) error
}
