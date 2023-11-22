package senml

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"testing"
)

type TestVector struct {
	format Format
	binary bool
	value  string
}

var testVectors = []TestVector{
	{JSON, false, "W3siYm4iOiJkZXYxMjMiLCJidCI6LTQ1LjY3LCJidSI6ImRlZ0MiLCJidmVyIjo1LCJuIjoidGVtcCIsInUiOiJkZWdDIiwidCI6LTEsInV0IjoxMCwidiI6MjIuMSwicyI6MH0seyJuIjoicm9vbSIsInQiOi0xLCJ2cyI6ImtpdGNoZW4ifSx7Im4iOiJkYXRhIiwidmQiOiJhYmMifSx7Im4iOiJvayIsInZiIjp0cnVlfV0="},
	{CBOR, true, "hKohZmRldjEyMyL7wEbVwo9cKPYjZGRlZ0MgBQBkdGVtcAFkZGVnQwb7v/AAAAAAAAAH+0AkAAAAAAAAAvtANhmZmZmZmgX7AAAAAAAAAACjAGRyb29tBvu/8AAAAAAAAANna2l0Y2hlbqIAZGRhdGEIY2FiY6IAYm9rBPU="},
	{XML, false, "PHNlbnNtbCB4bWxucz0idXJuOmlldGY6cGFyYW1zOnhtbDpuczpzZW5tbCI+PHNlbm1sIGJuPSJkZXYxMjMiIGJ0PSItNDUuNjciIGJ1PSJkZWdDIiBidmVyPSI1IiBuPSJ0ZW1wIiB1PSJkZWdDIiB0PSItMSIgdXQ9IjEwIiB2PSIyMi4xIiBzPSIwIj48L3Nlbm1sPjxzZW5tbCBuPSJyb29tIiB0PSItMSIgdnM9ImtpdGNoZW4iPjwvc2VubWw+PHNlbm1sIG49ImRhdGEiIHZkPSJhYmMiPjwvc2VubWw+PHNlbm1sIG49Im9rIiB2Yj0idHJ1ZSI+PC9zZW5tbD48L3NlbnNtbD4="},
}

func TestEncode(t *testing.T) {
	value := 22.1
	sum := 0.0
	vb := true
	vs := "kitchen"
	vd := "abc"
	s := Pack{
		Records: []Record{
			{BaseName: "dev123",
				BaseTime:    -45.67,
				BaseUnit:    "degC",
				BaseVersion: 5,
				Value:       &value, Unit: "degC", Name: "temp", Time: -1.0, UpdateTime: 10.0, Sum: &sum},
			{StringValue: &vs, Name: "room", Time: -1.0},
			{OpaqueValue: &vd, Name: "data"},
			{BoolValue: &vb, Name: "ok"},
		},
	}

	for i, vector := range testVectors {
		dataOut, err := Encode(s, vector.format)
		if err != nil {
			t.Fail()
		}

		if vector.binary {
			fmt.Print("Test Encode " + strconv.Itoa(i) + " got: ")
			fmt.Println(dataOut)
		} else {
			fmt.Println("Test Encode " + strconv.Itoa(i) + " got: " + string(dataOut))
		}

		//fmt.Println(base64.StdEncoding.EncodeToString(dataOut))

		if base64.StdEncoding.EncodeToString(dataOut) != vector.value {
			t.Error("Failed Encode for format " + strconv.Itoa(i) + " got: " + base64.StdEncoding.EncodeToString(dataOut))
		}
	}

}

func TestDecode(t *testing.T) {
	for i, vector := range testVectors {
		fmt.Println("Doing TestDecode for vector", i)

		data, err := base64.StdEncoding.DecodeString(vector.value)
		if err != nil {
			t.Fail()
		}

		s, err := Decode(data, vector.format)
		if err != nil {
			t.Fail()
		}

		dataOut, err := Encode(s, JSON)
		if err != nil {
			t.Fail()
		}

		fmt.Println("Test Decode " + strconv.Itoa(i) + " got: " + string(dataOut))
	}
}

