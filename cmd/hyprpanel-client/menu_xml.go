package main

import "encoding/xml"

type menuXMLInterface struct {
	XMLName xml.Name     `xml:"interface"`
	Menu    *menuXMLMenu `xml:"menu"`
}

type menuXMLMenu struct {
	XMLName  xml.Name              `xml:"menu"`
	ID       string                `xml:"id,attr,omitempty"`
	Sections []*menuXMLMenuSection `xml:"section"`
}

type menuXMLMenuSection struct {
	XMLName    xml.Name              `xml:"section"`
	ID         string                `xml:"id,attr,omitempty"`
	Items      []*menuXMLItem        `xml:"item"`
	Submenus   []*menuXMLMenuSubmenu `xml:"submenu"`
	Attributes []*menuXMLAttribute   `xml:"attribute"`
}

type menuXMLMenuSubmenu struct {
	XMLName    xml.Name              `xml:"submenu"`
	ID         string                `xml:"id,attr,omitempty"`
	Sections   []*menuXMLMenuSection `xml:"section"`
	Attributes []*menuXMLAttribute   `xml:"attribute"`
}

type menuXMLItem struct {
	XMLName    xml.Name            `xml:"item"`
	Attributes []*menuXMLAttribute `xml:"attribute"`
	Links      []*mnuXMLLink       `xml:"link"`
}

type menuXMLAttribute struct {
	XMLName      xml.Name `xml:"attribute"`
	Name         string   `xml:"name,attr"`
	Translatable string   `xml:"translatable,attr,omitempty"`
	Value        string   `xml:",chardata"`
}

type mnuXMLLink struct {
	XMLName xml.Name `xml:"link"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:",chardata"`
}
