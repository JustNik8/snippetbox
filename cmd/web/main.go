package main

import (
	"database/sql"
	"flag"
	"justnik.com/snippetbox/pkg/models/mysql"
	"log"
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
)

type application struct {
	errorLogger *log.Logger
	infoLogger  *log.Logger
	snippets    *mysql.SnippetModel
}

func main() {

	addr := flag.String("addr", ":4000", "network address HTTP")
	dsn := flag.String(
		"dsn",
		"web:pass@(127.0.0.1:3308)/snippetbox?parseTime=true",
		"Название MySQL источника данных",
	)
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO:\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR:\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := openDb(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()

	app := &application{
		errorLogger: errorLog,
		infoLogger:  infoLog,
		snippets:    &mysql.SnippetModel{DB: db},
	}

	infoLog.Printf("Запуск веб-сервера на %s\n", *addr)

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}
	// Для запуска нового веб сервера.
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

type neuteredFileSystem struct {
	fs http.FileSystem
}

func (nfs neuteredFileSystem) Open(path string) (http.File, error) {
	f, err := nfs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if s.IsDir() {
		index := filepath.Join(path, "index.html")
		if _, err := nfs.fs.Open(index); err != nil {
			closeErr := f.Close()
			if closeErr != nil {
				return nil, closeErr
			}

			return nil, err
		}
	}

	return f, nil
}

func openDb(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
