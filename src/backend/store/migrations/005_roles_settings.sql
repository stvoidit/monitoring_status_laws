CREATE TABLE
    monitoring_draft_laws.roles_settings (
        user_id INTEGER NOT NULL,
        is_admin BOOLEAN DEFAULT FALSE NOT NULL,
        is_responsible BOOLEAN DEFAULT FALSE NOT NULL
    );

ALTER TABLE ONLY monitoring_draft_laws.roles_settings
ADD CONSTRAINT roles_settings_un UNIQUE (user_id);

ALTER TABLE ONLY monitoring_draft_laws.roles_settings
ADD CONSTRAINT roles_settings_fk FOREIGN KEY (user_id) REFERENCES public.users (id) ON UPDATE CASCADE ON DELETE CASCADE;

INSERT INTO
    monitoring_draft_laws.roles_settings (user_id, is_admin, is_responsible)
VALUES
    (1000129, TRUE, TRUE);

---- create above / drop below ----
DROP TABLE monitoring_draft_laws.roles_settings
