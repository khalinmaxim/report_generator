CREATE ROLE accessors;
CREATE ROLE reporter;

CREATE SCHEMA reporter
    AUTHORIZATION reporter;

COMMENT
ON SCHEMA reporter
    IS ''Схема сервиса создания отчётов
1.0.0
'';

GRANT USAGE ON SCHEMA
reporter TO accessors;

GRANT ALL
ON SCHEMA reporter TO reporter;

CREATE TABLE reporter.templates
(
    entity_uuid      uuid                     NOT NULL DEFAULT uuid_generate_v4(),
    entity_created   timestamp with time zone NOT NULL DEFAULT now(),
    report_title     text COLLATE pg_catalog."default",
    report_query     text COLLATE pg_catalog."default",
    report_access    smallint                 NOT NULL DEFAULT 0,
    template_content bytea,
    report_fill      jsonb,
    template_name    text COLLATE pg_catalog."default",
    report_path      text COLLATE pg_catalog."default",
    CONSTRAINT templates_pkey PRIMARY KEY (entity_uuid)
)
    WITH (
        OIDS = FALSE
        )
    TABLESPACE pg_default;

ALTER TABLE reporter.templates
    OWNER to reporter;

GRANT SELECT ON TABLE reporter.templates TO accessors;

GRANT
ALL
ON TABLE reporter.templates TO reporter;

GRANT INSERT, SELECT, UPDATE ON TABLE reporter.templates TO reporters;


COMMENT
ON TABLE reporter.templates
    IS ''Шаблоны отчётов'';

COMMENT
ON COLUMN reporter.templates.report_title
    IS ''Наименование отчёта'';

COMMENT
ON COLUMN reporter.templates.report_query
    IS ''Запрос'';

COMMENT
ON COLUMN reporter.templates.report_access
    IS ''Группа доступа'';

COMMENT
ON COLUMN reporter.templates.report_fill
    IS ''Заполнение полей'';

COMMENT
ON COLUMN reporter.templates.template_name
    IS ''Имя файла шаблона'';

COMMENT
ON COLUMN reporter.templates.report_path
    IS ''Путь сохранения'';


CREATE TABLE reporter.reports
(
    -- Inherited from table observation.entity: entity_uuid uuid NOT NULL DEFAULT uuid_generate_v4(),
    -- Inherited from table observation.entity: entity_created timestamp with time zone NOT NULL DEFAULT now(),
    report_template uuid,
    report_content  bytea,
    report_title    text COLLATE pg_catalog."default",
    CONSTRAINT reports_pkey PRIMARY KEY (entity_uuid),
    CONSTRAINT reports_template_fkey FOREIGN KEY (report_template)
        REFERENCES reporter.templates (entity_uuid) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
) INHERITS (observation.entity)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE reporter.reports
    OWNER to reporter;

GRANT
ALL
ON TABLE reporter.reports TO reporter;

GRANT INSERT, SELECT, UPDATE ON TABLE reporter.reports TO reporters;

COMMENT
ON TABLE reporter.reports
    IS ''Готовые отчёты'';
-- Index: fki_reports_template_fkey

-- DROP INDEX reporter.fki_reports_template_fkey;

CREATE INDEX fki_reports_template_fkey
    ON reporter.reports USING btree
    (report_template ASC NULLS LAST)
    TABLESPACE pg_default;
-- Index: reports_created_idx

-- DROP INDEX reporter.reports_created_idx;

CREATE INDEX reports_created_idx
    ON reporter.reports USING btree
    (entity_created ASC NULLS LAST)
    TABLESPACE pg_default;


