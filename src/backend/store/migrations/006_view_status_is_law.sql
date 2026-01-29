CREATE VIEW
    monitoring_draft_laws.status_is_law AS
SELECT
    projectid,
    TRUE      AS is_law
FROM
    (
        SELECT
            dd.projectid,
            JSONB_ARRAY_ELEMENTS((dd.document_values -> 'journal'::TEXT)) AS j
        FROM
            monitoring_draft_laws.documents_data dd
        WHERE
            (dd.host = 'sozd.duma.gov.ru'::TEXT)
    ) jj
WHERE
    (
        (
            (j ->> 'header'::TEXT) = '5.3 Рассмотрение законопроекта Государственной Думой'::TEXT
        )
        AND (
            (j ->> 'decision'::TEXT) ~~* '%принять закон%'::TEXT
        )
    );

---- create above / drop below ----
DROP VIEW monitoring_draft_laws.status_is_law
