CREATE TABLE
    monitoring_draft_laws.tg_chats (
        chat_id BIGINT NOT NULL,
        enabled BOOLEAN NOT NULL,
        user_id INTEGER DEFAULT 0 NOT NULL,
        notification_type SMALLINT DEFAULT 0 NOT NULL
    );

ALTER TABLE ONLY monitoring_draft_laws.tg_chats
ADD CONSTRAINT tg_chats_pk PRIMARY KEY (chat_id);

ALTER TABLE ONLY monitoring_draft_laws.tg_chats
ADD CONSTRAINT tg_chats_users_fk FOREIGN KEY (user_id) REFERENCES public.users (id) ON UPDATE CASCADE ON DELETE CASCADE;

CREATE INDEX tg_chats_enabled_idx ON monitoring_draft_laws.tg_chats USING btree (enabled);

CREATE INDEX tg_chats_notification_type_idx ON monitoring_draft_laws.tg_chats USING btree (notification_type);

CREATE INDEX tg_chats_user_id_idx ON monitoring_draft_laws.tg_chats USING btree (user_id);

---- create above / drop below ----
DROP TABLE monitoring_draft_laws.tg_chats
