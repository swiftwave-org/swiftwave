package gitmanager

type GitUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Manager struct {
	GitUser GitUser
}

type Repository struct {
	Name      string `json:"name"`
	Username  string `json:"username"`
	Branch    string `json:"branch"`
	IsPrivate bool   `json:"is_private"`
}
