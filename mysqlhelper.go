package goweb

import (
	"database/sql"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"reflect"
)

var connectionPool chan *MysqlHelper = make(chan *MysqlHelper, 100)
var connectionLimit chan struct{}
var ConnectionLimitError = errors.New("Connection limit reached")
var dbconn string
var minpoolsize int
var maxpoolsize int
var isblockonlimit bool

type MysqlHelper struct {
	db *sql.DB
}

func NewMysqlHelper(cfg map[string]string) *MysqlHelper {
	mysqlhelper := &MysqlHelper{
		nil,
	}
	dbconn = cfg["dbconn"]
	minpoolsize = ToInt(cfg["minpoolsize"], 10)
	maxpoolsize = ToInt(cfg["maxpoolsize"], 0)
	isblockonlimit = ToBool(cfg["isblockonlimit"], false)
	if maxpoolsize > 0 {
		connectionLimit = make(chan struct{}, maxpoolsize)
	} else {
		connectionLimit = nil
	}
	return mysqlhelper
}

func (m *MysqlHelper) getConn() (*MysqlHelper, error) {
	if connectionLimit != nil {
		if isblockonlimit {
			connectionLimit <- struct{}{}
		} else {
			select {
			case connectionLimit <- struct{}{}:
			default:
				return nil, ConnectionLimitError
			}
		}
	}
	select {
	case m := <-connectionPool:
		return m, nil
	default:
	}
	db, err := sql.Open("mysql", dbconn)
	if err != nil {
		return nil, err
	}
	m = new(MysqlHelper)
	m.db = db
	return m, nil
}

func (m *MysqlHelper) close() error {
	if connectionLimit != nil {
		<-connectionLimit
	}
	select {
	case connectionPool <- m:
		return nil
	default:
	}
	return m.db.Close()
}

func (m *MysqlHelper) Insert(query string, args ...interface{}) (int64, error) {
	conn, err := m.getConn()
	if err != nil {
		return -1, err
	}
	defer m.close()

	stmt, err := conn.db.Prepare(query)
	if err != nil {
		return -1, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(args...)
	if err != nil {
		return -1, err
	}
	return res.LastInsertId()
}

func (m *MysqlHelper) Update(query string, args ...interface{}) (int64, error) {
	conn, err := m.getConn()
	if err != nil {
		return -1, err
	}
	defer m.close()

	stmt, err := conn.db.Prepare(query)
	if err != nil {
		return -1, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(args...)
	if err != nil {
		return -1, err
	}
	return res.RowsAffected()
}

func (m *MysqlHelper) QueryForMap(query string, args ...interface{}) (map[string]interface{}, error) {
	conn, err := m.getConn()
	if err != nil {
		return nil, err
	}
	defer m.close()

	stmt, err := conn.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	values := make([]interface{}, len(cols))

	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	if rows.Next() {
		if err := rows.Scan(scanArgs...); err != nil {
			return nil, err
		}
		result := make(map[string]interface{}, len(cols))
		for ii, key := range cols {
			if scanArgs[ii] == nil {
				continue
			}
			value := reflect.Indirect(reflect.ValueOf(scanArgs[ii]))
			if value.Elem().Kind() == reflect.Slice {
				result[key] = string(value.Interface().([]byte))
			} else {
				result[key] = value.Interface()
			}
		}
		return result, nil
	}
	return nil, nil
}

func (m *MysqlHelper) QueryForMapSlice(query string, args ...interface{}) ([]map[string]interface{}, error) {
	conn, err := m.getConn()
	if err != nil {
		return nil, err
	}
	defer m.close()

	stmt, err := conn.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	values := make([]interface{}, len(cols))

	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	var results []map[string]interface{}
	for rows.Next() {
		if err := rows.Scan(scanArgs...); err != nil {
			return nil, err
		}
		result := make(map[string]interface{}, len(cols))
		for ii, key := range cols {
			if scanArgs[ii] == nil {
				continue
			}
			value := reflect.Indirect(reflect.ValueOf(scanArgs[ii]))
			if value.Elem().Kind() == reflect.Slice {
				result[key] = string(value.Interface().([]byte))
			} else {
				result[key] = value.Interface()
			}
		}
		results = append(results, result)
	}
	return results, nil
}
