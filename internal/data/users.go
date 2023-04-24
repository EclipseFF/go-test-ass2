package data

import (
	"context" // New import
	"crypto/sha256"
	"database/sql" // New import
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"greenlight.bcc/internal/validator"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

var AnonymousUser = &User{}

type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes long")

	ValidateEmail(v, user.Email)

	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}

	if user.Password.hash == nil {
		panic("missing password hash for user")
	}

}

type password struct {
	plaintext *string
	hash      []byte
}

func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}
	p.plaintext = &plaintextPassword
	p.hash = hash
	return nil
}

func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

type UserModel struct {
	DB *sql.DB
}

func (m UserModel) Insert(user *User) error {
	query := `
	INSERT INTO users (name, email, password_hash, activated)
	VALUES ($1, $2, $3, $4)
	RETURNING id, created_at, version`
	args := []any{user.Name, user.Email, user.Password.hash, user.Activated}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}

func (m UserModel) GetByEmail(email string) (*User, error) {
	query := `
	SELECT id, created_at, name, email, password_hash, activated, version
	FROM users
	WHERE email = $1`
	var user User
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}

func (m UserModel) Update(user *User) error {
	query := `
	UPDATE users
	SET name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1
	WHERE id = $5 AND version = $6
	RETURNING version`
	args := []any{
		user.Name,
		user.Email,
		user.Password.hash,
		user.Activated,
		user.ID,
		user.Version,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

func (m UserModel) GetForToken(tokenScope, tokenPlaintext string) (*User, error) {
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	query := `
	SELECT users.id, users.created_at, users.name, users.email, users.password_hash, users.activated, users.version
	FROM users
	INNER JOIN tokens
	ON users.id = tokens.user_id
	WHERE tokens.hash = $1
	AND tokens.scope = $2
	AND tokens.expiry > $3`

	args := []any{tokenHash[:], tokenScope, time.Now()}
	var user User
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

type MockUserModel struct {
	DB *sql.DB
}

func (m MockUserModel) Insert(user *User) error {
	if user.Name == "invalid" {
		return errors.New("invalid name")
	}
	if user.Name == "permissions fall" {
		user.ID = 1
	}
	if user.Name == "token fall" {
		user.ID = 2
	}
	if user.Email != "email@gmail.com" {
		return ErrDuplicateEmail
	}
	return nil
}

func (m MockUserModel) GetByEmail(email string) (*User, error) {
	passwd := "TestPassword"
	sha, _ := bcrypt.GenerateFromPassword([]byte(passwd), 10)

	switch email[4] {
	case '1':
		return nil, ErrRecordNotFound
	case '2':
		return nil, errors.New("database has fallen")
	case '3':
		return &User{
			ID:        1,
			CreatedAt: time.Now(),
			Name:      "Test",
			Email:     "test@test.com",
			Password:  password{plaintext: &passwd, hash: []byte{}},
			Activated: true,
			Version:   1,
		}, nil
	case '4':
		invalidPassword := "invalid_password"
		sha2, _ := bcrypt.GenerateFromPassword([]byte(passwd), 10)
		return &User{
			ID:        1,
			CreatedAt: time.Now(),
			Name:      "Test",
			Email:     "test@test.com",
			Password:  password{plaintext: &invalidPassword, hash: sha2},
			Activated: true,
			Version:   1,
		}, nil
	case '5':
		return &User{
			ID:        2,
			CreatedAt: time.Now(),
			Name:      "Test",
			Email:     "test@test.com",
			Password:  password{plaintext: &passwd, hash: sha},
			Activated: true,
			Version:   1,
		}, nil

	}

	return &User{
		ID:        1,
		CreatedAt: time.Now(),
		Name:      "Test",
		Email:     "test@test.com",
		Password:  password{plaintext: &passwd, hash: sha},
		Activated: true,
		Version:   1,
	}, nil
}

func (m MockUserModel) Update(user *User) error {
	if user.Email == "testConflict@test.com" {
		return ErrEditConflict
	}
	if user.Email == "testErr@test.com" {
		return errors.New("something went wrong")
	}

	return nil
}

func (m MockUserModel) GetForToken(tokenScope, tokenPlaintext string) (*User, error) {

	passwd := "testPassword"
	sha, _ := bcrypt.GenerateFromPassword([]byte(passwd), 10)

	switch tokenPlaintext[len(tokenPlaintext)-1] {
	case '1':
		return nil, ErrRecordNotFound
	case '2':
		return nil, errors.New("some err")
	case '3':
		return &User{
			ID:        1,
			CreatedAt: time.Now(),
			Name:      "Test",
			Email:     "testConflict@test.com",
			Password:  password{plaintext: &passwd, hash: sha},
			Activated: true,
			Version:   1,
		}, nil
	case '4':
		return &User{
			ID:        1,
			CreatedAt: time.Now(),
			Name:      "Test",
			Email:     "testErr@test.com",
			Password:  password{plaintext: &passwd, hash: sha},
			Activated: true,
			Version:   1,
		}, nil
	case '5':
		return &User{
			ID:        2,
			CreatedAt: time.Now(),
			Name:      "Test",
			Email:     "test@test.com",
			Password:  password{plaintext: &passwd, hash: sha},
			Activated: true,
			Version:   1,
		}, nil

	}

	return &User{
		ID:        1,
		CreatedAt: time.Now(),
		Name:      "Test",
		Email:     "test@test.com",
		Password:  password{plaintext: &passwd, hash: sha},
		Activated: true,
		Version:   1,
	}, nil
}