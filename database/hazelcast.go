package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/hazelcast/hazelcast-go-client"
	_ "github.com/hazelcast/hazelcast-go-client/sql/driver"
)

type Hazelcast struct {
	ConnString string
	Database   *sql.DB
}

func (db *Hazelcast) Update(q *Update) {
	protoQuery, columnOrder := db.GenerateQuery(q)
	values := make([]interface{}, len(columnOrder))
	updateValues := q.GetValues()
	for i, v := range columnOrder {
		var u interface{}
		if i == 0 {
			u = q.Update
		} else {
			u = updateValues[v]
		}

		if u == nil {
			u = "NULL"
		}

		values[i] = u
	}
	tx, err := db.GetDatabaseReference().Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare(protoQuery)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	stmt.Exec(values...)
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
}

func (db *Hazelcast) GetFileName() string {
	return db.ConnString
}

func (db *Hazelcast) GetDatabaseReference() *sql.DB {
	return db.Database
}

func (db *Hazelcast) CloseDatabaseReference() {
	db.GetDatabaseReference().Close()
	db.Database = nil
}

func (db *Hazelcast) SetDatabaseReference(connString string) {
	database := GetDatabaseForFile(connString)
	db.ConnString = connString
	db.Database = database
}

func (db Hazelcast) GetPlaceholderForDatabaseType() string {
	return "?"
}

func (db Hazelcast) GetTableNames() ([]string, error) {
	client, err := clientFromConnectionString(db.ConnString)
	if err != nil {
		return nil, err
	}
	ts, err := client.GetDistributedObjectsInfo(context.Background())
	if err != nil {
		return nil, err
	}
	var names []string
	for _, t := range ts {
		if t.ServiceName == hazelcast.ServiceNameMap {
			names = append(names, t.Name)
		}
	}
	return names, nil
}

func (db *Hazelcast) GenerateQuery(u *Update) (string, []string) {
	var (
		query         string
		querySkeleton string
		valueOrder    []string
	)

	placeholder := db.GetPlaceholderForDatabaseType()

	querySkeleton = fmt.Sprintf("UPDATE %s"+
		" SET %s=%s ", u.TableName, u.Column, placeholder)
	valueOrder = append(valueOrder, u.Column)

	whereBuilder := strings.Builder{}
	whereBuilder.WriteString(" WHERE ")
	uLen := len(u.GetValues())
	i := 0
	for k := range u.GetValues() { // keep track of order since maps aren't deterministic
		assertion := fmt.Sprintf("%s=%s ", k, placeholder)
		valueOrder = append(valueOrder, k)
		whereBuilder.WriteString(assertion)
		if uLen > 1 && i < uLen-1 {
			whereBuilder.WriteString("AND ")
		}
		i++
	}
	query = querySkeleton + strings.TrimSpace(whereBuilder.String()) + ";"
	return query, valueOrder
}

func clientFromConnectionString(connStr string) (*hazelcast.Client, error) {
	/*
		config := hazelcast.Config{}
		addrsRest := strings.SplitN(connStr, ";", 2)
		if addrsRest[0] != "" {
			addrs := strings.Split(addrsRest[0], ",")
			config.Cluster.Network.SetAddresses(addrs...)
		}
	*/
	return hazelcast.StartNewClient(context.Background())
}
