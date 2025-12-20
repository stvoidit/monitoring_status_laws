import dayjs from "dayjs";
const DateTimeHumanFormat = "DD.MM.YYYY HH:mm";

type SourceType = "regulation.gov.ru" | "sozd.duma.gov.ru";

interface ExternalFiles {
    label: string;
    files: {
        id: string;
        name: string;
        href: string;
    }[];
}

/** User - пользователь, базовая сущность */
export class User {
    /** ID пользователя */
    id: number;
    /** ФИО пользователя полностью */
    fio: string;
    /** ФИО пользователя коротко "Фамилия И.О." */
    shortname: string;
    constructor(json: Partial<User>) {
        Object.entries(json).forEach(([ key, value ]) => {
            if (Object.hasOwn(this, key)) {
                this[key as keyof typeof User] = value;
            }
        });
    }

    static fromJSON(json: Partial<User>) {
        return new User(json);
    }
}

/** MegaplanUser - пользователь мегаплана */
export class MegaplanUser extends User {
    /** ID отдела */
    department_id: number;
    /** Название отдела */
    department_label: string;
    /** Должность */
    position: string;
    /** Является админом */
    is_admin: boolean;
    /** Является ответственным */
    is_responsible: boolean;
    constructor(json: Partial<MegaplanUser>) {
        super(json);
        Object.entries(json).forEach(([ key, value ]) => {
            if (Object.hasOwn(this, key)) {
                this[key as keyof typeof MegaplanUser] = value;
            }
        });
    }

    static fromJSON(json: Partial<MegaplanUser>) {
        return new MegaplanUser(json);
    }
}

/** JournalRow - запись журнала изменений */
export class JournalRow {
    /** Дата записи */
    date: Date;
    /** Статус */
    header: string;
    /** Описание */
    decision: string;
    constructor(json: Partial<JournalRow>) {
        Object.entries(json).forEach(([ key, value ]) => {
            if (Object.hasOwn(this, key)) {
                this[key as keyof typeof JournalRow] = value;
            }
        });
    }

    static fromJSON(json: Partial<JournalRow>) {
        if (json.date) json.date = dayjs(json.date, "DD.MM.YYYY").toDate();
        return new JournalRow(json);
    }

    get Date() {
        return dayjs(this.date).format("DD.MM.YYYY");
    }
}

/** LawDocument - документ законопроекта */
export class LawDocument {
    /** ID документа из источника */
    id: string;
    /** Проект */
    project: string;
    /** Название проекта */
    label: string;
    /** краткое наименование */
    short_label: string;
    /** Источник */
    source: SourceType;
    /** Дата документа */
    date: Date;
    /** Организация разработчик */
    department: string;
    /** Вид проекта НПА */
    kind: string;
    /** Область регулирования */
    scope: string;
    /** текущий этап */
    current_stage: string;
    /** текущий статус */
    status: string;
    /** Краткое содержание */
    desc: string;
    /** Примечания */
    note: string;
    /** Вид налога (сбора) */
    tax_type: string;
    /** Последнее обновление */
    updated: Date;
    /** Журнал изменений */
    journal: JournalRow[];
    /** Отменен */
    is_cancelled: boolean;
    /** В избранном пользователя */
    is_favorite: boolean;
    /** Ссылка на задачу в ССР */
    task_id: string;
    /** Номер в ЭДО */
    number_edo: string;
    /** Приоритет */
    priority: number;
    /** Актуализированный статус */
    actual_status: string;
    /** Черновик */
    is_draft: boolean;
    /** Закон или Законопроект */
    is_law: boolean;
    /** проект в архиве */
    archive: boolean;
    /** файлы */
    files?: Record<string, { id: string; name: string; href: string }[]>;

    constructor(json: Partial<LawDocument>) {
        Object.entries(json).forEach(([ key, value ]) => {
            if (Object.hasOwn(this, key)) {
                this[key as keyof typeof LawDocument] = value;
            }
        });
    }

    static fromJSON(json: Partial<LawDocument>) {
        if (json.date) json.date = dayjs(json.date).toDate();
        if (json.updated) json.date = dayjs(json.updated).toDate();
        if (json.journal) json.journal = json.journal.map(value => JournalRow.fromJSON(value));
        return new LawDocument(json);
    }

