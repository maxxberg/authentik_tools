package membership

import "time"

type User interface {
	GetPk() int32
	GetPkOk() (*int32, bool)
	GetUsername() string
	GetUsernameOk() (*string, bool)
	GetName() string
	GetNameOk() (*string, bool)
	GetIsActive() bool
	GetIsActiveOk() (*bool, bool)
	HasIsActive() bool
	GetLastLogin() time.Time
	GetLastLoginOk() (*time.Time, bool)
	HasLastLogin() bool
	GetEmail() string
	GetEmailOk() (*string, bool)
	HasEmail() bool
	GetAttributes() map[string]interface{}
	GetAttributesOk() (map[string]interface{}, bool)
	HasAttributes() bool
	GetUid() string
	GetUidOk() (*string, bool)
}