func TestNormalize(t *testing.T) {
	value := 22.1
	sum := 0.0
	vb := true
	vs := "kitchen"
	vd := "abc"
	s := Pack{
		Records: []Record{
			{BaseName: "dev123/",
				BaseTime:    897845.67,
				BaseUnit:    "degC",
				BaseVersion: 5,
				Value:       &value, Unit: "degC", Name: "temp", Time: -1.0, UpdateTime: 10.0, Sum: &sum},
			{StringValue: &vs, Name: "room", Time: -1.0},
			{OpaqueValue: &vd, Name: "data"},
			{BoolValue: &vb, Name: "ok"},
		},
	}

	p, err := Normalize(s)

	dataOut, err := Encode(p, JSON)
	if err != nil {
		t.Fail()
	}
	fmt.Println("Test Normalize got: " + string(dataOut))

	fmt.Println(base64.StdEncoding.EncodeToString(dataOut))

	if base64.StdEncoding.EncodeToString(dataOut) != "WwogIHsiYnZlciI6NSwibiI6ImRldjEyMy90ZW1wIiwidSI6ImRlZ0MiLCJ0Ijo4OTc4NDQuNjcsInV0IjoxMCwidiI6MjIuMSwicyI6MH0sCiAgeyJidmVyIjo1LCJuIjoiZGV2MTIzL3Jvb20iLCJ1IjoiZGVnQyIsInQiOjg5Nzg0NC42NywidnMiOiJraXRjaGVuIn0sCiAgeyJidmVyIjo1LCJuIjoiZGV2MTIzL2RhdGEiLCJ1IjoiZGVnQyIsInQiOjg5Nzg0NS42NywidmQiOiJhYmMifSwKICB7ImJ2ZXIiOjUsIm4iOiJkZXYxMjMvb2siLCJ1IjoiZGVnQyIsInQiOjg5Nzg0NS42NywidmIiOnRydWV9Cl0K" {
		t.Error("Failed Normalize got: " + base64.StdEncoding.EncodeToString(dataOut))
	}
}

func TestBadInput1(t *testing.T) {
	data := []byte(" foo ")
	_, err := Decode(data, JSON)
	if err == nil {
		t.Fail()
	}
}

func TestBadInput2(t *testing.T) {
	data := []byte(" { \"n\":\"hi\" } ")
	_, err := Decode(data, JSON)
	if err == nil {
		t.Fail()
	}
}

func TestBadInputNoValue(t *testing.T) {
	data := []byte("  [ { \"n\":\"hi\" } ] ")
	_, err := Decode(data, JSON)
	if err == nil {
		t.Fail()
	}
}

func TestInputNumericName(t *testing.T) {
	data := []byte("  [ { \"n\":\"3a\", \"v\":1.0 } ] ")
	_, err := Decode(data, JSON)
	if err != nil {
		t.Fail()
	}
}

func TestBadInputNumericName(t *testing.T) {
	data := []byte("  [ { \"n\":\"-3b\", \"v\":1.0 } ] ")
	_, err := Decode(data, JSON)
	if err == nil {
		t.Fail()
	}
}

func TestInputWeirdName(t *testing.T) {
	data := []byte("  [ { \"n\":\"Az3-:./_\", \"v\":1.0 } ] ")
	_, err := Decode(data, JSON)
	if err != nil {
		t.Fail()
	}
}

func TestBadInputWeirdName(t *testing.T) {
	data := []byte("  [ { \"n\":\"A;b\", \"v\":1.0 } ] ")
	_, err := Decode(data, JSON)
	if err == nil {
		t.Fail()
	}
}

func TestInputWeirdBaseName(t *testing.T) {
	data := []byte("[ { \"bn\": \"a\" , \"n\":\"/b\" , \"v\":1.0} ] ")
	_, err := Decode(data, JSON)
	if err != nil {
		t.Fail()
	}
}

func TestBadInputNumericBaseName(t *testing.T) {
	data := []byte("[ { \"bn\": \"/3h\" , \"n\":\"i\" , \"v\":1.0} ] ")
	_, err := Decode(data, JSON)
	if err == nil {
		t.Fail()
	}
}

func TestInputSumOnly(t *testing.T) {
	data := []byte("[ { \"n\":\"a\", \"s\":1.0 } ] ")
	_, err := Decode(data, JSON)
	if err != nil {
		t.Fail()
	}
}

func TestInputBoolean(t *testing.T) {
	data := []byte("[ { \"n\":\"a\", \"vd\": \"aGkgCg\" } ] ")
	_, err := Decode(data, JSON)
	if err != nil {
		t.Fail()
	}
}

func TestInputData(t *testing.T) {
	data := []byte("  [ { \"n\":\"a\", \"vb\": true } ] ")
	_, err := Decode(data, JSON)
	if err != nil {
		t.Fail()
	}
}

func TestInputString(t *testing.T) {
	data := []byte("  [ { \"n\":\"a\", \"vs\": \"Hi\" } ] ")
	_, err := Decode(data, JSON)
	if err != nil {
		t.Fail()
	}
}
