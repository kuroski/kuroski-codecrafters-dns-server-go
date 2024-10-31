package main

// Header
//   - ID (16 bit): This is 2 bytes (buf[0] and buf[1]).
//   - QR (1 bit): This is the 1st bit in the 3rd byte (buf[2]).
//   - OPCODE (4 bits): Bits 1-4 in the 3rd byte (buf[2]).
//   - AA (1 bit): This is the 5th bit in the 3rd byte (buf[2]).
//   - TC (1 bit): This is the 6th bit in the 3rd byte (buf[2]).
//   - RD (1 bit): This is the 7th bit in the 3rd byte (buf[2]).
//   - RA (1 bit): This is the 1st bit in the 4th byte (buf[3]).
//   - Z (3 bits): Bits 4-6 in the 4th byte (buf[3]).
//   - RCODE (4 bits): Bits 7-10 in the 4th byte (buf[3]).
type Header struct {
	ID      uint16 // packet Identifier (16 bits)
	QR      bool   // query/Response Indicator (1 bit)
	OPCODE  uint8  // operation Code (4 bits)
	AA      bool   // authoritative Answer (1 bit)
	TC      bool   // truncation (1 bit)
	RD      bool   // recursion desired (1 bit)
	RA      bool   // recursion Available (1 bit)
	Z       uint8  // reserved (1 bit)
	RCODE   uint8  // response code (4 bits)
	QDCOUNT uint16 // question count (16 bits)
	ANCOUNT uint16 // answer record count (16 bits)
	NSCOUNT uint16 // authority record count (16 bits)
	ARCOUNT uint16 // additional record count (16 bits)
}
