CREATE TABLE
    monitoring_draft_laws.favorites (
        user_id INTEGER NOT NULL,
        project_id TEXT NOT NULL
    );

ALTER TABLE ONLY monitoring_draft_laws.favorites
ADD CONSTRAINT favorites_un UNIQUE (user_id, project_id);

ALTER TABLE ONLY monitoring_draft_laws.favorites
ADD CONSTRAINT favorites_fk FOREIGN KEY (project_id) REFERENCES monitoring_draft_laws.documents_data (projectid) ON UPDATE CASCADE ON DELETE CASCADE;

---- create above / drop below ----
DROP TABLE monitoring_draft_laws.favorites
