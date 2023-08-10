package row

import (
	"bytes"
	"fmt"
)

const USERNAME_LEN uint32 = 32
const EMAIL_LEN uint32 = 256

type Row struct {
	Id       uint32             `json:"id"`
	Username [USERNAME_LEN]byte `json:"username"`
	Email    [EMAIL_LEN]byte    `json:"email"`
}

type RowDTO struct {
	Id       uint32 `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func (r *Row) ToString() string {
	return fmt.Sprintf("ID: %d, Username: %s, Email: %s", r.Id, string(r.Username[:]), string(r.Email[:]))
}

func (r *Row) ToRowDTO() *RowDTO {
	usernameBefore, _, _ := bytes.Cut(r.Username[:], []byte{0})
	emailBefore, _, _ := bytes.Cut(r.Email[:], []byte{0})

	ret := &RowDTO{
		Id:       r.Id,
		Username: string(usernameBefore),
		Email:    string(emailBefore),
	}

	return ret
}
