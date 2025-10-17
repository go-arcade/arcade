package id

import "github.com/rs/xid"

func GetXid() string {
	id := xid.New()
	return id.String()
}
