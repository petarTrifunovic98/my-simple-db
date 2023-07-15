package row

import "fmt"

const USERNAME_LEN uint32 = 32
const EMAIL_LEN uint32 = 256

type Row struct {
	Id       uint32
	Username [USERNAME_LEN]byte
	Email    [EMAIL_LEN]byte
}

func (r *Row) ToString() string {
	return fmt.Sprintf("ID: %d, Username: %s, Email: %s", r.Id, string(r.Username[:]), string(r.Email[:]))

}
