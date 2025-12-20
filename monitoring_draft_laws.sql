--
-- PostgreSQL database dump
--

\restrict XRhTb2PgPl45uqrzn1yG1l8HMriVJamGriEhG9m7JfIeD6e8CKPFcgMZCPQTceR

-- Dumped from database version 16.9
-- Dumped by pg_dump version 16.11 (Ubuntu 16.11-1.pgdg24.04+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: megaplan; Type: DATABASE; Schema: -; Owner: -
--

CREATE DATABASE megaplan WITH TEMPLATE = template0 ENCODING = 'UTF8' LOCALE_PROVIDER = libc LC_COLLATE = 'C' LC_CTYPE = 'C.UTF-8';


\unrestrict XRhTb2PgPl45uqrzn1yG1l8HMriVJamGriEhG9m7JfIeD6e8CKPFcgMZCPQTceR
\connect megaplan
\restrict XRhTb2PgPl45uqrzn1yG1l8HMriVJamGriEhG9m7JfIeD6e8CKPFcgMZCPQTceR

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: monitoring_draft_laws; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA monitoring_draft_laws;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: additional_files; Type: TABLE; Schema: monitoring_draft_laws; Owner: -
--

CREATE TABLE monitoring_draft_laws.additional_files (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    document_id text NOT NULL,
    meta_info jsonb DEFAULT '{}'::jsonb NOT NULL,
    object_id oid,
    created timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: cancellation_policy; Type: TABLE; Schema: monitoring_draft_laws; Owner: -
--

CREATE TABLE monitoring_draft_laws.cancellation_policy (
    host text NOT NULL,
    cancellation_phrases text[] DEFAULT '{}'::text[] NOT NULL
);


--
-- Name: documents_data; Type: TABLE; Schema: monitoring_draft_laws; Owner: -
--

CREATE TABLE monitoring_draft_laws.documents_data (
    projectid text NOT NULL,
    document_values jsonb NOT NULL,
    created timestamp with time zone DEFAULT now() NOT NULL,
    host text NOT NULL,
    updated timestamp with time zone DEFAULT now() NOT NULL,
    journal jsonb DEFAULT '[]'::jsonb NOT NULL,
    archive boolean DEFAULT false NOT NULL
);


--
-- Name: COLUMN documents_data.updated; Type: COMMENT; Schema: monitoring_draft_laws; Owner: -
--

COMMENT ON COLUMN monitoring_draft_laws.documents_data.updated IS 'автообноление';


--
-- Name: favorites; Type: TABLE; Schema: monitoring_draft_laws; Owner: -
--

CREATE TABLE monitoring_draft_laws.favorites (
    user_id integer NOT NULL,
    project_id text NOT NULL
);


--
-- Name: logs; Type: TABLE; Schema: monitoring_draft_laws; Owner: -
--

CREATE TABLE monitoring_draft_laws.logs (
    created timestamp with time zone DEFAULT now() NOT NULL,
    doc_id text NOT NULL,
    user_id integer NOT NULL,
    "values" jsonb DEFAULT '{}'::jsonb NOT NULL
);


--
-- Name: roles_settings; Type: TABLE; Schema: monitoring_draft_laws; Owner: -
--

CREATE TABLE monitoring_draft_laws.roles_settings (
    user_id integer NOT NULL,
    is_admin boolean DEFAULT false NOT NULL,
    is_responsible boolean DEFAULT false NOT NULL
);


--
-- Name: status_is_law; Type: VIEW; Schema: monitoring_draft_laws; Owner: -
--

CREATE VIEW monitoring_draft_laws.status_is_law AS
 SELECT projectid,
    true AS is_law
   FROM ( SELECT dd.projectid,
            jsonb_array_elements((dd.document_values -> 'journal'::text)) AS j
           FROM monitoring_draft_laws.documents_data dd
          WHERE (dd.host = 'sozd.duma.gov.ru'::text)) jj
  WHERE (((j ->> 'header'::text) = '5.3 Рассмотрение законопроекта Государственной Думой'::text) AND ((j ->> 'decision'::text) ~~* '%принять закон%'::text));


--
-- Name: tg_chats; Type: TABLE; Schema: monitoring_draft_laws; Owner: -
--

CREATE TABLE monitoring_draft_laws.tg_chats (
    chat_id bigint NOT NULL,
    enabled boolean NOT NULL,
    user_id integer DEFAULT 0 NOT NULL,
    notification_type smallint DEFAULT 0 NOT NULL
);


--
-- Name: COLUMN tg_chats.chat_id; Type: COMMENT; Schema: monitoring_draft_laws; Owner: -
--

COMMENT ON COLUMN monitoring_draft_laws.tg_chats.chat_id IS 'ID чата в телеграмм';


--
-- Name: COLUMN tg_chats.user_id; Type: COMMENT; Schema: monitoring_draft_laws; Owner: -
--

COMMENT ON COLUMN monitoring_draft_laws.tg_chats.user_id IS 'ID зера в мегаплане';


--
-- Name: COLUMN tg_chats.notification_type; Type: COMMENT; Schema: monitoring_draft_laws; Owner: -
--

COMMENT ON COLUMN monitoring_draft_laws.tg_chats.notification_type IS 'тип уведомления, фильтр';


--
-- Name: additional_files additional_files_pk; Type: CONSTRAINT; Schema: monitoring_draft_laws; Owner: -
--

ALTER TABLE ONLY monitoring_draft_laws.additional_files
    ADD CONSTRAINT additional_files_pk PRIMARY KEY (id);


--
-- Name: additional_files additional_files_unique; Type: CONSTRAINT; Schema: monitoring_draft_laws; Owner: -
--

ALTER TABLE ONLY monitoring_draft_laws.additional_files
    ADD CONSTRAINT additional_files_unique UNIQUE (object_id);


--
-- Name: cancellation_policy cancellation_policy_un; Type: CONSTRAINT; Schema: monitoring_draft_laws; Owner: -
--

ALTER TABLE ONLY monitoring_draft_laws.cancellation_policy
    ADD CONSTRAINT cancellation_policy_un UNIQUE (host);


--
-- Name: documents_data documents_data_pk; Type: CONSTRAINT; Schema: monitoring_draft_laws; Owner: -
--

ALTER TABLE ONLY monitoring_draft_laws.documents_data
    ADD CONSTRAINT documents_data_pk PRIMARY KEY (projectid);


--
-- Name: favorites favorites_un; Type: CONSTRAINT; Schema: monitoring_draft_laws; Owner: -
--

ALTER TABLE ONLY monitoring_draft_laws.favorites
    ADD CONSTRAINT favorites_un UNIQUE (user_id, project_id);


--
-- Name: roles_settings roles_settings_un; Type: CONSTRAINT; Schema: monitoring_draft_laws; Owner: -
--

ALTER TABLE ONLY monitoring_draft_laws.roles_settings
    ADD CONSTRAINT roles_settings_un UNIQUE (user_id);


--
-- Name: tg_chats tg_chats_pk; Type: CONSTRAINT; Schema: monitoring_draft_laws; Owner: -
--

ALTER TABLE ONLY monitoring_draft_laws.tg_chats
    ADD CONSTRAINT tg_chats_pk PRIMARY KEY (chat_id);


--
-- Name: additional_files_id_idx; Type: INDEX; Schema: monitoring_draft_laws; Owner: -
--

CREATE UNIQUE INDEX additional_files_id_idx ON monitoring_draft_laws.additional_files USING btree (id);


--
-- Name: cancellation_phrases_idx; Type: INDEX; Schema: monitoring_draft_laws; Owner: -
--

CREATE INDEX cancellation_phrases_idx ON monitoring_draft_laws.cancellation_policy USING gin (cancellation_phrases);


--
-- Name: document_valuess_idx; Type: INDEX; Schema: monitoring_draft_laws; Owner: -
--

CREATE INDEX document_valuess_idx ON monitoring_draft_laws.documents_data USING gin (document_values);


--
-- Name: documents_data_host_idx; Type: INDEX; Schema: monitoring_draft_laws; Owner: -
--

CREATE INDEX documents_data_host_idx ON monitoring_draft_laws.documents_data USING btree (host);


--
-- Name: logs_created_idx; Type: INDEX; Schema: monitoring_draft_laws; Owner: -
--

CREATE INDEX logs_created_idx ON monitoring_draft_laws.logs USING btree (created);


--
-- Name: logs_doc_id_idx; Type: INDEX; Schema: monitoring_draft_laws; Owner: -
--

CREATE INDEX logs_doc_id_idx ON monitoring_draft_laws.logs USING btree (doc_id);


--
-- Name: logs_user_id_idx; Type: INDEX; Schema: monitoring_draft_laws; Owner: -
--

CREATE INDEX logs_user_id_idx ON monitoring_draft_laws.logs USING btree (user_id);


--
-- Name: tg_chats_enabled_idx; Type: INDEX; Schema: monitoring_draft_laws; Owner: -
--

CREATE INDEX tg_chats_enabled_idx ON monitoring_draft_laws.tg_chats USING btree (enabled);


--
-- Name: tg_chats_notification_type_idx; Type: INDEX; Schema: monitoring_draft_laws; Owner: -
--

CREATE INDEX tg_chats_notification_type_idx ON monitoring_draft_laws.tg_chats USING btree (notification_type);


--
-- Name: tg_chats_user_id_idx; Type: INDEX; Schema: monitoring_draft_laws; Owner: -
--

CREATE INDEX tg_chats_user_id_idx ON monitoring_draft_laws.tg_chats USING btree (user_id);


--
-- Name: additional_files additional_files_documents_data_fk; Type: FK CONSTRAINT; Schema: monitoring_draft_laws; Owner: -
--

ALTER TABLE ONLY monitoring_draft_laws.additional_files
    ADD CONSTRAINT additional_files_documents_data_fk FOREIGN KEY (document_id) REFERENCES monitoring_draft_laws.documents_data(projectid) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: favorites favorites_fk; Type: FK CONSTRAINT; Schema: monitoring_draft_laws; Owner: -
--

ALTER TABLE ONLY monitoring_draft_laws.favorites
    ADD CONSTRAINT favorites_fk FOREIGN KEY (project_id) REFERENCES monitoring_draft_laws.documents_data(projectid) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: logs logs_fk; Type: FK CONSTRAINT; Schema: monitoring_draft_laws; Owner: -
--

ALTER TABLE ONLY monitoring_draft_laws.logs
    ADD CONSTRAINT logs_fk FOREIGN KEY (doc_id) REFERENCES monitoring_draft_laws.documents_data(projectid) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: roles_settings roles_settings_fk; Type: FK CONSTRAINT; Schema: monitoring_draft_laws; Owner: -
--

ALTER TABLE ONLY monitoring_draft_laws.roles_settings
    ADD CONSTRAINT roles_settings_fk FOREIGN KEY (user_id) REFERENCES public.users(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: tg_chats tg_chats_users_fk; Type: FK CONSTRAINT; Schema: monitoring_draft_laws; Owner: -
--

ALTER TABLE ONLY monitoring_draft_laws.tg_chats
    ADD CONSTRAINT tg_chats_users_fk FOREIGN KEY (user_id) REFERENCES public.users(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict XRhTb2PgPl45uqrzn1yG1l8HMriVJamGriEhG9m7JfIeD6e8CKPFcgMZCPQTceR

