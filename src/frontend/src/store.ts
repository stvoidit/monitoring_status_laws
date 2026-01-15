import { LawDocument } from "@/api/models";
import { computed, reactive, ref, toValue, watch } from "vue";
import { GetDocuments, FetchUpdateAll, DownloadDocuments } from "@/api";
import { MessageAlert, MessageSuccess } from "@/composibles/useAlert";
import { defineStore } from "pinia";
import useFilters from "./composibles/useFilters";
import { fnMapData } from "@/utils/makePDF";

const useStore = defineStore("main", () => {
    const loading = ref(false);
    const toggleLoading = () => loading.value = !loading.value;

    const {
        filters,
        sliceStartEnd,
        resetFilters,
        filterRow,
        filtersPipeline,
        indexMethod,
    } = useFilters();

    const tableData = ref<LawDocument[]>([]);
    /** опции для фильтров, scopes генерируются пр получении данных */
    const scopeOptions = reactive({
        sources: [
            "regulation.gov.ru",
            "sozd.duma.gov.ru",
        ],
        scopes: [] as string[],
        lawtypes: [
            "Закон",
            "Законопроект",
        ],
    });

    const uniqueScopes = (docs: LawDocument[]) => {
        scopeOptions.scopes = Array.from(new Set(docs.map(d => d.scope))).filter(s => s !== "").sort();
    };
    const fetchDocuments = async () => {
        /** генерация уникальных значений для фильтра */
        toggleLoading();
        try {
            tableData.value = await GetDocuments();
            if (!scopeOptions.scopes.length) {
                uniqueScopes(tableData.value);
            }
        } catch (error) {
            MessageAlert((error as Error).message);
        } finally {
            toggleLoading();
        }
    };
    const fetchUpdateAll = async () => {
        toggleLoading();
        try {
            const response = await FetchUpdateAll();
            MessageSuccess(response.message);
        } catch (error) {
            MessageAlert((error as Error).message);
        } finally {
            toggleLoading();
        }
    };

    const sortOptions = reactive<{ prop: string | null; dict: number }>({
        prop: null,
        dict: 1,
    });
    const fields = ref(LawDocument.TableFieldsList());
    const loadFieldsSettings = () => {
        const settings = window.localStorage.getItem("fieldsSettings");
        if (!settings) return;
        const loadedFields: typeof fields.value = JSON.parse(settings);
        if (!Array.isArray(loadedFields) || !loadedFields.length) return;
        if (loadedFields.length !== fields.value.length) return;
        const originalPropsCount = fields.value.reduce((sum, f) => sum += Object.keys(f).length, 0);
        const loadedPropsCount = loadedFields.reduce((sum, f) => sum += Object.keys(f as object).length, 0);
        if (originalPropsCount !== loadedPropsCount) {
            return;
        }
        for (const ofield of fields.value) {
            for (const lfield of loadedFields) {
                if (lfield.prop !== ofield.prop) continue;
                if (!lfield.minWidth || lfield.minWidth < ofield.minWidth) {
                    lfield.minWidth = ofield.minWidth;
                }
                if (!lfield.align || lfield.align !== ofield.align) {
                    lfield.align = ofield.align;
                }
            }
        }
        fields.value = loadedFields;
    };

    const tableFields = computed(() => fields.value.filter(f => f.show));

    /** подсчет кол-ва документов для бокового меню */
    const statusesCount = computed(() => ({
        total: tableData.value.reduce((sum, d) => filterRow(d) ? sum += 1 : sum, 0),
        completed: tableData.value.reduce((sum, d) => filterRow(d) && !d.archive && d.percent === "100%" && !d.is_cancelled ? sum += 1 : sum, 0),
        oncontrol: tableData.value.reduce((sum, d) => filterRow(d) && !d.archive && d.percent !== "100%" && !d.is_cancelled ? sum += 1 : sum, 0),
        cancelled: tableData.value.reduce((sum, d) => filterRow(d) && !d.archive && d.is_cancelled ? sum += 1 : sum, 0),
        Important: tableData.value.reduce((sum, d) => filterRow(d) && d.priority <= 2 ? sum += 1 : sum, 0),
        favorites: tableData.value.reduce((sum, d) => filterRow(d) && d.is_favorite ? sum += 1 : sum, 0),
        draft: tableData.value.reduce((sum, d) => filterRow(d) && d.is_draft ? sum += 1 : sum, 0),
        archive: tableData.value.reduce((sum, d) => filterRow(d) && d.archive ? sum += 1 : sum, 0),
    }));

    const computedTableData = computed(() => {
        return tableData.value.filter(filtersPipeline);
    });
    /** список документов после фильтрации с учетом сортировки и фильтров */
    const computedPaginationTableData = computed(() => {
        const copyArray = toValue(computedTableData);
        const sortProp = sortOptions.prop;
        if (sortProp) {
            switch (sortProp) {
                case "Date":
                    copyArray.sort((a, b) => a.date < b.date ? 1 * sortOptions.dict : -1 * sortOptions.dict);
                    break;
                case "changed":
                    copyArray.sort((a, b) => {
                        if (!a.LastChangeDate || !b.LastChangeDate) return 1;
                        return a.LastChangeDate < b.LastChangeDate ? 1 * sortOptions.dict : -1 * sortOptions.dict;
                    });
                    break;
                case "percent":
                    copyArray.sort((a, b) => {
                        const aperc = Number(a.percent.replace("%", ""));
                        const bperc = Number(b.percent.replace("%", ""));
                        return aperc < bperc ? 1 * sortOptions.dict : -1 * sortOptions.dict;
                    });
                    break;
                default:
                    copyArray.sort((a, b) => {
                        if (!a[sortProp] || !b[sortProp]) return 1;
                        return a[sortProp].toString().toLowerCase() < b[sortProp].toString().toLowerCase() ? 1 * sortOptions.dict : -1 * sortOptions.dict;
                    });
                    break;
            }
        }
        return copyArray.slice(sliceStartEnd.value.start, sliceStartEnd.value.end);
    });

    function buildOutData() {
        const visibleFields = fields.value.filter(f => f.show);
        const visibleHeaders = visibleFields.map(f => f.label);
        const visibleProps = visibleFields.map(f => f.prop);
        visibleHeaders.unshift("#");
        visibleProps.unshift("#");
        // const fnMapData = (d: LawDocument, index: number) => d.ArrayValues(index + 1, visibleProps);
        const contentData = computedTableData.value.map(fnMapData(visibleProps));
        const buildData = {
            headers: visibleHeaders,
            data: contentData,
        };
        return buildData;
    };
    /** скачать список документов в excel */
    async function downloadDocuments() {
        const buildData = buildOutData();
        await DownloadDocuments(buildData);
    };
    const scrollTable = reactive({
        scrollLeft: 0,
        scrollTop: 0,
    });

    watch(() => filters.currentPage, (newCurrentPage, oldCurrentPage) => {
        /**
         * манипуляция скролом при пагинации, т.к. мы полностью перехватываем ивент,
         * то необходимо реализовать возврат скролла к началу так же вручную
         */
        if (newCurrentPage !== oldCurrentPage) {
            scrollTable.scrollTop = 0;
        }
    }, {
        immediate: true,
        flush: "sync",
    });

    const paginationSize = computed(() => computedTableData.value.length);

    async function downloadPDF() {
        const { downloadPDF } = await import("@/utils/makePDF");
        await downloadPDF(toValue(computedTableData.value));
    };

    return {
        fields,
        filters,
        resetFilters,
        tableData,
        tableFields,
        sortOptions,
        statusesCount,
        computedTableData,
        computedPaginationTableData,
        sliceStartEnd,
        downloadDocuments,
        scrollTable,
        fetchDocuments,
        fetchUpdateAll,
        paginationSize,
        loading,
        scopeOptions,
        downloadPDF,
        loadFieldsSettings,
        indexMethod,
    };
});
export default useStore;
