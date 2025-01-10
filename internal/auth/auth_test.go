package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T) {
	token, err := MakeJWT(uuid.New(), "Secret", time.Hour)
	if err != nil {
		t.Log(err)
		return
	}
	t.Log("token: ", token)
}

func TestValidateJWTOK(t *testing.T) {
	newuid := uuid.New()
	t.Log("UID is: ", newuid)
	token, err := MakeJWT(newuid, "Secret", time.Hour)
	if err != nil {
		t.Log(err)
		return
	}

	uid, err := ValidateJWT(token, "Secret")
	if err != nil {
		t.Log(err)
		return
	}
	t.Log("Recovered UID: ", uid)

}

func TestValidateJWTNOK(t *testing.T) {
	newuid := uuid.New()
	t.Log("UID is: ", newuid)
	token, err := MakeJWT(newuid, "Secret", time.Hour)
	if err != nil {
		t.Log(err)
		return
	}

	uid, err := ValidateJWT(token, "Secret2")
	if err != nil {
		t.Log(err)
		return
	}
	t.Log("Recovered UID: ", uid)

}

func TestValidateJWTExpired(t *testing.T) {
	newuid := uuid.New()
	t.Log("UID is: ", newuid)
	token, err := MakeJWT(newuid, "Secret", -time.Hour)
	if err != nil {
		t.Log(err)
		return
	}

	uid, err := ValidateJWT(token, "Secret")
	if err != nil {
		t.Log(err)
		return
	}
	t.Log("Recovered UID: ", uid)

}
