CREATE TABLE marks (
	nickname TEXT NOT NULL,
	comment TEXT,
	latitude FLOAT8 NOT NULL,
	longitude FLOAT8 NOT NULL,
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
