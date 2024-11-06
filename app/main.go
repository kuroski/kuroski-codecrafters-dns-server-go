package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"strings"
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
	questionCount       = 1 // QDCOUNT for the number of question entries (0 for this example)
	answerCount         = 1 // ANCOUNT for the number of answer entries (0 for this example)
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

type DNSQuestion struct {
	Name  string
	Type  uint16
	Class uint16
}

func (q *DNSQuestion) Serialize() []byte {
	var buf []byte
	labels := strings.Split(q.Name, ".")
	for _, label := range labels {
		buf = append(buf, byte(len(label)))
		buf = append(buf, []byte(label)...)
	}
	buf = append(buf, 0) // end of the Name

	qType := make([]byte, 2)
	binary.BigEndian.PutUint16(qType, q.Type)
	buf = append(buf, qType...)

	class := make([]byte, 2)
	binary.BigEndian.PutUint16(class, q.Class)
	buf = append(buf, class...)

	return buf
}

type DNSAnswer struct {
	Name     string
	Type     uint16
	Class    uint16
	TTL      uint32
	RDLength uint16
	RData    []byte
}

func (a *DNSAnswer) Serialize() []byte {
	var buf []byte

	labels := strings.Split(a.Name, ".")
	for _, label := range labels {
		buf = append(buf, byte(len(label)))
		buf = append(buf, []byte(label)...)
	}
	buf = append(buf, 0) // end of the Name

	typeBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(typeBytes, a.Type)
	buf = append(buf, typeBytes...)

	classBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(classBytes, a.Class)
	buf = append(buf, classBytes...)

	ttlBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(ttlBytes, a.TTL)
	buf = append(buf, ttlBytes...)

	rdLengthBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(rdLengthBytes, a.RDLength)
	buf = append(buf, rdLengthBytes...)

	buf = append(buf, a.RData...)

	return buf
}

// Create a new DNS reply message based on the specified values
func createDNSReply(question DNSQuestion, answer DNSAnswer) []byte {
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

	return append(append(header.Serialize(), question.Serialize()...), answer.Serialize()...)
}

func parseDNSQuestion(data []byte) (DNSQuestion, error) {
	var question DNSQuestion

	// Decode Name
	var name []string
	for {
		labelSize := int(data[0])
		if labelSize == 0 {
			break
		}

		name = append(name, string(data[1:labelSize+1]))
		data = data[labelSize+1:]
	}
	question.Name = strings.Join(name, ".")

	// Consume the 0 byte - \x00 is the null byte that terminates the domain name
	data = data[1:]

	// Decode Type and Class
	if len(data) < 4 {
		return question, fmt.Errorf("invalid question format")
	}
	question.Type = binary.BigEndian.Uint16(data[:2])
	data = data[2:]
	question.Class = binary.BigEndian.Uint16(data[:2])

	return question, nil
}

func handleDNSRequest(conn *net.UDPConn, addr *net.UDPAddr, data []byte) {
	// Log the received packet
	log.Printf("Received DNS query from %s", addr.String())

	// Parse the incoming DNS question
	question, err := parseDNSQuestion(data[12:]) // Skip the first 12 bytes (DNS header)
	if err != nil {
		log.Printf("Failed to parse DNS question: %v", err)
		return
	}

	// Construct a sample answer
	answer := DNSAnswer{
		Name:     question.Name,
		Type:     1, // A record
		Class:    1, // IN (Internet)
		TTL:      60,
		RDLength: 4,
		RData:    []byte{8, 8, 8, 8},
	}

	// Generate DNS reply
	reply := createDNSReply(question, answer)
	_, err = conn.WriteToUDP(reply, addr)
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
