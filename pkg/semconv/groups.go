package semconv

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

	CanonicalId string
}