CREATE TABLE reporter.queue
(
    queue_object  uuid,
    report_params jsonb,
    entity_uuid   uuid NOT NULL DEFAULT uuid_generate_v4(),
    CONSTRAINT queue_pkey PRIMARY KEY (entity_uuid),
    CONSTRAINT queue_template_fkey FOREIGN KEY (queue_object)
        REFERENCES reporter.templates (entity_uuid) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
    WITH (
        OIDS = FALSE
        )
    TABLESPACE pg_default;

ALTER TABLE reporter.queue
    OWNER to reporter;

GRANT
ALL
ON TABLE reporter.queue TO reporter;

GRANT DELETE, INSERT, SELECT, UPDATE ON TABLE reporter.queue TO reporters;

COMMENT
ON TABLE reporter.queue
    IS ''Очередь заданий отчётов'';

COMMENT
ON COLUMN reporter.queue.queue_object
    IS ''Шаблон'';

COMMENT
ON COLUMN reporter.queue.report_params
    IS ''Параметры отчёта'';
-- Index: fki_queue_template_fkey

-- DROP INDEX reporter.fki_queue_template_fkey;

CREATE INDEX fki_queue_template_fkey
    ON reporter.queue USING btree
    (queue_object ASC NULLS LAST)
    TABLESPACE pg_default;


CREATE
OR REPLACE FUNCTION reporter.add_queue(
	a_tmp uuid,
	a_params jsonb)
    RETURNS uuid
    LANGUAGE ''plpgsql''
    COST 100
    VOLATILE PARALLEL UNSAFE
AS $BODY$
DECLARE
result uuid;
BEGIN

WITH t AS (
    SELECT *
    FROM reporter.templates
    WHERE entity_uuid = a_tmp
    LIMIT 1
    )
INSERT
INTO reporter.queue
    (queue_object, report_params)
VALUES
    (a_tmp, jsonb_build_object(
    '' query '', a_params,
    '' title '', replace((SELECT report_title FROM t LIMIT 1), '' '', ''_'')
    || ''_'' || to_char(now(), '' DD.MM.YY '')
    || ''.xlsx ''
    ))
    RETURNING entity_uuid
INTO result;

RETURN result;
END
$BODY$;

ALTER FUNCTION reporter.add_queue(uuid, jsonb)
    OWNER TO reporter;

COMMENT
ON FUNCTION reporter.add_queue(uuid, jsonb)
    IS ''Добавление задания'';


CREATE
OR REPLACE FUNCTION reporter.add_report(
	a_template uuid,
	a_title text,
	a_content bytea)
    RETURNS uuid
    LANGUAGE ''plpgsql''
    COST 100
    VOLATILE PARALLEL UNSAFE
AS $BODY$
DECLARE
result uuid;
BEGIN

INSERT INTO reporter.reports
    (report_template, report_content, report_title)
VALUES (a_template, a_content, a_title) RETURNING entity_uuid
INTO result;

RETURN result;

END
$BODY$;

ALTER FUNCTION reporter.add_report(uuid, text, bytea)
    OWNER TO reporter;

COMMENT
ON FUNCTION reporter.add_report(uuid, text, bytea)
    IS ''Добавление отчёта'';

CREATE
OR REPLACE FUNCTION reporter.add_template(
	a_id uuid,
	a_title text,
	a_query text,
	a_access smallint,
	a_template text,
	a_content bytea,
	a_fill jsonb,
	a_path text)
    RETURNS uuid
    LANGUAGE ''plpgsql''
    COST 100
    VOLATILE PARALLEL UNSAFE
AS $BODY$
DECLARE
result uuid;
BEGIN

IF
a_id IS NULL THEN
	INSERT INTO reporter.templates
	(report_title,report_query,report_access,template_content,template_name,report_fill,report_path)
	VALUES
	(a_title,a_query,a_access,a_content,a_template,a_fill,a_path)
	RETURNING entity_uuid INTO result;
ELSE
UPDATE reporter.templates
SET report_title=a_title,
    report_query=a_query,
    report_access=a_access,
    template_content=coalesce(a_content, template_content),
    template_name=coalesce(a_template, template_name),
    report_fill=a_fill,
    report_path=a_path
WHERE entity_uuid = a_id;
END IF;

RETURN coalesce(result, a_id);

END
$BODY$;

ALTER FUNCTION reporter.add_template(uuid, text, text, smallint, text, bytea, jsonb, text)
    OWNER TO reporter;

COMMENT
ON FUNCTION reporter.add_template(uuid, text, text, smallint, text, bytea, jsonb, text)
    IS ''Добавление шаблона'';