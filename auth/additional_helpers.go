package auth

import (
	"crypto/sha512"
	"encoding/base64"
)

func (h *UsersHelper) AddUser(username, password string, isSuper bool, groups []string, personID string) (*UserHelper, error) {
	user, err := h.New(username)
	if err != nil {
		return nil, err
	}

	if err = user.SetUsername(username); err != nil {
		return user, err
	}
	if err = user.SetIsSuper(isSuper); err != nil {
		return user, err
	}
	if err = user.SetPersonID(personID); err != nil {
		return user, err
	}
	for _, group := range groups {
		if err := user.Groups().Add(group); err != nil {
			return user, err
		}
	}

	return user, nil
}

func (h *UserHelper) SetPassword(password string) error {
	hash := sha512.New()
	hash.Write([]byte(password))
	passwordHash := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	if err := h.SetPasswordHash(passwordHash); err != nil {
		return err
	}
	if err := h.SetPasswordHashType("sha512"); err != nil {
		return err
	}

	return nil
}