    /** Список полей для отрисовки таблицы */
    static TableFieldsList() {
        const fieldsSettings: ITableField[] = [
            { label: "Избранное", prop: "is_favorite", show: true, minWidth: 120, sortable: false, align: "center" },
            { label: "Идентификатор", prop: "project", show: true, minWidth: 260, sortable: false },
            { label: "Название", prop: "label", show: true, minWidth: 200, sortable: false },
            { label: "Краткое название", prop: "short_label", show: true, minWidth: 200, sortable: false },
            { label: "Источник", prop: "source", show: true, minWidth: 160, sortable: false },
            { label: "Приоритет", prop: "priority", show: true, minWidth: 130, sortable: true, align: "center" },
            { label: "Этап рассмотрения", prop: "ReviewStage", show: true, minWidth: 200, sortable: "custom" },
            { label: "Последнее событие", prop: "LastEventHeader", show: true, minWidth: 220, sortable: "custom" },
            { label: "Дата последнего события", prop: "changed", show: true, minWidth: 180, sortable: "custom", align: "center" },
            { label: "Статус", prop: "status", show: true, minWidth: 160, sortable: "custom" },
            { label: "Актуализированный статус", prop: "actual_status", show: true, minWidth: 200, sortable: false },
            { label: "Процент рассмотрения", prop: "percent", show: true, minWidth: 180, sortable: "custom", align: "center" },
            { label: "Краткое содержание", prop: "desc", show: true, minWidth: 200, sortable: false },
            { label: "Дата создания", prop: "Date", show: true, minWidth: 180, sortable: "custom", align: "center" },
            { label: "Разработчик", prop: "department", show: true, minWidth: 140, sortable: false },
            { label: "Вид проекта НПА", prop: "kind", show: true, minWidth: 200, sortable: "custom" },
            { label: "Вид налога (сбора)", prop: "tax_type", show: true, minWidth: 160, sortable: false },
            { label: "Область регулирования", prop: "scope", show: true, minWidth: 160, sortable: false },
            { label: "Примечания", prop: "note", show: true, minWidth: 200, sortable: false },
            { label: "Последняя проверка", prop: "Updated", show: true, minWidth: 200, sortable: "custom" },
            { label: "Ссылка на задачу в ССР", prop: "task_id", show: true, minWidth: 300, sortable: false },
            { label: "Номер в ЭДО", prop: "number_edo", show: true, minWidth: 180, sortable: false },
            { label: "Вид документа", prop: "LawType", show: true, minWidth: 160, sortable: false },
            // { label: "Особая важность", prop: "Important", show: true, minWidth: 100, sortable: false },
            // { label:"Последнее изменение", prop: "current_stage", show: true, minWidth: 100, sortable: false },
        ] as const;
        return fieldsSettings.map((f) => {
            f.align ??= "left";
            return f;
        });
    }

    get Date() {
        // небольшой костыль, т.к. в оригинальной структуре данных эта дата обязательна, но в черновиках отсутствует
        const d = dayjs(this.date);
        if (d.unix() <= 0) {
            return "";
        }
        return d.format(DateTimeHumanFormat);
    }

    get Updated() {
        return dayjs(this.updated).format(DateTimeHumanFormat);
    }

    /** Последнее изменение из журнала */
    get last_change() {
        if (!this.journal.length) return "";
        else return this.journal[this.journal.length - 1].header;
    }

    get changed() {
        if (!this.journal.length) return "";
        else return dayjs(this.journal[this.journal.length - 1].date).format("DD.MM.YYYY");
    }

    get LastChangeDate(): Date | null {
        if (!this.journal.length) return null;
        else return this.journal[this.journal.length - 1].date;
    }

    get DocumentSourceURL() {
        switch (this.source) {
            case "regulation.gov.ru":
                return `https://regulation.gov.ru/projects/${this.id}`;
            case "sozd.duma.gov.ru":
                return `https://sozd.duma.gov.ru/bill/${this.id}`;
            default:
                return "";
        }
    }

    get LawType() {
        return this.is_law ? "Закон" : "Законопроект";
    }

    /** расчет процента прогресса рассмотрения документа */
    get percent(): string {
        let perc = "";
        if (this.source === "regulation.gov.ru") {
            const percMap = {
                "Не определен": "0%",
                "Уведомление": "10%",
                "Текст": "40%",
                "Оценка": "60%",
                "Завершение": "70%",
                "Принятие": "100%",
            };
            perc = percMap[this.current_stage] ?? "";
        }
        if (this.source === "sozd.duma.gov.ru") {
            if (this.current_stage.startsWith("1.")) {
                perc = "10%";
            } else if (this.current_stage.startsWith("2.")) {
                perc = "20%";
            } else if (this.current_stage.startsWith("3.")) {
                perc = "40%";
            } else if (this.current_stage.startsWith("4.")) {
                perc = "60%";
            } else if (this.current_stage.startsWith("5.")) {
                perc = "70%";
            } else if (this.current_stage.startsWith("6.")
                || this.current_stage.startsWith("7.")
                || this.current_stage.startsWith("9.")) {
                perc = "80%";
            } else if (this.current_stage.startsWith("8.1")) {
                perc = "90%";
            } else if (this.current_stage.startsWith("8.2")) {
                perc = "100%";
            }
        }
        return perc;
    }

