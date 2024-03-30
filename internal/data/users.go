package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"

	"github.com/m0hh/smart-logitics/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
	TraineeRole       = "trainee"
	CoachRole         = "coach"
	AdminRole         = "admin"
)

var AnonymousUser = &User{}

type User struct {
	Id        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
	Role      string    `json:"role"`
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
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
	v.Check(validator.In(user.Role, "admin", "coach", "trainee", "gym"), "role", "role must be one of these 'admin','coach','trainee','gym' ")

	ValidateEmail(v, user.Email)

	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}

	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}

type UserModel struct {
	DB *sql.DB
}

func (m UserModel) Insert(user *User) error {
	stmt := `INSERT INTO users (name, email,password_hash,role, activated)
	VALUES ($1,$2,$3,$4,$5)
	RETURNING id,created_at,version
	`
	args := []interface{}{user.Name, user.Email, user.Password.hash, user.Role, user.Activated}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, stmt, args...).Scan(&user.Id, &user.CreatedAt, &user.Version)
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

func (m UserModel) RetrieveByEmail(email string) (*User, error) {
	stmt := `SELECT id, name, email,version,role, activated, created_at, password_hash
	FROM users WHERE email = $1`

	var user User
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, stmt, email).Scan(
		&user.Id,
		&user.Name,
		&user.Email,
		&user.Version,
		&user.Role,
		&user.Activated,
		&user.CreatedAt,
		&user.Password.hash,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRcordNotFound
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

	args := []interface{}{
		user.Name,
		user.Email,
		user.Password.hash,
		user.Activated,
		user.Id,
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
        SELECT users.id, users.created_at, users.name, users.email, users.password_hash, users.activated, users.version, users.role
        FROM users
        INNER JOIN tokens
        ON users.id = tokens.user_id
        WHERE tokens.hash = $1
        AND tokens.scope = $2 
        AND tokens.expiry > $3`

	args := []interface{}{tokenHash[:], tokenScope, time.Now()}

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.Id,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
		&user.Role,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRcordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

type UserCard struct {
	Id            int64 `json:"id"`
	Owner         int64 `json:"owner"`
	Coach         int64 `json:"coach"`
	CurrentPlan   int64 `json:"current_plan"`
	Weight        int   `json:"weight"`
	CurrentExPlan int64 `json:"current_exercise_plan"`
}

func (m UserModel) CreateUserCardRegistration(usercard UserCard) error {
	stmt := ` INSERT INTO user_card (owner, coach) VALUES ($1,$2) RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, stmt, usercard.Owner, usercard.Coach).Scan(&usercard.Id)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "user_card_owner_key"`:
			return ErrWrongForeignKey
		default:
			return err
		}
	}

	return nil

}

func (m UserModel) ListCoachUsers(user User, filter Filters) ([]*User, Metadata, error) {
	stmt := `
	SELECT 
	count(*) OVER(), users.id, users.name, users.email from user_card INNER JOIN users ON 
	user_card.owner = users.id
	WHERE user_card.coach = $1
	LIMIT $2 OFFSET $3
	`
	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(context, stmt, user.Id, filter.limit(), filter.offset())

	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	var users []*User

	totalRecords := 0
	for rows.Next() {
		var user User
		err = rows.Scan(&totalRecords, &user.Id, &user.Name, &user.Email)
		if err != nil {
			return nil, Metadata{}, err
		}
		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filter.Page, filter.PageSize)

	return users, metadata, nil
}

func (m UserModel) RetrieveUserCard(user_card *UserCard) error {
	stmt := `SELECT  id, coach, current_plan, current_exercise_plan, current_weight  FROM user_card WHERE owner = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var u_card_coach interface{}
	var u_card_plan interface{}
	var u_card_ex_plan interface{}

	err := m.DB.QueryRowContext(ctx, stmt, user_card.Owner).Scan(&user_card.Id, &u_card_coach, &u_card_plan, &u_card_ex_plan, &user_card.Weight)

	var Id int64
	Id, ok := u_card_coach.(int64)
	if ok {
		user_card.Coach = Id

	}

	Id, ok = u_card_plan.(int64)
	if ok {
		user_card.CurrentPlan = Id
	}

	Id, ok = u_card_ex_plan.(int64)
	if ok {
		user_card.CurrentExPlan = Id
	}

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrRcordNotFound
		default:
			return err
		}
	}

	return nil
}

