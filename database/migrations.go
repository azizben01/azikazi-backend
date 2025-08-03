package database

// Migration functions to add isDeleted columns to existing tables
func GetMigrationQueries() []string {
	return []string{
		// Add isDeleted column to users table if it doesn't exist
		`DO $$ 
		BEGIN 
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'isdeleted') THEN
				ALTER TABLE users ADD COLUMN isDeleted BOOLEAN DEFAULT FALSE;
			END IF;
		END $$;`,

		// Add isDeleted column to task table if it doesn't exist
		`DO $$ 
		BEGIN 
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'task' AND column_name = 'isdeleted') THEN
				ALTER TABLE task ADD COLUMN isDeleted BOOLEAN DEFAULT FALSE;
			END IF;
		END $$;`,

		// Add isDeleted column to task_applications table if it doesn't exist
		`DO $$ 
		BEGIN 
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'task_applications' AND column_name = 'isdeleted') THEN
				ALTER TABLE task_applications ADD COLUMN isDeleted BOOLEAN DEFAULT FALSE;
			END IF;
		END $$;`,

		// Remove CASCADE constraints from task_applications table
		`DO $$ 
		BEGIN 
			-- Drop existing foreign key constraints if they exist
			IF EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE constraint_name = 'task_applications_task_id_fkey') THEN
				ALTER TABLE task_applications DROP CONSTRAINT task_applications_task_id_fkey;
			END IF;
			IF EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE constraint_name = 'task_applications_applicant_id_fkey') THEN
				ALTER TABLE task_applications DROP CONSTRAINT task_applications_applicant_id_fkey;
			END IF;
			
			-- Add new foreign key constraints without CASCADE
			ALTER TABLE task_applications ADD CONSTRAINT task_applications_task_id_fkey 
				FOREIGN KEY (task_id) REFERENCES task(task_id);
			ALTER TABLE task_applications ADD CONSTRAINT task_applications_applicant_id_fkey 
				FOREIGN KEY (applicant_id) REFERENCES users(user_id);
		END $$;`,
	}
}; 