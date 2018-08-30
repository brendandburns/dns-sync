package dns

type Config struct {
	Zone    Zone     `json:"zone" yaml:"zone"`
	Records []Record `json:"records" yaml:"records"`
}

type Zone struct {
	Name        string   `json:"name" yaml:"name"`
	DNSName     string   `json:"dnsName" yaml:"dnsName"`
	Nameservers []string `json:"nameservers" yaml:"nameservers"`
	Description string   `json:"description" yaml:"description"`
}

type Record interface {
	Type() string
	RecordName() string
	TimeToLive() int64
	RRData() []string
}

type BaseRecord struct {
	Name string `json:"name" yaml:"name"`
	TTL  int64  `json:"ttl" yaml:"ttl"`
	Kind string `json:"kind" yaml:"kind"`
}

func (b BaseRecord) Type() string {
	return b.Kind
}

func (b BaseRecord) RecordName() string {
	return b.Name
}

func (b BaseRecord) TimeToLive() int64 {
	return b.TTL
}

type AddressRecord struct {
	BaseRecord `json:",inline" yaml:",inline"`
	Addresses  []string `json:"addresses" yaml:"addresses"`
}

func (a AddressRecord) RRData() []string {
	return a.Addresses
}

var _ = Record(AddressRecord{})

type CNameRecord struct {
	BaseRecord
	CanonicalName string `json:"canonicalName" yaml:"canonicalName"`
}

func (c CNameRecord) RRData() []string {
	return []string{c.CanonicalName}
}

var _ = Record(CNameRecord{})

type NSRecord struct {
	BaseRecord
	Nameservers []string `json:"nameservers" yaml:"nameservers"`
}

func (n NSRecord) RRData() []string {
	return n.Nameservers
}

var _ = Record(NSRecord{})
