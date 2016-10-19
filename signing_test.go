package matrixfederation

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"golang.org/x/crypto/ed25519"
	"testing"
)

func TestVerifyJSON(t *testing.T) {
	// Check JSON verification using the test vectors from https://matrix.org/docs/spec/appendices.html
	seed, err := base64.RawStdEncoding.DecodeString("YJDBA9Xnr2sVqXD9Vj7XVUnmFZcZrlw8Md7kMW+3XA1")
	if err != nil {
		t.Fatal(err)
	}
	random := bytes.NewBuffer(seed)
	entityName := "domain"
	keyID := "ed25519:1"

	publicKey, _, err := ed25519.GenerateKey(random)
	if err != nil {
		t.Fatal(err)
	}

	testVerifyOK := func(input string) {
		err := VerifyJSON(entityName, keyID, publicKey, []byte(input))
		if err != nil {
			t.Fatal(err)
		}
	}

	testVerifyNotOK := func(reason, input string) {
		err := VerifyJSON(entityName, keyID, publicKey, []byte(input))
		if err == nil {
			t.Fatal("Expected VerifyJSON to fail for input %v because %v", input, reason)
		}
	}

	testVerifyOK(`{
		"signatures": {
			"domain": {
				"ed25519:1": "K8280/U9SSy9IVtjBuVeLr+HpOB4BQFWbg+UZaADMtTdGYI7Geitb76LTrr5QV/7Xg4ahLwYGYZzuHGZKM5ZAQ"
			}
		}
	}`)

	testVerifyNotOK("the json is modified", `{
		"a new key": "a new value",
		"signatures": {
			"domain": {
				"ed25519:1": "K8280/U9SSy9IVtjBuVeLr+HpOB4BQFWbg+UZaADMtTdGYI7Geitb76LTrr5QV/7Xg4ahLwYGYZzuHGZKM5ZAQ"
			}
		}
	}`)
	testVerifyNotOK("the signature is modified", `{
		"a new key": "a new value",
		"signatures": {
			"domain": {
				"ed25519:1": "modifiedSSy9IVtjBuVeLr+HpOB4BQFWbg+UZaADMtTdGYI7Geitb76LTrr5QV/7Xg4ahLwYGYZzuHGZKM5ZAQ"
			}
		}
	}`)
	testVerifyNotOK("there are no signatures", `{}`)
	testVerifyNotOK("there are no signatures", `{"signatures": {}}`)

	testVerifyNotOK("there are not signatures for domain", `{
		"signatures": {"domain": {}}
	}`)
	testVerifyNotOK("the signature has the wrong key_id", `{
		"signatures": { "domain": {
			"ed25519:2":"KqmLSbO39/Bzb0QIYE82zqLwsA+PDzYIpIRA2sRQ4sL53+sN6/fpNSoqE7BP7vBZhG6kYdD13EIMJpvhJI+6Bw"
		}}
	}`)
	testVerifyNotOK("the signature is too short for ed25519", `{"signatures": {"domain": {"ed25519:1":"not/a/valid/signature"}}}`)
	testVerifyNotOK("the signature has base64 padding that it shouldn't have", `{
		"signatures": { "domain": {
			"ed25519:1": "K8280/U9SSy9IVtjBuVeLr+HpOB4BQFWbg+UZaADMtTdGYI7Geitb76LTrr5QV/7Xg4ahLwYGYZzuHGZKM5ZAQ=="
		}}
	}`)
}

func TestSignJSON(t *testing.T) {
	random := bytes.NewBuffer([]byte("Some 32 randomly generated bytes"))
	entityName := "example.com"
	keyID := "ed25519:my_key_id"
	input := []byte(`{"this":"is","my":"message"}`)

	publicKey, privateKey, err := ed25519.GenerateKey(random)
	if err != nil {
		t.Fatal(err)
	}

	signed, err := SignJSON(entityName, keyID, privateKey, input)
	if err != nil {
		t.Fatal(err)
	}

	err = VerifyJSON(entityName, keyID, publicKey, signed)
	if err != nil {
		t.Errorf("VerifyJSON(%q)", signed)
		t.Fatal(err)
	}
}

func IsJSONEqual(a, b []byte) bool {
	canonicalA, err := CanonicalJSON(a)
	if err != nil {
		panic(err)
	}
	canonicalB, err := CanonicalJSON(b)
	if err != nil {
		panic(err)
	}
	return string(canonicalA) == string(canonicalB)
}

func TestSignJSONTestVectors(t *testing.T) {
	// Check JSON signing using the test vectors from https://matrix.org/docs/spec/appendices.html
	seed, err := base64.RawStdEncoding.DecodeString("YJDBA9Xnr2sVqXD9Vj7XVUnmFZcZrlw8Md7kMW+3XA1")
	if err != nil {
		t.Fatal(err)
	}
	random := bytes.NewBuffer(seed)
	entityName := "domain"
	keyID := "ed25519:1"

	_, privateKey, err := ed25519.GenerateKey(random)
	if err != nil {
		t.Fatal(err)
	}

	testSign := func(input string, want string) {
		signed, err := SignJSON(entityName, keyID, privateKey, []byte(input))
		if err != nil {
			t.Fatal(err)
		}

		if !IsJSONEqual([]byte(want), signed) {
			t.Fatalf("VerifyJSON(%q): want %v got %v", input, want, string(signed))
		}
	}

	testSign(`{}`, `{
		"signatures":{
			"domain":{
				"ed25519:1":"K8280/U9SSy9IVtjBuVeLr+HpOB4BQFWbg+UZaADMtTdGYI7Geitb76LTrr5QV/7Xg4ahLwYGYZzuHGZKM5ZAQ"
			}
		}
	}`)

	testSign(`{"one":1,"two":"Two"}`, `{
		"one": 1,
		"signatures": {
			"domain": {
				"ed25519:1": "KqmLSbO39/Bzb0QIYE82zqLwsA+PDzYIpIRA2sRQ4sL53+sN6/fpNSoqE7BP7vBZhG6kYdD13EIMJpvhJI+6Bw"
			}
		},
		"two": "Two"
	}`)
}

type MyMessage struct {
	Unsigned   *json.RawMessage `json:"unsigned"`
	Content    *json.RawMessage `json:"content"`
	Signatures *json.RawMessage `json:"signature,omitempty"`
}

func TestSignJSONWithUnsigned(t *testing.T) {
	random := bytes.NewBuffer([]byte("Some 32 randomly generated bytes"))
	entityName := "example.com"
	keyID := "ed25519:my_key_id"
	content := json.RawMessage(`{"signed":"data"}`)
	unsigned := json.RawMessage(`{"unsigned":"data"}`)
	message := MyMessage{&unsigned, &content, nil}

	input, err := json.Marshal(&message)
	if err != nil {
		t.Fatal(err)
	}

	publicKey, privateKey, err := ed25519.GenerateKey(random)
	if err != nil {
		t.Fatal(err)
	}

	signed, err := SignJSON(entityName, keyID, privateKey, input)
	if err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal(signed, &message); err != nil {
		t.Fatal(err)
	}
	newUnsigned := json.RawMessage(`{"different":"data"}`)
	message.Unsigned = &newUnsigned
	input, err = json.Marshal(&message)
	if err != nil {
		t.Fatal(err)
	}

	err = VerifyJSON(entityName, keyID, publicKey, signed)
	if err != nil {
		t.Errorf("VerifyJSON(%q)", signed)
		t.Fatal(err)
	}
}
