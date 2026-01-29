CREATE TABLE
    monitoring_draft_laws.logs (
        created TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
        doc_id TEXT NOT NULL,
        user_id INTEGER NOT NULL,
        "values" jsonb DEFAULT '{}'::jsonb NOT NULL
    );

ALTER TABLE ONLY monitoring_draft_laws.logs
ADD CONSTRAINT logs_fk FOREIGN KEY (doc_id) REFERENCES monitoring_draft_laws.documents_data (projectid) ON UPDATE CASCADE ON DELETE CASCADE;

CREATE INDEX logs_created_idx ON monitoring_draft_laws.logs USING btree (created);

CREATE INDEX logs_doc_id_idx ON monitoring_draft_laws.logs USING btree (doc_id);

CREATE INDEX logs_user_id_idx ON monitoring_draft_laws.logs USING btree (user_id);

---- create above / drop below ----
DROP TABLE monitoring_draft_laws.logs
