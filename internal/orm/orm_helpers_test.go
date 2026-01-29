package orm

import (
	"reflect"
	"testing"
)

func TestInt64ListValueAndScan(t *testing.T) {
	original := Int64List{1, 2, 3}
	value, err := original.Value()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var fromString Int64List
	if err := fromString.Scan(value); err != nil {
		t.Fatalf("expected scan to succeed, got %v", err)
	}
	if !reflect.DeepEqual(original, fromString) {
		t.Fatalf("expected %v, got %v", original, fromString)
	}

	var fromBytes Int64List
	if err := fromBytes.Scan([]byte("[4,5]")); err != nil {
		t.Fatalf("expected scan to succeed, got %v", err)
	}
	if !reflect.DeepEqual(Int64List{4, 5}, fromBytes) {
		t.Fatalf("expected %v, got %v", Int64List{4, 5}, fromBytes)
	}

	fromBytes = Int64List{9}
	if err := fromBytes.Scan(nil); err != nil {
		t.Fatalf("expected scan to succeed, got %v", err)
	}
	if fromBytes != nil {
		t.Fatalf("expected nil list after scan, got %v", fromBytes)
	}
}

func TestInt64ListScanInvalidType(t *testing.T) {
	list := Int64List{}
	if err := list.Scan(42); err == nil {
		t.Fatalf("expected error for unsupported type")
	}
}

func TestDealiasResource(t *testing.T) {
	alias := uint32(14)
	DealiasResource(&alias)
	if alias != 4 {
		t.Fatalf("expected 4, got %d", alias)
	}

	unknown := uint32(99)
	DealiasResource(&unknown)
	if unknown != 99 {
		t.Fatalf("expected 99, got %d", unknown)
	}
}

func TestToInt64List(t *testing.T) {
	list := ToInt64List([]uint32{1, 2, 3})
	if !reflect.DeepEqual(list, Int64List{1, 2, 3}) {
		t.Fatalf("expected [1 2 3], got %v", list)
	}
	if empty := ToInt64List(nil); len(empty) != 0 {
		t.Fatalf("expected empty list, got %v", empty)
	}
}

func TestToUint32List(t *testing.T) {
	list := ToUint32List(Int64List{4, 5})
	if !reflect.DeepEqual(list, []uint32{4, 5}) {
		t.Fatalf("expected [4 5], got %v", list)
	}
	if empty := ToUint32List(nil); len(empty) != 0 {
		t.Fatalf("expected empty list, got %v", empty)
	}
}
