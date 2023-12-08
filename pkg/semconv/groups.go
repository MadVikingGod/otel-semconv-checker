// SPDX-License-Identifier: Apache-2.0

packagesemconv

type Group struct {
	Id         string
	Type       string
	Extends    string
	Attributes []Attribute

	Prefix string
}

type Attribute struct {
	Id  string
	Ref string
	// Type string

	// This is space to hold the prefix.name after parsing.
	CanonicalId string
}
