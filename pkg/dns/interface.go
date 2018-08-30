package dns

type Service interface {
	Zones() ([]Zone, error)
	WriteZone(zone Zone, create bool) error
	DeleteZone(zone Zone) error

	Records(zone Zone) ([]Record, error)
	WriteRecord(zone Zone, oldRecord, record Record) error
	DeleteRecord(zone Zone, record Record) error
}
