package cloudscale

type SchemaType int

const (
	RESOURCE SchemaType = iota
	DATA_SOURCE
)

func (t SchemaType) isDataSource() bool {
	switch t {
	case RESOURCE:
		return false
	case DATA_SOURCE:
		return true
	}
	panic("unknown SchemaType")
}

func (t SchemaType) isResource() bool {
	switch t {
	case RESOURCE:
		return true
	case DATA_SOURCE:
		return false
	}
	panic("unknown SchemaType")
}
