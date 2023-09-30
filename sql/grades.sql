CREATE TABLE grades (
    id SERIAL PRIMARY KEY,
    grader_name TEXT NOT NULL,
    proposal_id TEXT NOT NULL,
    grade INT NOT NULL
);

CREATE UNIQUE INDEX ON grades(grader_name, proposal_id);
