package sql

type User struct {
	ID        int
	Email     string
	Password  string `db:"pass"`
	Signature string
}

func NewUser(email string, password string) User {
	return User{
		Email:     email,
		Password:  password,
		Signature: "",
	}
}

func (db *sqlImpl) NewUser(email string, password string) error {
	user := NewUser(email, password)
	_, err := db.tx.NamedExec("INSERT INTO users (email, pass, signature) VALUES (:email, :pass, :signature)", user)
	if err != nil {
		return err
	}
	err = db.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (db *sqlImpl) UpdateUserData(user User) error {
	sql := `
	UPDATE users SET
		email=:email,
		signature=:signature WHERE id=:id
	`
	_, err := db.db.NamedExec(
		sql,
		user,
	)
	return err
}

func (db *sqlImpl) GetUserByEmail(email string) (User, error) {
	var user User
	err := db.db.Get(&user, "SELECT * FROM users WHERE email=$1", email)
	return user, err
}
