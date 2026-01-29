CREATE TABLE
    monitoring_draft_laws.documents_data (
        projectid TEXT NOT NULL,
        document_values jsonb NOT NULL,
        created TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
        HOST TEXT NOT NULL,
        updated TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
        journal jsonb DEFAULT '[]'::jsonb NOT NULL,
        archive BOOLEAN DEFAULT FALSE NOT NULL
    );

ALTER TABLE ONLY monitoring_draft_laws.documents_data
ADD CONSTRAINT documents_data_pk PRIMARY KEY (projectid);

CREATE INDEX document_valuess_idx ON monitoring_draft_laws.documents_data USING gin (document_values);

CREATE INDEX documents_data_host_idx ON monitoring_draft_laws.documents_data USING btree (HOST);

---- create above / drop below ----
DROP TABLE monitoring_draft_laws.documents_data
