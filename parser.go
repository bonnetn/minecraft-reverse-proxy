package main

import (
	"fmt"
	"io"
)

/*
public static int readVarInt() {
    int numRead = 0;
    int result = 0;
    byte read;
    do {
        read = readByte();
        int value = (read & 0b01111111);
        result |= (value << (7 * numRead));

        numRead++;
        if (numRead > 5) {
            throw new RuntimeException("VarInt is too big");
        }
    } while ((read & 0b10000000) != 0);

    return result;
}
public static long readVarLong() {
    int numRead = 0;
    long result = 0;
    byte read;
    do {
        read = readByte();
        int value = (read & 0b01111111);
        result |= (value << (7 * numRead));

        numRead++;
        if (numRead > 10) {
            throw new RuntimeException("VarLong is too big");
        }
    } while ((read & 0b10000000) != 0);

    return result;
}
*/

func readByte(reader io.Reader) (byte, error) {
	var buf [1]byte
	_, err := reader.Read(buf[:])
	if err != nil {
		return 0, err
	}
	return buf[0], nil
}

func readVarInt(reader io.Reader) (int32, error) {
	var numRead int32
	var result int32
	for {
		readByte, err := readByte(reader)
		if err != nil {
			return 0, err
		}
		value := int32(readByte & 0b01111111)
		result |= value << (7 * numRead)

		numRead++
		if numRead > 5 {
			return 0, fmt.Errorf("VarInt is too big")
		}
		if (readByte & 0b10000000) == 0 {
			break
		}
	}
	return result, nil

}

type Packet struct {
	Length          int32
	PacketID        int32
	ProtocolVersion int32
	ServerAddress   string
	ServerPort      int32
	NextState       int32
}

func getNextPacket(reader io.Reader) (Packet, error) {
	length, err := readVarInt(reader)
	if err != nil {
		return Packet{}, fmt.Errorf("failed to read length: %w", err)
	}
	packetID, err := readVarInt(reader)
	if err != nil {
		return Packet{}, fmt.Errorf("failed to read length: %w", err)
	}

	if packetID != 0 {
		return Packet{}, fmt.Errorf("packetID is not 0")
	}

	protocolVersion, err := readVarInt(reader)
	if err != nil {
		return Packet{}, fmt.Errorf("failed to read protocolVersion: %w", err)
	}

	serverAddress, err := readString(reader)
	if err != nil {
		return Packet{}, fmt.Errorf("failed to read serverAddress: %w", err)
	}

	serverPort, err := readUnsignedShort(reader)
	if err != nil {
		return Packet{}, fmt.Errorf("failed to read serverPort: %w", err)
	}

	nextState, err := readVarInt(reader)
	if err != nil {
		return Packet{}, fmt.Errorf("failed to read nextState: %w", err)
	}

	return Packet{
		Length:          length,
		PacketID:        packetID,
		ProtocolVersion: protocolVersion,
		ServerAddress:   serverAddress,
		ServerPort:      serverPort,
		NextState:       nextState,
	}, nil
}

func readUnsignedShort(reader io.Reader) (int32, error) {
	var buf [2]byte
	_, err := reader.Read(buf[:])
	if err != nil {
		return 0, err
	}
	return int32(buf[0])<<8 + int32(buf[1]), nil
}

func readString(reader io.Reader) (string, error) {
	length, err := readVarInt(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read length: %w", err)
	}

	var buf = make([]byte, length)
	_, err = reader.Read(buf)
	if err != nil {
		return "", fmt.Errorf("failed to read string: %w", err)
	}

	return string(buf), nil
}
