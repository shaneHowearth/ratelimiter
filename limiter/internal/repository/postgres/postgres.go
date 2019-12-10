package postgres

import (
	"database/sql"
	"fmt"
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
func (p *Datastore) CreateAndCheck(ip string, limit int, timestamp time.Time, timespan time.Duration) (bool, time.Duration, error) {
	// Check if the limit has been reached
	over, minWait, err := p.ReachedMax(ip, limit, timespan)
	if err != nil {
		return true, time.Hour, err
	}
	// Add this connection attempt to the DB
	if !over {
		err = p.Create(ip, timestamp)
	}
	return over, minWait, err
}

// ReachedMax -
func (p *Datastore) ReachedMax(ip string, limit int, timespan time.Duration) (bool, time.Duration, error) {
	if p.db == nil {
		if perr := p.Connect(); perr != nil {
			log.Panic("Unable to connect to database, dying")
		}
	}

	// Get a count of the number of connections stored in the DB for this ip, between now and now - timespan
	count := 0
	var accessTime time.Time
	err := p.db.QueryRow(`SELECT MAX(access.access_time), count(*) FROM access WHERE access.ip = $1 and access.access_time > (Now() - $2) LIMIT 1`, ip, timespan).Scan(&accessTime, &count)
	if err != nil {
		return true, time.Hour, err
	}

	// >= because this attempt would be over the limit
	if count >= limit {
		wait := time.Second*3600 - (time.Since(accessTime))
		return true, wait, nil
	}

	return false, time.Second, nil
}

// Create -
func (p *Datastore) Create(ip string, timestamp time.Time) error {
	if p.db == nil {
		if perr := p.Connect(); perr != nil {
			log.Panic("Unable to connect to database, dying")
		}
	}
	err := p.db.QueryRow(`INSERT INTO access(ip, access_time)
	VALUES($1, $2)`, ip, timestamp)
	if err != nil {
		log.Printf("Create error %#+v", err)
		return fmt.Errorf("datastore Create returned error: %v", err)
	}
	return nil
}
