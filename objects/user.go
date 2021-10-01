package objects

type User struct {
	ID       int
	Email    string
	Password string `db:"pass"`
}

func NewUser(email string, password string) User {
	return User{
		Email:    email,
		Password: password,
	}
}
