package packets

const (
	UPDATE_CHECK_PACKET = 10800
	HTTP_GET_SERVERS    = 8239 // Non-conventional packet, GET HTTP request / response
	AUTH_PACKET         = 10021
)

/*
Packet body -> packer.lua :
  - 2 bytes (uint16) : packet size
  - 1 byte (uint8)   : 0x00
  - 2 bytes (uint16) : packet id
  - 2 bytes (uint16) : packet index (always 0x0000 for some reason, seems unused), is 0x0001 if frame has more than 1 packet
  - rest    : content
*/
func GetPacketId(offset int, buffer *[]byte) int {
	var id int
	id = int((*buffer)[3+offset]) << 8
	id += int((*buffer)[4+offset])
	return id
}

func GetPacketSize(offset int, buffer *[]byte) int {
	var size int
	size = int((*buffer)[0+offset]) << 8
	size += int((*buffer)[1+offset])
	return size
}

func GetPacketIndex(offset int, buffer *[]byte) int {
	var index int
	index = int((*buffer)[5+offset]) << 8
	index += int((*buffer)[6+offset])
	return index
}
