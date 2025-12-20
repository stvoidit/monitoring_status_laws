import type { AxiosError, AxiosResponse, AxiosHeaders } from "axios";
import { LawDocument, LogRow, MegaplanUser } from "./models";
import { type UploadFile } from "element-plus";
import { MessageAlert } from "@/composibles/useAlert";
import axios from "axios";
import qs from "qs";
import { reactive } from "vue";

function downloadBlobFile(filename: string, response: AxiosResponse<Blob>) {
    const headers = response.headers as AxiosHeaders;
    const file = window.URL.createObjectURL(response.data);
    const link = document.createElement("a");
    link.setAttribute("href", file);
    if (!headers.has("content-disposition", /filename.*/)) {
        link.setAttribute("download", filename);
    }
    document.body.appendChild(link);
    link.click();
    link.remove();
    window.URL.revokeObjectURL(file);
}

class AppUser {
    id: number;
    name: string;
    position: string;
    department: unknown;
    constructor(json: Partial<AppUser>) {
        Object.entries(json).forEach(([ key, value ]) => {
            if (Object.hasOwn(this, key)) {
                this[key as keyof typeof AppUser] = value;
            }
        });
    }

    static fromJSON(json: Partial<Omit<AppUser, "id">> & { id?: number | string } = {}) {
        if (json.id) json.id = Number(json.id);
        return new AppUser(json as AppUser);
    }
}

class AppInfo {
    appUUID = "";
    megaplanDomain = "";
    isAdmin = false;
    isResponsible = false;
    tg_bot_name = "";
    ntype = 0;
    token?: string;
    user?: AppUser;
    constructor(json: Partial<AppInfo>) {
        Object.entries(json).forEach(([ key, value ]) => {
            if (Object.hasOwn(this, key)) {
                this[key as keyof typeof AppInfo] = value;
            }
        });
    }

    static fromJSON(json: Partial<AppInfo>) {
        if (json.user) json.user = AppUser.fromJSON(json.user);
        return new AppInfo(json);
    }
}

const alertMessage = (message?: string) => {
    message ??= "Произошла ошибка, обратитесь к администратору или повторите позже";
    MessageAlert(message);
};

interface ResponseErrorBody {
    appUUID: string;
    megaplanDomain: string;
    error?: string;
}

const onError = (error: AxiosError<ResponseErrorBody>) => {
    // если статус 401, то 99% пользователь открыл окно не внутри мегаплана и не авторизован
    // просто открываем ему окно сразу на страницу приложения и не паримся
    if (error.response && [ 401, 403 ].includes(error.response.status)) {
        window.localStorage.removeItem("token");
        const { appUUID, megaplanDomain } = error.response.data;
        const applicationPageURL = `${megaplanDomain}/application/page/${appUUID}`;
        const el = window.document.createElement("a");
        el.setAttribute("href", applicationPageURL);
        el.textContent = applicationPageURL;
        document.body.textContent = "Для авторизации перейдите по ссылке в SSR: ";
        document.body.appendChild(el);
    } else {
        alertMessage();
    }
};

const createClient = () => {
    const client = axios.create({
        baseURL: "/api/",
        adapter: "fetch",
        timeout: 5 * 60 * 1000,
        paramsSerializer: params => qs.stringify(params, { indices: false, skipNulls: true }),
        timeoutErrorMessage: "Сервер не отвечает, попробуйте позже",
    });
    client.interceptors.response.use(
        response => response,
        (error) => {
            // console.error(error);
            if (axios.isCancel(error)) {
                return error;
            }
            if (axios.isAxiosError<ResponseErrorBody>(error)) {
                onError(error);
            } else {
                alertMessage();
            }
        });
    const token = window.localStorage.getItem("token");
    if (token) {
        client.defaults.headers.common.Authorization = `bearer ${token}`;
    }
    return client;
};

export const appInfo = reactive({
    isAdmin: false,
    isResponsible: false,
    tg_bot_name: "",
} as AppInfo);

const client = createClient();

/** Инициализация клиента, авторизация и аутентификация пользователя, получение данных приложения */
export async function init() {
    const params = new URLSearchParams(location.search);
    const response = await client.get<AppInfo>("/init", { params, timeout: 5000 });
    const info = AppInfo.fromJSON(response.data);
    Object.assign(appInfo, info);
    if (appInfo.token) {
        window.localStorage.setItem("token", appInfo.token);
    }
}

/** Получение списка документов */
export async function GetDocuments() {
    const response = await client.get<LawDocument[]>("/documents");
    return response.data.map(v => LawDocument.fromJSON(v));
}

/** Добавление документа */
export async function AddNewDocument(data: unknown, { signal }: { signal?: AbortSignal }) {
    return await client.post<LawDocument>("/documents", data, { signal }).then(r => r.data);
}

/** Получить документ по ID */
export async function GetDocument(id: string) {
    return await client.get<LawDocument>("/document", { params: { id } }).then(r => LawDocument.fromJSON(r.data));
}

/** Обновить документ */
export async function UpdateDocument(doc: LawDocument) {
    return await client.put<LawDocument>("/document", doc).then(r => LawDocument.fromJSON(r.data));
}

/** Удалить документ */
export async function DeleteDocument(id: string) {
    return await client.delete("/document", { params: { id } });
}

/** Обновить документ из источника */
export async function FetchUpdate(doc: LawDocument) {
    return await client.post<LawDocument>("/document/fetch", doc).then(r => LawDocument.fromJSON(r.data));
}

/** Обновить все документы из источников - фоновая задача бэка */
export async function FetchUpdateAll() {
    return await client.post<{ message: string }>("/documents/fetch").then(response => response.data);
}

/** Скачать документы в excel */
export async function DownloadDocuments(data) {
    const response = await client.post<Blob>("/documents/download", data, { responseType: "blob" });
    const filename = "документы.xlsx";
    downloadBlobFile(filename, response);
}

/** Получить список пользователей для настройки ролей */
export async function SelectUsers() {
    return await client.get<MegaplanUser[]>("/roles_settings").then(
        response => response.data.map(u => MegaplanUser.fromJSON(u)),
    );
}

/** Изменить роль пользователя */
export async function ChangeRole(user: MegaplanUser) {
    return await client.post<null>("/roles_settings", user);
}

/** Добавить в избранное */
export async function DoFavorite(project_id: string, is_favorite: boolean) {
    return await client.put("/favorite", { project_id, is_favorite });
}

/** Патч черновика - установить оригинальынй ID документа */
export async function PatchDraftDocument(payload: unknown) {
    return await client.patch("/document", payload);
}

/** Изменить тип уведомления */
export async function ChangeNotificationType(ntype: number) {
    return await client.post("/ntype", { ntype });
}

/** Логи изменений юзером */
export async function GetLogs(id: string) {
    return await client.get<LogRow[]>("/changeslogs", { params: { id } }).then(
        response => response.data.map(l => LogRow.fromJSON(l)),
    );
}

export async function GetFiles(id: string) {
    return await client.get<UploadFile[]>("/files", { params: { id } }).then(response => response.data);
}

export async function DeleteFile(id: string) {
    return await client.delete("/file", { params: { id } });
}

export async function ChangeArchiveStatus(id: string) {
    return await client.patch("/archive", { id });
}

export const downloadExternalFile = async ({ id, source, name }: { id: string; source: string; name: string }) => {
    const response = await client.get<Blob>("/proxy_download", {
        responseType: "blob",
        params: { id, source, name },
    });
    downloadBlobFile(name, response);
};

export const isCanceledError = axios.isCancel;
