package gac

import "encoding/xml"

const namespace = "http://knx.org/xml/ga-export/01"

type exchangeDocument struct {
	XMLName xml.Name        `xml:"GroupAddress-Export"`
	XMLNS   string          `xml:"xmlns,attr,omitempty"`
	Ranges  []exchangeRange `xml:"GroupRange"`
}

type exchangeRange struct {
	Name       string          `xml:"Name,attr"`
	RangeStart uint16          `xml:"RangeStart,attr"`
	RangeEnd   uint16          `xml:"RangeEnd,attr"`
	Addresses  []exchangeGroup `xml:"GroupAddress"`
	Ranges     []exchangeRange `xml:"GroupRange"`
}

type exchangeGroup struct {
	Name    string `xml:"Name,attr"`
	Address string `xml:"Address,attr"`
	DPTs    string `xml:"DPTs,attr"`
}
