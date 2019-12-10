package postgres

import (
	"database/sql"
	"log"
	"math/rand"
	"time"

	_ "github.com/lib/pq" // Postgres
)

// Datastore -
type Datastore struct {
	db    *sql.DB
	Retry int
	URI   string
}

var sqlOpen = sql.Open

// Connect - Create the connection to the database
func (p *Datastore) Connect() (err error) {
	// Retry MUST be >= 1
	if p.Retry == 0 {
		log.Print("Cannot use a Retry of zero, this process will to default retry to 5")
		p.Retry = 5
	}
	if p.URI == "" {
		log.Panicf("no Datastore URI configured")
	}

	log.Printf("Using DB URI: %s", p.URI)
	// Infinite loop
	// Keep trying forever
	for {
		for i := 0; i < p.Retry; i++ {
			p.db, err = sqlOpen("postgres", p.URI)
			if err == nil {
				if pingerr := p.db.Ping(); pingerr != nil {
					log.Printf("Unable to ping database with error %v", pingerr)
				} else {
					// Successful connection
					log.Print("Successfully connected to datastore DB")
					return nil
				}
			}
			time.Sleep(1 * time.Second)
		}

		backoff := time.Duration(p.Retry*rand.Intn(10)) * time.Second
		log.Printf("ALERT: Trouble connecting to Datastore, error: %v, going to re-enter retry loop in %s seconds", err, backoff.String())
		time.Sleep(backoff)
	}
}

// CreateAndCheck -
func (p *Datastore) CreateAndCheck(ip string, limit int, timestamp time.Time, timespan time.Duration) (bool, float64, error) {
	// Check if the limit has been reached
	over, minWait, err := p.ReachedMax(ip, limit, timespan)
	if err != nil {
		return true, float64(time.Hour.Seconds()), err
	}
	// Add this connection attempt to the DB
	if !over {
		err = p.Create(ip, timestamp)
	}
	return over, minWait, err
}

// ReachedMax -
func (p *Datastore) ReachedMax(ip string, limit int, timespan time.Duration) (bool, float64, error) {
	if p.db == nil {
		if perr := p.Connect(); perr != nil {
			log.Panic("Unable to connect to database, dying")
		}
	}

	log.Printf("Looking for %v", timespan)
	// Get a count of the number of connections stored in the DB for this ip, between now and now - timespan
	count := 0
	// var accessTime time.Time
	var accessTime sql.NullInt64
	// current_timestamp-interval '1 hour'
	err := p.db.QueryRow(`SELECT (CURRENT_TIMESTAMP - MIN(access.access_time)), count(*) FROM access WHERE access.ip = $1 AND access.access_time > CURRENT_TIMESTAMP- $2 * INTERVAL '1 SECOND' LIMIT 1`, ip, timespan.Seconds()).Scan(&accessTime, &count)
	if err != nil {
		log.Printf("Query generated error %v", err)
		return true, float64(time.Hour.Seconds()), err
	}

	log.Printf("Returned count %#v, wait %#v, err %v", count, accessTime, err)
	// >= because this attempt would be over the limit
	if count >= limit {
		var wait float64
		if accessTime.Valid {
			wait = timespan.Seconds() - float64(accessTime.Int64)
		}
		return true, wait, nil
	}

	log.Printf("No problem")
	return false, float64(0), nil
}

// Create -
func (p *Datastore) Create(ip string, timestamp time.Time) error {
	if p.db == nil {
		if perr := p.Connect(); perr != nil {
			log.Panic("Unable to connect to database, dying")
		}
	}
	log.Printf("Trying to insert %s with %#v", ip, timestamp)
	rows := p.db.QueryRow(`INSERT INTO access(ip, access_time)
	VALUES($1, $2)`, ip, timestamp)
	log.Printf("Rows inserted %#+v", rows)
	return nil
}
