package orm

import "testing"

func TestORMConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  ORMConfig
		wantErr bool
	}{
		{
			name: "PostgreSQL",
			config: ORMConfig{
				Driver: DriverPostgres,
				DSN:    "host=localhost user=app dbname=app port=5432 sslmode=disable",
			},
		},
		{
			name: "MySQL",
			config: ORMConfig{
				Driver: DriverMySQL,
				DSN:    "app:secret@tcp(localhost:3306)/app?multiStatements=true",
			},
		},
		{
			name: "SQLite",
			config: ORMConfig{
				Driver: DriverSQLite,
				DSN:    "database.sqlite",
			},
		},
		{
			name: "Missing driver",
			config: ORMConfig{
				DSN: "database.sqlite",
			},
			wantErr: true,
		},
		{
			name: "Unsupported driver",
			config: ORMConfig{
				Driver: "sqlserver",
			},
			wantErr: true,
		},
		{
			name: "Missing DSN",
			config: ORMConfig{
				Driver: DriverMySQL,
			},
			wantErr: true,
		},
		{
			name: "MySQL DSN without migration support",
			config: ORMConfig{
				Driver: DriverMySQL,
				DSN:    "app:secret@tcp(localhost:3306)/app",
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.config.Validate()
			if (err != nil) != test.wantErr {
				t.Fatalf("Validate() error = %v, want error %t", err, test.wantErr)
			}
		})
	}
}