func (m UserModel) UpdateWeightUserCard(weight int, owner int64) error {
	stmt := `UPDATE user_card SET current_weight = $1 WHERE owner = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, stmt, weight, owner)

	if err != nil {

		return err
	}

	return nil
}

type UserHistory struct {
	Id              int64             `json:"id"`
	Plan            PlanMealCoachList `json:"plan"`
	ExercisePlan    string            `json:"ex_plan"`
	From            time.Time         `json:"form"`
	To              time.Time         `json:"to"`
	StartingWeight  int               `json:"starting_weight"`
	FinishingWeight int               `json:"finishing_weight"`
	Current         bool              `json:"current"`
}

func (m UserModel) ListHistory(user_id int64, filter Filters) ([]*UserHistory, Metadata, error) {
	stmt := `SELECT 
    count(*) OVER(),
    hist.id,
    hist.from_date,
    hist.to_date,
    hist.weight_start,
    hist.weight_finish,
    hist.is_now,
    d1.name ,
    d2.name ,
    d3.name ,
    d4.name ,
    d5.name ,
    d6.name ,
    d7.name ,
	e.name
	FROM 
		user_history AS hist
	JOIN 
		exercise_plan AS e ON hist.exercise_plan_done = e.id
	JOIN
	
		plan_meal AS pm ON hist.plan_done = pm.id
	JOIN 
		day AS d1 ON pm.first_day = d1.id
	JOIN 
		day AS d2 ON pm.second_day = d2.id
	JOIN 
		day AS d3 ON pm.third_day = d3.id
	JOIN 
		day AS d4 ON pm.fourth_day = d4.id
	JOIN 
		day AS d5 ON pm.fifth_day = d5.id
	JOIN 
		day AS d6 ON pm.sixth_day = d6.id
	JOIN 
		day AS d7 ON pm.seventh_day = d7.id
	WHERE hist.owner = $1
	LIMIT $2 OFFSET $3 `

	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(context, stmt, user_id, filter.limit(), filter.offset())

	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	var histories []*UserHistory

	totalRecords := 0
	for rows.Next() {
		var history UserHistory
		var to_time interface{}

		err = rows.Scan(
			&totalRecords,
			&history.Id,
			&history.From,
			&to_time,
			&history.StartingWeight,
			&history.FinishingWeight,
			&history.Current,
			&history.Plan.FirstDay,
			&history.Plan.SecondDay,
			&history.Plan.ThirdDay,
			&history.Plan.FourthDay,
			&history.Plan.FifthDay,
			&history.Plan.SixthDay,
			&history.Plan.SeventhDay,
			&history.ExercisePlan,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		var Time time.Time
		Time, ok := to_time.(time.Time)
		if ok {
			history.To = Time

		}
		histories = append(histories, &history)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filter.Page, filter.PageSize)

	return histories, metadata, nil
}

func (m UserModel) CoachPermitted(user_id, coach_id int64) (bool, error) {
	stmt := `SELECT users.id FROM user_card JOIN users ON user_card.owner = users.id WHERE user_card.coach = $1`

	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(context, stmt, coach_id)

	if err != nil {
		return false, err
	}
	defer rows.Close()

	var ids []int64

	for rows.Next() {
		var id int64
		err = rows.Scan(&id)
		if err != nil {
			return false, err
		}
		ids = append(ids, id)

	}

	for idx := range ids {
		if ids[idx] == user_id {
			return true, nil
		}
	}

	return false, nil

}
