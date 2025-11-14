package commhandler

import (	
	"google.golang.org/protobuf/proto"

	packet "helios/generated/go/transport"
)

func CreateTrackportPacket(id int32, address string, data []byte) (*packet.TransportPacket, error) {
	pkt := &packet.TransportPacket{
		Id:      id,
		Address: address,
		Data:    data,
	}
	return pkt, nil
}

func MarshalTrackportPacket(pkt *packet.TransportPacket) ([]byte, error) {
	data, err := proto.Marshal(pkt)
	if err != nil {
		return nil, err
	}
	return data, nil
}