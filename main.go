package main

import (
	"database/sql"
	"flag"
	// we discussed how the following import entry only exectues the ``side-effects'' of the
	// package and doesn't import any symbols into our namespace.
	_ "github.com/lib/pq"
	"html/template"
	"io"
	// we only did this to demonstrate that package names can be changed.
	olog "log"
	"net/http"
	"os"
	"time"
)

type Statement struct {
	DB     *sql.DB
	ID     int
	Name   string
	Email  string
	Ledger []Transaction
	Namer  func() string
}

func (s *Statement) EmailName() string {
	return s.Email + " " + s.Name
}

type Transaction struct {
	ID          int
	Account     int
	Timestamp   string
	Description string
	Amount      string
	Fee         string
	Balance     string
}

func (s *Statement) Get() error {
	if err := s.GetAccount(); err != nil {
		return err
	}
	if err := s.GetLedger(); err != nil {
		return err
	}
	return nil
}

func (s *Statement) GetLedger() error {
	rows, err := s.DB.Query(`select id, account, "timestamp", description, amount, fee, balance from ledger where account=$1;`, s.ID)

	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		tx := Transaction{}
		err := rows.Scan(&tx.ID, &tx.Account, &tx.Timestamp, &tx.Description, &tx.Amount, &tx.Fee, &tx.Balance)
		if err != nil {
			return err
		}
		s.Ledger = append(s.Ledger, tx)
	}

	return nil
}

func (s *Statement) GetAccount() error {
	row := s.DB.QueryRow("select name, email from account where id=$1;", s.ID)

	err := row.Scan(&s.Name, &s.Email)

	if err != nil {
		return err
	}

	return nil
}

func dbConnect() (*sql.DB, error) {
	db, err := sql.Open("postgres", "user=ayan host=/var/run/postgresql dbname=bank")
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func currentdate() string {
	return time.Now().String()
}

type App struct {
	tmpl      *template.Template
	Statement Statement
	Title     string
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.tmpl.ExecuteTemplate(w, "index.html", a)
}

func main() {
	account := flag.Int("account", 1, "Account ID of statement")
	quiet := flag.Bool("quiet", false, "Disable log output")
	flag.Parse()

	logWriter := func() io.Writer {
		if *quiet == true {
			davefile, _ := os.OpenFile("dave.log", os.O_WRONLY|os.O_CREATE, 0644)
			return davefile
		}
		return os.Stderr
	}()

	log := olog.New(logWriter, "DEBUG ", olog.Lshortfile|olog.LstdFlags)

	db, err := dbConnect()

	if err != nil {
		log.Fatal(err)
	}

	stmt := Statement{
		ID: *account,
		DB: db,
	}

	if err := stmt.Get(); err != nil {
		log.Fatal(err)
	}

	tmpl, _ := template.New("dave").Parse(`HELLO DAVE`)
	tmpl = tmpl.Funcs(template.FuncMap{
		"currentdate": currentdate,
	})
	tmpl, _ = tmpl.ParseGlob("*.html")

	myapp := App{
		tmpl:      tmpl,
		Statement: stmt,
		Title:     "my first template page",
	}

	svr := http.Server{
		Addr:    ":8000",
		Handler: &myapp,
	}

	log.Fatal(svr.ListenAndServe())
}
