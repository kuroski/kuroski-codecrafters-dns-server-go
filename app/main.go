package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

const (
	// Bit masks
	AA_Mask = 1 << 2 // 0x04 == 0000 0100
	TC_Mask = 1 << 1 // 0x02 == 0000 0010
	RD_Mask = 1      // 0x01 == 0000 0001

	// Shift and mask for OPCODE
	// OPCODE_Shift and OPCODE_Mask (bits 1-4 in byte 3)

	// to move bits from positions 4-7 to 0-3
	OPCODE_Shift = 3
	// to isolate bits 0000 1111 (binary)
	OPCODE_Mask = 0x0F
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:2053")
	if err != nil {
		fmt.Println("Failed to resolve UDP address:", err)
		return
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("Failed to bind to address:", err)
		return
	}
	defer udpConn.Close()

	buf := make([]byte, 512)

	for {
		size, source, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error receiving data:", err)
			break
		}

		header, err := parseDnsHeader(buf[:size])
		if err != nil {
			fmt.Println("Error parsing DNS header:", err)
			continue
		}

		response, err := buildResponse(header)
		if err != nil {
			fmt.Println("Error building response:", err)
			continue
		}

		_, err = udpConn.WriteToUDP(response, source)
		if err != nil {
			fmt.Println("Failed to send response:", err)
		}
	}
}

func parseDnsHeader(buf []byte) (*Header, error) {
	// The header section is always 12 bytes long
	if len(buf) < 12 {
		return nil, fmt.Errorf("buffer too short")
	}

	header := &Header{
		ID:      binary.BigEndian.Uint16(buf[0:2]), // read first 2 bytes from buffer (16 bits)
		QR:      buf[2]&0x80 != 0,                  // 0x80 == 1000 0000, checks the highest bit (7th bit)
		OPCODE:  (buf[2] >> OPCODE_Shift) & OPCODE_Mask,
		AA:      buf[2]&AA_Mask != 0,
		TC:      buf[2]&TC_Mask != 0,
		RD:      buf[2]&RD_Mask != 0,
		RA:      buf[3]&0x80 != 0,     // 0x80 == 1000 0000, checks the highest bit (7th bit)
		Z:       (buf[3] >> 4) & 0x07, // Z is 3 bits wide
		RCODE:   buf[3] & 0x0F,
		QDCOUNT: binary.BigEndian.Uint16(buf[4:6]),
		ANCOUNT: binary.BigEndian.Uint16(buf[6:8]),
		NSCOUNT: binary.BigEndian.Uint16(buf[8:10]),
		ARCOUNT: binary.BigEndian.Uint16(buf[10:12]),
	}

	return header, nil
}

func buildResponse(header *Header) ([]byte, error) {
	// Prepare the header for the response
	header.QR = true // Set QR to true for response

	// Create a new buffer and write the Header to it
	binaryBuf := new(bytes.Buffer)
	binary.Write(binaryBuf, binary.BigEndian, header.ID)

	// +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	// |QR|   Opcode  |AA|TC|RD|RA|   Z    |   RCODE   |
	// +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//  1        4     1  1  1  1    3        4
	// Prepare Flags (2 bytes)
	flags := uint16(0)
	// QR (1 bit): 1st bit (most significant bit)
	flags |= 0x8000 // QR = 1 (response)
	// Opcode (4 bits): 2nd-5th bits
	flags |= (uint16(header.OPCODE) & OPCODE_Mask) << 11
	// AA (1 bit): 6th bit
	flags |= boolToUint16(header.AA) << 10
	// TC (1 bit): 7th bit
	flags |= boolToUint16(header.TC) << 9
	// RD (1 bit): 8th bit
	flags |= boolToUint16(header.RD) << 8
	// RA (1 bit): 9th bit
	flags |= boolToUint16(header.RA) << 7
	// Z (3 bits): 10th-12th bits
	flags |= (uint16(header.Z) & 0x07) << 4
	// RCODE (4 bits): 13th-16th bits (least significant bits)
	flags |= uint16(header.RCODE) & 0x0F
	binary.Write(binaryBuf, binary.BigEndian, flags)

	binary.Write(binaryBuf, binary.BigEndian, header.QDCOUNT)
	binary.Write(binaryBuf, binary.BigEndian, header.ANCOUNT)
	binary.Write(binaryBuf, binary.BigEndian, header.NSCOUNT)
	binary.Write(binaryBuf, binary.BigEndian, header.ARCOUNT)

	return binaryBuf.Bytes(), nil
}

func boolToUint16(b bool) uint16 {
	if b {
		return 1
	}
	return 0
}
