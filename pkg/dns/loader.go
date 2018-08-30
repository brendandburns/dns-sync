package dns

import (
	"encoding/json"
	"fmt"
	"strings"
)

func (c *Config) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage
	err := json.Unmarshal(b, &objMap)
	if err != nil {
		return err
	}
	zoneMessage, exists := objMap["zone"]
	if exists {
		if err := json.Unmarshal(*zoneMessage, &c.Zone); err != nil {
			return err
		}
	}

	recordMessage, exists := objMap["records"]
	if !exists || recordMessage == nil {
		return nil
	}
	var recordMessages []*json.RawMessage
	if err := json.Unmarshal(*recordMessage, &recordMessages); err != nil {
		return err
	}

	c.Records = make([]Record, len(recordMessages))

	for ix, msg := range recordMessages {
		obj := map[string]interface{}{}
		if err := json.Unmarshal(*msg, &obj); err != nil {
			return err
		}
		kind := strings.ToUpper(obj["kind"].(string))
		switch kind {
		case "A":
			record := AddressRecord{}
			json.Unmarshal(*msg, &record)
			c.Records[ix] = record
		case "NS":
			record := NSRecord{}
			json.Unmarshal(*msg, &record)
			c.Records[ix] = record
		case "CNAME":
			record := CNameRecord{}
			json.Unmarshal(*msg, &record)
			c.Records[ix] = record
		default:
			return fmt.Errorf("Unknown record type: %v", kind)
		}
	}
	return nil
}
