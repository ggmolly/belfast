package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

var validSC60000 protobuf.SC_60000

func CommanderGuildData(buffer *[]byte, client *connection.Client) (int, int, error) {
	validSC60000.Guild.Base.Id = proto.Uint32(0)

	return client.SendMessage(60000, &validSC60000)
}

func init() {
	data := []byte{0x0a, 0xfc, 0x08, 0x0a, 0x45, 0x08, 0xb4, 0x8c, 0x81, 0x20, 0x10, 0x02, 0x18, 0x02, 0x22, 0x10, 0x61, 0x72, 0x61, 0x63, 0x68, 0x69, 0x64, 0x65, 0x20, 0x65, 0x78, 0x74, 0x72, 0x65, 0x6d, 0x65, 0x28, 0x05, 0x32, 0x0b, 0x61, 0x7a, 0x7a, 0x20, 0x65, 0x73, 0x74, 0x20, 0x67, 0x61, 0x79, 0x3a, 0x10, 0x6e, 0x6f, 0x69, 0x73, 0x65, 0x74, 0x74, 0x65, 0x20, 0x65, 0x78, 0x74, 0x72, 0x65, 0x6d, 0x65, 0x40, 0xc0, 0x2c, 0x48, 0x02, 0x50, 0x00, 0x58, 0x00, 0x12, 0x3e, 0x08, 0xe3, 0x23, 0x10, 0x02, 0x18, 0x8a, 0x9f, 0xd8, 0x22, 0x22, 0x07, 0x41, 0x7a, 0x7a, 0x6e, 0x65, 0x6b, 0x6f, 0x28, 0x3d, 0x32, 0x00, 0x38, 0x00, 0x40, 0xf8, 0xb8, 0xe1, 0x96, 0x06, 0x4a, 0x17, 0x08, 0xd2, 0xbe, 0x18, 0x10, 0xd6, 0xbe, 0x18, 0x18, 0xb7, 0x02, 0x20, 0x00, 0x28, 0x00, 0x30, 0x88, 0xdf, 0x9a, 0x92, 0x06, 0x38, 0x01, 0x60, 0xfb, 0xc4, 0xf3, 0x91, 0x06, 0x12, 0x3d, 0x08, 0xc5, 0x8d, 0x01, 0x10, 0x01, 0x18, 0x9e, 0xa1, 0xd8, 0x22, 0x22, 0x09, 0x4e, 0x75, 0x74, 0x74, 0x79, 0x31, 0x33, 0x33, 0x37, 0x28, 0x5a, 0x32, 0x00, 0x38, 0x01, 0x40, 0xf3, 0x9f, 0xd1, 0xab, 0x06, 0x4a, 0x13, 0x08, 0xf6, 0x9d, 0x06, 0x10, 0xf5, 0x9d, 0x06, 0x18, 0xc0, 0x02, 0x20, 0x00, 0x28, 0x00, 0x30, 0x00, 0x38, 0x00, 0x60, 0xd6, 0xc3, 0xf3, 0x91, 0x06, 0x1a, 0x1c, 0x08, 0x05, 0x10, 0xc2, 0xe0, 0xdd, 0xa4, 0x06, 0x18, 0x9e, 0xa1, 0xd8, 0x22, 0x22, 0x09, 0x4e, 0x75, 0x74, 0x74, 0x79, 0x31, 0x33, 0x33, 0x37, 0x28, 0xd3, 0xb5, 0x06, 0x1a, 0x1c, 0x08, 0x05, 0x10, 0xcb, 0x8f, 0xcd, 0xa4, 0x06, 0x18, 0x9e, 0xa1, 0xd8, 0x22, 0x22, 0x09, 0x4e, 0x75, 0x74, 0x74, 0x79, 0x31, 0x33, 0x33, 0x37, 0x28, 0xd3, 0xb5, 0x06, 0x1a, 0x1c, 0x08, 0x05, 0x10, 0xc5, 0xc7, 0xcc, 0xa4, 0x06, 0x18, 0x9e, 0xa1, 0xd8, 0x22, 0x22, 0x09, 0x4e, 0x75, 0x74, 0x74, 0x79, 0x31, 0x33, 0x33, 0x37, 0x28, 0xfb, 0xc1, 0x0c, 0x1a, 0x1c, 0x08, 0x05, 0x10, 0xa5, 0xf8, 0x97, 0xa0, 0x06, 0x18, 0x9e, 0xa1, 0xd8, 0x22, 0x22, 0x09, 0x4e, 0x75, 0x74, 0x74, 0x79, 0x31, 0x33, 0x33, 0x37, 0x28, 0xcb, 0xf3, 0x18, 0x1a, 0x1c, 0x08, 0x05, 0x10, 0xb8, 0x97, 0xfc, 0x9f, 0x06, 0x18, 0x9e, 0xa1, 0xd8, 0x22, 0x22, 0x09, 0x4e, 0x75, 0x74, 0x74, 0x79, 0x31, 0x33, 0x33, 0x37, 0x28, 0xdb, 0xac, 0x0c, 0x1a, 0x1c, 0x08, 0x05, 0x10, 0xa0, 0xa6, 0xf3, 0x9f, 0x06, 0x18, 0x9e, 0xa1, 0xd8, 0x22, 0x22, 0x09, 0x4e, 0x75, 0x74, 0x74, 0x79, 0x31, 0x33, 0x33, 0x37, 0x28, 0x97, 0xd7, 0x12, 0x1a, 0x1c, 0x08, 0x05, 0x10, 0x84, 0xc1, 0xde, 0x9f, 0x06, 0x18, 0x9e, 0xa1, 0xd8, 0x22, 0x22, 0x09, 0x4e, 0x75, 0x74, 0x74, 0x79, 0x31, 0x33, 0x33, 0x37, 0x28, 0xdb, 0xac, 0x0c, 0x1a, 0x1c, 0x08, 0x05, 0x10, 0xed, 0xc0, 0xde, 0x9f, 0x06, 0x18, 0x9e, 0xa1, 0xd8, 0x22, 0x22, 0x09, 0x4e, 0x75, 0x74, 0x74, 0x79, 0x31, 0x33, 0x33, 0x37, 0x28, 0xdf, 0xd1, 0x0c, 0x1a, 0x1c, 0x08, 0x05, 0x10, 0xaa, 0xc0, 0xde, 0x9f, 0x06, 0x18, 0x9e, 0xa1, 0xd8, 0x22, 0x22, 0x09, 0x4e, 0x75, 0x74, 0x74, 0x79, 0x31, 0x33, 0x33, 0x37, 0x28, 0xd5, 0xc2, 0x0c, 0x1a, 0x1c, 0x08, 0x05, 0x10, 0xde, 0xbf, 0xde, 0x9f, 0x06, 0x18, 0x9e, 0xa1, 0xd8, 0x22, 0x22, 0x09, 0x4e, 0x75, 0x74, 0x74, 0x79, 0x31, 0x33, 0x33, 0x37, 0x28, 0xdb, 0xac, 0x0c, 0x1a, 0x1c, 0x08, 0x05, 0x10, 0xf4, 0xf1, 0x87, 0x9f, 0x06, 0x18, 0x9e, 0xa1, 0xd8, 0x22, 0x22, 0x09, 0x4e, 0x75, 0x74, 0x74, 0x79, 0x31, 0x33, 0x33, 0x37, 0x28, 0xb3, 0xc7, 0x12, 0x1a, 0x1c, 0x08, 0x05, 0x10, 0xba, 0xdb, 0xc3, 0x9e, 0x06, 0x18, 0x9e, 0xa1, 0xd8, 0x22, 0x22, 0x09, 0x4e, 0x75, 0x74, 0x74, 0x79, 0x31, 0x33, 0x33, 0x37, 0x28, 0xa9, 0xac, 0x0c, 0x1a, 0x1c, 0x08, 0x05, 0x10, 0xcd, 0xb5, 0xb9, 0x9e, 0x06, 0x18, 0x9e, 0xa1, 0xd8, 0x22, 0x22, 0x09, 0x4e, 0x75, 0x74, 0x74, 0x79, 0x31, 0x33, 0x33, 0x37, 0x28, 0x89, 0xab, 0x0c, 0x1a, 0x1c, 0x08, 0x05, 0x10, 0xc7, 0xf6, 0xb3, 0x9e, 0x06, 0x18, 0x9e, 0xa1, 0xd8, 0x22, 0x22, 0x09, 0x4e, 0x75, 0x74, 0x74, 0x79, 0x31, 0x33, 0x33, 0x37, 0x28, 0xe3, 0xec, 0x2a, 0x1a, 0x1c, 0x08, 0x05, 0x10, 0xbb, 0xd9, 0xa4, 0x9e, 0x06, 0x18, 0x9e, 0xa1, 0xd8, 0x22, 0x22, 0x09, 0x4e, 0x75, 0x74, 0x74, 0x79, 0x31, 0x33, 0x33, 0x37, 0x28, 0xf1, 0xb5, 0x06, 0x1a, 0x1c, 0x08, 0x05, 0x10, 0x96, 0xc8, 0xa4, 0x9e, 0x06, 0x18, 0x9e, 0xa1, 0xd8, 0x22, 0x22, 0x09, 0x4e, 0x75, 0x74, 0x74, 0x79, 0x31, 0x33, 0x33, 0x37, 0x28, 0xb3, 0xb3, 0x12, 0x1a, 0x1c, 0x08, 0x05, 0x10, 0xb5, 0xa9, 0xa2, 0x9e, 0x06, 0x18, 0x9e, 0xa1, 0xd8, 0x22, 0x22, 0x09, 0x4e, 0x75, 0x74, 0x74, 0x79, 0x31, 0x33, 0x33, 0x37, 0x28, 0xe3, 0xec, 0x2a, 0x1a, 0x1c, 0x08, 0x05, 0x10, 0xe4, 0xa2, 0x97, 0x9e, 0x06, 0x18, 0x9e, 0xa1, 0xd8, 0x22, 0x22, 0x09, 0x4e, 0x75, 0x74, 0x74, 0x79, 0x31, 0x33, 0x33, 0x37, 0x28, 0x8d, 0xe7, 0x24, 0x1a, 0x1c, 0x08, 0x05, 0x10, 0xad, 0xd0, 0x84, 0x9e, 0x06, 0x18, 0x9e, 0xa1, 0xd8, 0x22, 0x22, 0x09, 0x4e, 0x75, 0x74, 0x74, 0x79, 0x31, 0x33, 0x33, 0x37, 0x28, 0xf1, 0xb5, 0x06, 0x1a, 0x1c, 0x08, 0x05, 0x10, 0xba, 0xa6, 0x80, 0x9e, 0x06, 0x18, 0x9e, 0xa1, 0xd8, 0x22, 0x22, 0x09, 0x4e, 0x75, 0x74, 0x74, 0x79, 0x31, 0x33, 0x33, 0x37, 0x28, 0xb3, 0x9e, 0x37, 0x1a, 0x1c, 0x08, 0x05, 0x10, 0xa3, 0x81, 0xfa, 0x9d, 0x06, 0x18, 0x9e, 0xa1, 0xd8, 0x22, 0x22, 0x09, 0x4e, 0x75, 0x74, 0x74, 0x79, 0x31, 0x33, 0x33, 0x37, 0x28, 0x9b, 0xcf, 0x12, 0x1a, 0x1c, 0x08, 0x05, 0x10, 0xf4, 0xaf, 0xef, 0x9d, 0x06, 0x18, 0x9e, 0xa1, 0xd8, 0x22, 0x22, 0x09, 0x4e, 0x75, 0x74, 0x74, 0x79, 0x31, 0x33, 0x33, 0x37, 0x28, 0xb9, 0xcf, 0x12, 0x1a, 0x1c, 0x08, 0x05, 0x10, 0xc3, 0xbb, 0xe5, 0x9d, 0x06, 0x18, 0x9e, 0xa1, 0xd8, 0x22, 0x22, 0x09, 0x4e, 0x75, 0x74, 0x74, 0x79, 0x31, 0x33, 0x33, 0x37, 0x28, 0xf1, 0xb5, 0x06, 0x1a, 0x1c, 0x08, 0x05, 0x10, 0xdd, 0x8c, 0xe0, 0x9d, 0x06, 0x18, 0x9e, 0xa1, 0xd8, 0x22, 0x22, 0x09, 0x4e, 0x75, 0x74, 0x74, 0x79, 0x31, 0x33, 0x33, 0x37, 0x28, 0xb3, 0xc7, 0x12, 0x1a, 0x1c, 0x08, 0x05, 0x10, 0xf6, 0xdf, 0xdf, 0x9d, 0x06, 0x18, 0x9e, 0xa1, 0xd8, 0x22, 0x22, 0x09, 0x4e, 0x75, 0x74, 0x74, 0x79, 0x31, 0x33, 0x33, 0x37, 0x28, 0xd7, 0xcc, 0x18, 0x1a, 0x1c, 0x08, 0x05, 0x10, 0xa8, 0xca, 0xdf, 0x9d, 0x06, 0x18, 0x9e, 0xa1, 0xd8, 0x22, 0x22, 0x09, 0x4e, 0x75, 0x74, 0x74, 0x79, 0x31, 0x33, 0x33, 0x37, 0x28, 0xdd, 0xc1, 0x0c, 0x1a, 0x1c, 0x08, 0x05, 0x10, 0xf0, 0xac, 0xdc, 0x9d, 0x06, 0x18, 0x9e, 0xa1, 0xd8, 0x22, 0x22, 0x09, 0x4e, 0x75, 0x74, 0x74, 0x79, 0x31, 0x33, 0x33, 0x37, 0x28, 0xd3, 0xac, 0x13, 0x1a, 0x1c, 0x08, 0x05, 0x10, 0xa5, 0x83, 0xdc, 0x9d, 0x06, 0x18, 0x9e, 0xa1, 0xd8, 0x22, 0x22, 0x09, 0x4e, 0x75, 0x74, 0x74, 0x79, 0x31, 0x33, 0x33, 0x37, 0x28, 0xfb, 0xe4, 0x2a, 0x1a, 0x1c, 0x08, 0x05, 0x10, 0x8b, 0x83, 0xdc, 0x9d, 0x06, 0x18, 0x9e, 0xa1, 0xd8, 0x22, 0x22, 0x09, 0x4e, 0x75, 0x74, 0x74, 0x79, 0x31, 0x33, 0x33, 0x37, 0x28, 0xfb, 0xe4, 0x2a, 0x1a, 0x1c, 0x08, 0x05, 0x10, 0xcb, 0x82, 0xdc, 0x9d, 0x06, 0x18, 0x9e, 0xa1, 0xd8, 0x22, 0x22, 0x09, 0x4e, 0x75, 0x74, 0x74, 0x79, 0x31, 0x33, 0x33, 0x37, 0x28, 0xd7, 0xcc, 0x18, 0x22, 0x30, 0x08, 0xb5, 0x07, 0x12, 0x06, 0x08, 0x00, 0x10, 0x00, 0x18, 0x00, 0x18, 0xef, 0xf3, 0x9d, 0x97, 0x06, 0x22, 0x08, 0x08, 0x88, 0x27, 0x10, 0x01, 0x18, 0xde, 0x01, 0x22, 0x07, 0x08, 0xed, 0x07, 0x10, 0x00, 0x18, 0x42, 0x28, 0x00, 0x30, 0x00, 0x38, 0xef, 0x89, 0xd4, 0x96, 0x06, 0x40, 0x00}
	proto.Unmarshal(data, &validSC60000)
}