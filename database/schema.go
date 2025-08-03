package database

func GetTableQueries() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS users (
	user_id             SERIAL PRIMARY KEY,
	name VARCHAR(25)    NOT NULL,
	email VARCHAR(100)  UNIQUE NOT NULL,
	password            TEXT NOT NULL,
	role VARCHAR(20)    DEFAULT 'user', -- 'admin', 'client', 'hustler', or 'both'
	created             TEXT NOT NULL,
	updated             TEXT,
	resettoken          TEXT,
  resettokenexpiry    TIMESTAMP WITH TIME ZONE,
	isDeleted           BOOLEAN DEFAULT FALSE

  )`,

		`CREATE TABLE IF NOT EXISTS task (
  task_id 				SERIAL PRIMARY KEY,
  title			      TEXT NOT NULL,
  description		  TEXT,
  category		 	  VARCHAR(50),
  location			  VARCHAR(100),
  time_preference VARCHAR(50),
  price           INTEGER,
  status			    VARCHAR(20) DEFAULT 'open',
  posted_by 		  INTEGER REFERENCES users(user_id),
  assigned_to 		INTEGER REFERENCES users(user_id),
  expires_at 		  TIMESTAMP,
  created_at 		  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at 		  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  isDeleted           BOOLEAN DEFAULT FALSE
)`,

		`CREATE TABLE IF NOT EXISTS task_applications (
  application_id SERIAL PRIMARY KEY,
  task_id        INTEGER REFERENCES task(task_id),
  applicant_id   INTEGER REFERENCES users(user_id),
  message        TEXT,
  status			    VARCHAR(20) DEFAULT 'pending',
  created_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  isDeleted           BOOLEAN DEFAULT FALSE
);
`,
	}
}