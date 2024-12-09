package models

type ModelFiles struct {
	fs FileStorager
	us UsersStorager
	as AuthStorager
}

type ModelUsers struct {
	us UsersStorager
}

type ModelAuth struct {
	as AuthStorager
	us UsersStorager
}

func NewModelFiles(fs FileStorager, us UsersStorager, as AuthStorager) ModelFiles {
	return ModelFiles{fs, us, as}
}
func NewModelUsers(us UsersStorager) ModelUsers {
	return ModelUsers{us}
}
func NewModelAuth(as AuthStorager, us UsersStorager) ModelAuth {
	return ModelAuth{as, us}
}
