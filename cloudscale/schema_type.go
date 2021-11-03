package cloudscale

type SchemaType int

const (
	RESOURCE SchemaType = iota
	DATASOURCE
)

func (t SchemaType) isDatasource() bool {
	switch t {
	case RESOURCE:
		return false
	case DATASOURCE:
		return true
	}
	panic("unknown SchemaType")
}

func (t SchemaType) isResource() bool {
	switch t {
	case RESOURCE:
		return true
	case DATASOURCE:
		return false
	}
	panic("unknown SchemaType")
}
