package authz

import "net/http"

type Operation string

const (
	ReadSelf  Operation = "read_self"
	ReadAny   Operation = "read_any"
	WriteSelf Operation = "write_self"
	WriteAny  Operation = "write_any"
)

type Capability struct {
	ReadSelf  bool `json:"read_self"`
	ReadAny   bool `json:"read_any"`
	WriteSelf bool `json:"write_self"`
	WriteAny  bool `json:"write_any"`
}

func (cap Capability) Allowed(op Operation) bool {
	switch op {
	case ReadSelf:
		return cap.ReadSelf || cap.ReadAny
	case ReadAny:
		return cap.ReadAny
	case WriteSelf:
		return cap.WriteSelf || cap.WriteAny
	case WriteAny:
		return cap.WriteAny
	default:
		return false
	}
}

func MergeCapabilities(current Capability, next Capability) Capability {
	current.ReadSelf = current.ReadSelf || next.ReadSelf
	current.ReadAny = current.ReadAny || next.ReadAny
	current.WriteSelf = current.WriteSelf || next.WriteSelf
	current.WriteAny = current.WriteAny || next.WriteAny
	return current
}

func OperationForMethod(method string, readOp Operation, writeOp Operation) Operation {
	if method == http.MethodGet || method == http.MethodHead {
		return readOp
	}
	return writeOp
}
