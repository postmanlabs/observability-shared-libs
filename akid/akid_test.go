package akid

import (
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestBaseAkitaIDReflexiveParse(t *testing.T) {
	akitaID := GenerateUserID()
	strAkitaID := String(akitaID)

	reAkitaID, err := ParseID(strAkitaID)
	if err != nil {
		t.Fatal("could not parse Akita ID", err)
	}

	if reAkitaID.GetType() != "usr" {
		t.Fatal("did not get expcted type 'usr' for parsed Akita ID, instead got", reAkitaID.GetType())
	}

	if reAkitaID.GetUUID() != akitaID.GetUUID() {
		t.Fatal("id of parsed Akita ID does not match original Akita ID.")
	}

	_, ok := reAkitaID.(UserID)
	if !ok {
		t.Fatal("parsed ID did not have correct underlying type")
	}
}

type akitaIDParseTest struct {
	idStr   string
	tag     string
	success bool
}

func TestBaseAkitaIDParse(t *testing.T) {
	tests := []akitaIDParseTest{
		// Only some of these tests use valid UUIDs. GetUUID is tested in
		// TestBaseAkitaIDReflexiveParse
		akitaIDParseTest{"usr_21raAVTqUKOHvmxgK0ySCZ", "usr", true},
		akitaIDParseTest{"lrn_7IObUDuFf0fddZ4Ix1DPAC", "lrn", true},
		akitaIDParseTest{"svc_6NiejyYEVpWfziUXJgovV6", "svc", true},
		akitaIDParseTest{"svc_s8VseBBxopFn5eDnVEoRa", "svc", true},
		akitaIDParseTest{"svc_0s8VseBBxopFn5eDnVEoRa", "svc", true},
		akitaIDParseTest{"svc_0", "svc", true},
		akitaIDParseTest{"svc_7n42DGM5Tflk9n8mt7Fhc7", "svc", true},
		// failure case because xxx is not in the set of valid ID prefixes
		akitaIDParseTest{"svc_7n42DGM5Tflk9n8mt7Fhc8", "svc", false},  // overflows
		akitaIDParseTest{"xxx_21raAVTqUKOHvmxgK0ySCZ", "xxx", false},
		akitaIDParseTest{"f!o_21raAVTqUKOHvmxgK0ySCZ", "f!o", false},
		akitaIDParseTest{"Foo_21raAVTqUKOHvmxgK0ySCZ", "Foo", false},
		akitaIDParseTest{"lar_aaaaaaaaaaaaaaaaaaaaa!", "lar", false},
		akitaIDParseTest{"derp_aaaaaaaaaaaaaaaaaaaaaa", "derp", false},
		akitaIDParseTest{"qux_asdf", "qux", false},
		akitaIDParseTest{"zed__aaaaaaaaaaaaaaaaaaaa", "zed", false},
	}

	for _, tst := range tests {
		akid, err := ParseID(tst.idStr)
		if err != nil && tst.success {
			t.Fatal("failed to parse valid Akita ID", err, tst)
		} else if err == nil && !tst.success {
			t.Fatal("success parsing invalid Akita ID", tst)
		} else if akid == nil && tst.success {
			t.Fatal("parse returned nil Akita ID", tst)
		} else if err == nil && akid.GetType() != tst.tag {
			t.Fatal("success parsing Akita ID but mismatched tags", tst.tag, akid.GetType())
		}
	}
}

func TestParsedIDHasCorrectUnderlyingType(t *testing.T) {
	idStr := "usr_21raAVTqUKOHvmxgK0ySCZ"

	akid, _ := ParseID(idStr)
	if _, ok := akid.(UserID); !ok {
		t.Fatal("parsed ID did not have correct underlying type")
	}
}

func TestParseIDAs(t *testing.T) {
	idStr := "usr_21raAVTqUKOHvmxgK0ySCZ"
	var userID UserID
	err := ParseIDAs(idStr, &userID)

	if err != nil {
		t.Fatal("could not parse ID", err)
	}

	if String(userID) != idStr {
		t.Fatal("incorrectedly parsed ID", idStr, "!=", String(userID))
	}

	var projectID ProjectID

	err = ParseIDAs(idStr, &projectID)
	if err == nil {
		t.Fatal("parsed akid string to wrong type incorrectly")
	}
}

func TestNilID(t *testing.T) {
	var defaultValueUserID UserID
	nilUserID := NewUserID(uuid.Nil)

	if defaultValueUserID != nilUserID {
		t.Fatal("default value of UserID should be equal to NewUserID(uuid.Nil)")
	}
}

func TestScanner(t *testing.T) {
	var userID UserID
	someUUID := uuid.New()
	bs := [16]byte(someUUID)

	var scanner sql.Scanner
	scanner = &userID

	err := scanner.Scan(bs[:])
	if err != nil {
		t.Fatal("error while scanning", err)
	}

	if userID.GetUUID() != someUUID {
		t.Fatalf("expected matching UUID value, have: %v, want: %v", userID.GetUUID(), someUUID)
	}
}

func TestValuer(t *testing.T) {
	userID := GenerateUserID()
	userIDVal, _ := userID.Value()
	userIDValStr := userIDVal.(string)
	uuidStr := userID.GetUUID().String()

	if uuidStr != userIDValStr {
		t.Fatalf("expected matching userID and uuidStr, got: %s want: %s", userIDValStr, uuidStr)
	}
}

func TestJSON(t *testing.T) {
	specID := GenerateAPISpecID()

	b, err := json.Marshal(specID)
	if err != nil {
		t.Errorf("failed to marshal as JSON: %v", err)
		return
	} else if string(b) != `"`+String(specID)+`"` {
		t.Errorf("unexpected marshaled form: %s vs %s", `"`+String(specID)+`"`, string(b))
		return
	}

	var unmarshaled APISpecID
	if err := json.Unmarshal(b, &unmarshaled); err != nil {
		t.Errorf("failed to unmarshal as JSON: %v", err)
		return
	} else if specID != unmarshaled {
		t.Errorf("mismatch after unmarshal: %s vs %s", String(specID), String(unmarshaled))
	}
}

func TestDecode15ByteUUID(t *testing.T) {
	var id UserID

	// Intentionally chosen such that the UUID part only takes up 15 bytes.
	if err := ParseIDAs("usr_111k0VhXEsqpG1J0DzuWU", &id); err != nil {
		t.Errorf("failed to parse from string: %v", err)
		return
	}

	if id.GetUUID().String() != "0089eaab-6427-cdc6-bb57-5b24649face6" {
		t.Errorf("unique part mismatch, got %s", id.GetUUID())
		return
	}
}

// Regression test for zero padding endianess bug.
func TestCollision(t *testing.T) {
	var id1, id2 LearnSessionID

	if err := ParseIDAs("lrn_1KEVR4fLE7asNBSBTJBZp", &id1); err != nil {
		t.Errorf("failed to parse from string: %v", err)
		return
	}
	if err := ParseIDAs("lrn_5TXtnnGheJMGVjINN1Dnua", &id2); err != nil {
		t.Errorf("failed to parse from string: %v", err)
		return
	}

	if id1 == id2 {
		t.Errorf("they should not collide")
	}
}