    /** Стадия рассмотрения */
    get ReviewStage() {
        const stages = [
            { perc: "10%", name: "Внесен в Госдуму" },
            { perc: "20%", name: "Предварительное рассмотрение" },
            { perc: "40%", name: "Первое чтение" },
            { perc: "60%", name: "Второе чтение" },
            { perc: "70%", name: "Третье чтение" },
            { perc: "80%", name: "Совет Федерации" },
            { perc: "90%", name: "Повторное рассмотрение" },
            { perc: "100%", name: "Президент РФ" },
        ];
        switch (this.source) {
            case "regulation.gov.ru":
                return this.current_stage;
                break;
            case "sozd.duma.gov.ru":
                const stg = stages.find(s => s.perc === this.percent);
                return stg?.name ?? "";
            default:
                return "";
        }
    }

    get LastEventHeader() {
        if (this.journal.length) {
            return this.journal[this.journal.length - 1].header;
        }
        return "";
    }

    /** Список значений полей документа для выгрузки в excel */
    ArrayValues(index: number, props: string[]) {
        const arr: string[] = [];
        for (const prop of props) {
            if (prop === "#") {
                arr.push(index.toString());
                continue;
            }
            const value = this[prop];
            if (value === null) arr.push("");
            else {
                if (value instanceof Date) {
                    arr.push(dayjs(value).format("DD.MM.YYYY"));
                } else if (typeof value == "boolean") {
                    arr.push(value ? "да" : "нет");
                } else {
                    arr.push(value.toString() as string);
                }
            }
        }
        return arr;
    }

    get Files2read(): ExternalFiles | undefined {
        if (this.source === "sozd.duma.gov.ru") {
            const rgx = new RegExp(/текст.*втор/i);
            for (const files of Object.values(this.files ?? {})) {
                const searchFile2read = files.filter(file => rgx.test(file.name));
                if (searchFile2read.length) {
                    return {
                        label: "Рассмотрение законопроекта во втором чтении",
                        files: searchFile2read,
                    };
                }
            }
        }
    }

    get Files3read(): ExternalFiles | undefined {
        if (this.source === "sozd.duma.gov.ru") {
            const rgx = new RegExp(/текст.*треть/i);
            for (const files of Object.values(this.files ?? {})) {
                const searchFile3read = files.filter(file => rgx.test(file.name));
                if (searchFile3read.length) {
                    return {
                        label: "Рассмотрение законопроекта в третьем чтении",
                        files: searchFile3read,
                    };
                }
            }
        }
    }

    get FilesAll() {
        const arr: ExternalFiles[] = [];
        if (this.source === "sozd.duma.gov.ru") return arr;
        for (const [ key, files ] of Object.entries(this.files ?? {})) {
            arr.push({
                label: key,
                files: files,
            });
        }
        arr.sort((a, b) => b.label.localeCompare(a.label));
        return arr;
    }
}

const fieldsNames = new Map(Object.entries({
    short_label: "Краткое наименование",
    scope: "Область регулирования",
    desc: "Краткое содержание",
    note: "Примечания",
    tax_type: "Вид налога (сбора)",
    priority: "Приоритет",
    task_id: "Ссылка на задачу в ССР",
    number_edo: "Номер в ЭДО",
    actual_status: "Актуализированный статус",
    is_draft: "Черновик",
    archive: "Архив",
}));

export class LogRow {
    created: dayjs.Dayjs;
    user: User;
    changes: Record<string, {
        before?: object;
        after?: object;
    }>;

    constructor(json: Partial<LogRow>) {
        Object.entries(json).forEach(([ key, value ]) => {
            if (Object.hasOwn(this, key)) {
                this[key as keyof typeof LogRow] = value;
            }
        });
    }

    static fromJSON(json: Partial<LogRow>) {
        if (json.created) json.created = dayjs(json.created);
        if (json.user) json.user = User.fromJSON(json.user);
        return new LogRow(json);
    }

    get Created() {
        return this.created.format("DD.MM.YYYY HH:mm:ss");
    }

    get FieldName() {
        const fields = Object.keys(this.changes);
        return fields.length ? fields[0] : "";
    }

    get ReadebleFieldName() {
        return fieldsNames.get(this.FieldName);
    }

    get After() {
        if (this.FieldName === "archive") {
            return this.changes[this.FieldName].after ? "Отправил в архив" : "Вернул из архива";
        }
        return this.changes[this.FieldName].after;
    }

    get Before() {
        if (this.FieldName === "archive") {
            return this.changes[this.FieldName].before ? "Отправил в архив" : "Вернул из архива";
        }
        return this.changes[this.FieldName].before;
    }
}

export type FilterType = "total"
    | "completed"
    | "oncontrol"
    | "cancelled"
    | "Important"
    | "favorites"
    | "draft"
    | "archive";

/** параметры для фильтра таблицы */
export interface IFilters {
    source: string;
    scope: string;
    dates: [Date, Date] | [string, string] | undefined;
    types: FilterType;
    currentPage: number;
    pageSize: number;
    search: string;
    lawtype: string;
}

export interface ITableField {
    label: string;
    prop: string;
    show: boolean;
    minWidth: number;
    sortable?: boolean | string;
    align?: "left" | "center";
}
