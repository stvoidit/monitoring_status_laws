CREATE TABLE
    monitoring_draft_laws.additional_files (
        id UUID DEFAULT gen_random_uuid () NOT NULL,
        document_id TEXT NOT NULL,
        meta_info jsonb DEFAULT '{}'::jsonb NOT NULL,
        object_id oid,
        created TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
    );

ALTER TABLE ONLY monitoring_draft_laws.additional_files
ADD CONSTRAINT additional_files_pk PRIMARY KEY (id);

ALTER TABLE ONLY monitoring_draft_laws.additional_files
ADD CONSTRAINT additional_files_unique UNIQUE (object_id);

ALTER TABLE ONLY monitoring_draft_laws.additional_files
ADD CONSTRAINT additional_files_documents_data_fk FOREIGN KEY (document_id) REFERENCES monitoring_draft_laws.documents_data (projectid) ON UPDATE CASCADE ON DELETE CASCADE;

CREATE UNIQUE INDEX additional_files_id_idx ON monitoring_draft_laws.additional_files USING btree (id);

CREATE INDEX cancellation_phrases_idx ON monitoring_draft_laws.cancellation_policy USING gin (cancellation_phrases);

---- create above / drop below ----
DROP TABLE monitoring_draft_laws.additional_files
