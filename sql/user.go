package sql

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

func (db *sqlImpl) NewUser(email string, password string) error {
	user := NewUser(email, password)
	_, err := db.tx.NamedExec("INSERT INTO users (email, pass) VALUES (:email, :pass)", user)
	if err != nil {
		return err
	}
	err = db.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (db *sqlImpl) GetUserByEmail(email string) (User, error) {
	var user User
	err := db.db.Get(&user, "SELECT * FROM users WHERE email=$1", email)
	return user, err
}
