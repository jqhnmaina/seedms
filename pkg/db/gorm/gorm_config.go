package gorm

// Option allows extra configuration for instantiating Roach. Use the With...
// functions to set options e.g.
//     nameOpt := WithDBName("my_app_db")
type Option func(*Gorm)

// WithDSN sets the DSN to be used by Roach.
func WithDSN(dsn string) Option {
	return func(r *Gorm) {
		r.dsn = dsn
	}
}

// WithDBName sets the name of the cockroach database to be used by Roach.
func WithDBName(db string) Option {
	return func(r *Gorm) {
		r.dbName = db
	}
}
