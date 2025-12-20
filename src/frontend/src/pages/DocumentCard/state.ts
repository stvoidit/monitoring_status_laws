import {
    appInfo,
    GetDocument,
    GetLogs,
    UpdateDocument,
    FetchUpdate,
    DeleteDocument,
    PatchDraftDocument,
} from "@/api";
import type { LawDocument, LogRow } from "@/api/models";
import { ref, computed, shallowRef } from "vue";
import { isAxiosError } from "axios";
import { useRouter } from "vue-router";
import { ElNotification, ElMessageBox } from "element-plus";

export default function useState(documentId: string) {
    const documentFields = {
        baseInformation: [
            { name: "Название", prop: "label", editable: false, span1: 6, span2: 18 },
            { name: "Дата создания", prop: "Date", editable: false, span1: 6, span2: 18 },
            { name: "Идентификатор", prop: "project", editable: false, span1: 6, span2: 18 },
            { name: "Источник", prop: "source", editable: false, span1: 6, span2: 18 },
            { name: "Разработчик", prop: "department", editable: false, span1: 6, span2: 18 },
            { name: "Вид проекта НПА", prop: "kind", editable: false, span1: 6, span2: 18 },
        ],
        lawReview: [
            { name: "Этап рассмотрения", prop: "ReviewStage", editable: false, span1: 7, span2: 17 },
            { name: "Процент рассмотрения", prop: "percent", editable: false, span1: 7, span2: 17 },
            { name: "Последнее событие", prop: "LastEventHeader", editable: false, span1: 7, span2: 17 },
            { name: "Статус", prop: "status", editable: false, span1: 7, span2: 17 },
            { name: "Дата последнего события", prop: "changed", editable: false, span1: 7, span2: 17 },
            { name: "Актуализированный статус", prop: "actual_status", editable: true, span1: 7, span2: 17 },
        ],
        additionInformation: [
            { name: "Краткое название", prop: "short_label", isSelect: false, isDatepicker: false, span1: 4, span2: 20 },
            { name: "Краткое содержание", prop: "desc", isSelect: false, isDatepicker: false, span1: 4, span2: 20 },
            { name: "Вид налога (сбора)", prop: "tax_type", isSelect: false, isDatepicker: false, span1: 4, span2: 20 },
            { name: "Область регулирования", prop: "scope", isSelect: false, isDatepicker: false, span1: 4, span2: 20 },
            { name: "Приоритет", prop: "priority", isSelect: true, isDatepicker: false, span1: 4, span2: 20 },
            { name: "Ссылка на задачу в ССР", prop: "task_id", isSelect: false, isDatepicker: false, span1: 4, span2: 20 },
            { name: "Номер в ЭДО", prop: "number_edo", isSelect: false, isDatepicker: false, span1: 4, span2: 20 },
            { name: "Примечания", prop: "note", isSelect: false, span1: 4, isDatepicker: false, span2: 20 },
        ],
    } as const;

    const router = useRouter();
    const editing = ref(false);
    const loading = ref(false);

    const doc = ref<LawDocument>();
    const logs = shallowRef<LogRow[]>([]);
    const userHasRights = computed(() => appInfo.isAdmin || appInfo.isResponsible);
    const fetchDocument = async () => {
        try {
            [
                doc.value,
                logs.value,
            ] = await Promise.all([
                GetDocument(documentId),
                GetLogs(documentId),
            ]);
        } catch (error) {
            if (isAxiosError<{ error?: string }>(error)) {
                if (error.response?.status === 404) {
                    await router.push("/");
                } else {
                    alert(error.response?.data.error);
                }
            } else {
                alert((error as Error).toString());
            }
        }
    };

    const onChangeUpdate = async () => {
        if (doc.value) {
            try {
                doc.value = await UpdateDocument(doc.value);
                ElNotification({
                    title: "Сохранено",
                    type: "success",
                    showClose: false,
                    duration: 750,
                });
                logs.value = await GetLogs(documentId);
            } catch (error) {
                if (error instanceof Error) {
                    console.error(error);
                    alert(error);
                }
            }
        }
    };
    const fetchUpdate = async () => {
        if (!doc.value) return;
        loading.value = true;
        try {
            doc.value = await FetchUpdate(doc.value);
            ElNotification({
                title: "",
                message: "Документ обновлен",
                type: "success",
                duration: 1000,
                showClose: false,
            });
        } catch (error) {
            if (error instanceof Error) {
                console.error(error);
                alert(error);
            }
        } finally {
            loading.value = false;
        };
    };
    const deleteDocument = async (id: string) => {
        try {
            const result = await ElMessageBox.confirm("Вы уверены что хотите удалить документ?", {
                confirmButtonText: "Да",
                cancelButtonText: "Нет",
                confirmButtonClass: "buttton-danger",
                type: "error",
            });
            if (result === "confirm") {
                await DeleteDocument(id);
                await router.push("/");
            }
        } catch (error) {
            if (error instanceof Error) {
                console.error(error);
            }
        }
    };

    const setSource = async (payload: { id: string; source: string }) => {
        loading.value = true;
        return PatchDraftDocument(payload)
            .then(response =>
                router.replace({ name: "DocumentCard", query: { id: response.data } }).then(fetchDocument))
            .catch((error: unknown) => {
                if (isAxiosError<{ error?: string }>(error)) {
                    ElMessageBox.alert(error.response?.data.error, "Ошибка").catch(console.warn);
                }
            })
            .finally(() => loading.value = false);
    };

    return {
        documentFields,
        doc, logs,
        editing,
        loading,

        userHasRights,
        fetchDocument,
        fetchUpdate,
        onChangeUpdate,
        deleteDocument,
        setSource,
    };
};
