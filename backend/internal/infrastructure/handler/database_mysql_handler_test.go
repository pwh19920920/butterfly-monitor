package handler

import "testing"

func TestValidateReadOnlySQL(t *testing.T) {
	cases := []struct {
		name    string
		sql     string
		wantErr bool
	}{
		{"normal select", "SELECT COUNT(*) FROM orders", false},
		{"select with where", "SELECT avg(latency) FROM metrics WHERE host='a'", false},
		{"leading paren select", "(SELECT max(v) FROM t)", false},
		{"lowercase select", "select 1", false},
		{"show", "SHOW TABLES", false},
		{"with cte", "WITH c AS (SELECT 1) SELECT * FROM c", false},
		{"desc", "DESC orders", false},
		{"explain", "EXPLAIN SELECT * FROM t", false},
		{"empty", "", true},
		{"multi-statement", "SELECT 1; DROP TABLE t", true},
		{"trailing semicolon", "SELECT 1;", true},
		{"insert", "INSERT INTO t VALUES(1)", true},
		{"update", "UPDATE t SET v=1", true},
		{"delete", "DELETE FROM t", true},
		{"drop", "DROP TABLE t", true},
		{"create", "CREATE TABLE t (v int)", true},
		{"alter", "ALTER TABLE t ADD c int", true},
		{"truncate", "TRUNCATE TABLE t", true},
		{"select into outfile", "SELECT * FROM t INTO OUTFILE '/x'", true},
		{"select for update", "SELECT * FROM t FOR UPDATE", true},
		{"grant", "GRANT ALL ON *.* TO 'x'", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := validateReadOnlySQL(c.sql)
			if c.wantErr && err == nil {
				t.Errorf("validateReadOnlySQL(%q) expected error, got nil", c.sql)
			}
			if !c.wantErr && err != nil {
				t.Errorf("validateReadOnlySQL(%q) expected nil, got %v", c.sql, err)
			}
		})
	}
}
