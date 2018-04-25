package client_announcer

import (
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
)

func TestGetChannelUsers(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO ws_channel_members").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
}
