package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
)

const (
	packetID            = 1234
	queryResponse       = 1 // QR bit indicates a query (0) or response (1)
	opcode              = 0 // OPCODE bits for standard query
	authoritativeAnswer = 0 // AA bit for authoritative answer (0: non-authoritative)
	truncation          = 0 // TC bit for truncation (0: not truncated)
	recursionDesired    = 0 // RD bit for recursion desired (0: do not ask recursive query)
	recursionAvailable  = 0 // RA bit for recursion available (0: recursion not available)
	responseCode        = 0 // RCODE bits for response code (0: no error)
	questionCount       = 0 // QDCOUNT for the number of question entries (0 for this example)
	answerCount         = 0 // ANCOUNT for the number of answer entries (0 for this example)
	authorityCount      = 0 // NSCOUNT for the number of authority records (0 for this example)
	additionalCount     = 0 // ARCOUNT for the number of additional records (0 for this example)
)

// DNSHeader represents a DNS message header
type DNSHeader struct {
	ID      uint16 // Packet Identifier (16 bits)
	Flags   uint16 // Flags (16 bits) - contains multiple sub-fields
	QDCOUNT uint16 // Number of Question Records
	ANCOUNT uint16 // Number of Answer Records
	NSCOUNT uint16 // Number of Authority Records
	ARCOUNT uint16 // Number of Additional Records
}

// Serialize serializes the DNSHeader into a byte array
func (h *DNSHeader) Serialize() []byte {
	buf := make([]byte, 12) // DNS header is always 12 bytes long
	binary.BigEndian.PutUint16(buf[0:2], h.ID)
	binary.BigEndian.PutUint16(buf[2:4], h.Flags)
	binary.BigEndian.PutUint16(buf[4:6], h.QDCOUNT)
	binary.BigEndian.PutUint16(buf[6:8], h.ANCOUNT)
	binary.BigEndian.PutUint16(buf[8:10], h.NSCOUNT)
	binary.BigEndian.PutUint16(buf[10:12], h.ARCOUNT)
	return buf
}

// Create a new DNS reply message based on the specified values
func createDNSReply() []byte {
	// Construct the 16-bit Flags field
	// | QR  | OPCODE |  AA | TC | RD | RA | Z   | RCODE |
	// |  1  | 0000   |  1  |  0 |  0 |  0 | 000 | 0000  |
	// ---------------------------------------------------
	//  16-15  14-11    10    9    8    7    6-4   3-0
	// ---------------------------------------------------
	// QR = 1
	// OPCODE = 0 (0000)
	// AA = 1
	// TC = 0
	// RD = 0
	// RA = 0
	// Z = 0 (000)
	// RCODE = 0 (0000)
	// ---------------------------------------------------
	// 1000 0000 0000 0000  (QR << 15)
	// OR 0000 0000 0000 0000  (OPCODE << 11)
	// OR 0000 0100 0000 0000  (AA << 10)
	// OR 0000 0000 0000 0000  (TC << 9)
	// OR 0000 0000 0000 0000  (RD << 8)
	// OR 0000 0000 0000 0000  (RA << 7)
	// OR 0000 0000 0000 0000  (RCODE)
	// = 1000 0100 0000 0000 (combined)
	flags := (queryResponse << 15) | // QR bit (1 bit)
		(opcode << 11) | // OPCODE (4 bits)
		(authoritativeAnswer << 10) | // AA bit (1 bit)
		(truncation << 9) | // TC bit (1 bit)
		(recursionDesired << 8) | // RD bit (1 bit)
		(recursionAvailable << 7) | // RA bit (1 bit)
		(responseCode) // RCODE (4 bits)

	header := &DNSHeader{
		ID:      packetID,
		Flags:   uint16(flags),
		QDCOUNT: questionCount,
		ANCOUNT: answerCount,
		NSCOUNT: authorityCount,
		ARCOUNT: additionalCount,
	}
	return header.Serialize()
}

func handleDNSRequest(conn *net.UDPConn, addr *net.UDPAddr, data []byte) {
	// Log the received packet
	log.Printf("Received DNS query from %s", addr.String())

	// Generate DNS reply
	reply := createDNSReply()
	_, err := conn.WriteToUDP(reply, addr)
	if err != nil {
		log.Printf("Failed to send DNS reply: %v", err)
		return
	}

	log.Printf("Sent DNS reply to %s", addr.String())
}

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

	for {
		buf := make([]byte, 512) // DNS messages are usually limited to 512 bytes
		n, addr, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("Failed to read UDP packet: %v", err)
			continue
		}

		go handleDNSRequest(udpConn, addr, buf[:n])
	}
}
