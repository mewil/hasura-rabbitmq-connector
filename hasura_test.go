package main

import (
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	listTablesTestResponse = []byte(`
{
    "result_type": "TuplesOk",
    "result": [
        [
            "schemaname",
            "tablename",
            "tableowner",
            "tablespace",
            "hasindexes",
            "hasrules",
            "hastriggers",
            "rowsecurity"
        ],
        [
            "hdb_catalog",
            "hdb_version",
            "doadmin",
            "NULL",
            "t",
            "f",
            "f",
            "f"
        ],
        [
            "mewil",
            "test",
            "user",
            "NULL",
            "t",
            "f",
            "t",
            "f"
        ]
    ]
}`)
)

func TestDefaultHasuraClient_ListTables(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		require.Nil(t, err)
		require.Equal(t, body, []byte(listTablesRequest))
		_, err = w.Write(listTablesTestResponse)
		require.Nil(t, err)
	}))
	defer s.Close()
	c := NewDefaultHasuraClient(s.URL, "mewil")
	tables, err := c.ListTables()
	require.Nil(t, err)
	require.Equal(t, []string{"test"}, tables)
}
