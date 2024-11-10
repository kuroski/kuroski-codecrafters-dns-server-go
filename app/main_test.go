package main

import (
	"testing"
)

func TestParseDNSQuestions(t *testing.T) {
	data := []byte{16, 129, 1, 0, 0, 2, 0, 0, 0, 0, 0, 0, 3, 97, 98, 99, 17, 108, 111, 110, 103, 97, 115, 115, 100, 111, 109, 97, 105, 110, 110, 97, 109, 101, 3, 99, 111, 109, 0, 0, 1, 0, 1, 3, 100, 101, 102, 192, 16, 0, 1, 0, 1}

	var header DNSHeader
	header.Parse(data[:12]) // The first 12 bytes are the header

	t.Logf("Parsed DNS header: %+v", header)

	questions, err := parseDNSQuestions(data[12:], header) // Parse the questions from the remaining bytes
	if err != nil {
		t.Fatalf("Failed to parse DNS question: %v", err)
	}

	t.Logf("Parsed DNS questions: %+v", questions)

	if len(questions) != 2 {
		t.Errorf("Expected 2 questions, but got %d", len(questions))
	}
	if questions[0].Name != "abc.longassdomainname.com" {
		t.Errorf("Expected question 0 name to be abc.longassdomainname.com, but got %s", questions[0].Name)
	}
	if questions[1].Name != "def.longassdomainname.com" { // Since it's a pointer, it should resolve to the same base part
		t.Errorf("Expected question 1 name to be def.longassdomainname.com, but got %s", questions[1].Name)
	}
}
