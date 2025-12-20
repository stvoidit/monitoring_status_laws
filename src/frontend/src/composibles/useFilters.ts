import { type IFilters, LawDocument } from "@/api/models";
import { computed, reactive, toValue } from "vue";
import dayjs from "dayjs";

export default () => {
    const filters: IFilters = reactive({
        source: "",
        scope: "",
        dates: undefined,
        types: "total",
        currentPage: 1,
        pageSize: 50,
        search: "",
        lawtype: "",
    });

    /** сброс фильтров */
    const resetFilters = () => {
        filters.source = "";
        filters.scope = "";
        filters.dates = undefined;
        filters.search = "";
        filters.lawtype = "";
        filters.currentPage = 1;
    };
    /** расчет нарезки массива документов */
    const sliceStartEnd = computed(() => {
        const start = (filters.currentPage - 1) * filters.pageSize;
        const end = filters.currentPage * filters.pageSize;
        return { start, end };
    });
    /** функция расчета индекса для таблицы */
    function indexMethod(index: number) {
        return sliceStartEnd.value.start + 1 + index;
    }

    const cbFilter = {
        total: (_: LawDocument) => true,
        completed: (d: LawDocument) => d.percent === "100%" && !d.archive && !d.is_cancelled,
        oncontrol: (d: LawDocument) => d.percent !== "100%" && !d.archive && !d.is_cancelled,
        cancelled: (d: LawDocument) => !d.archive && d.is_cancelled,
        Important: (d: LawDocument) => d.priority <= 2,
        favorites: (d: LawDocument) => d.is_favorite,
        draft: (d: LawDocument) => d.is_draft,
        archive: (d: LawDocument) => d.archive,
    } as const;
    /** фильтрация по типу */
    const filterTypes = (d: LawDocument) => cbFilter[filters.types](d);
    /** фильтрация списка документов */
    const filterText = (s: string, props: string[]) => {
        return (sd: LawDocument) => {
            if (!s) return true;
            let find = false;
            for (const prop of props) {
                const field = sd[prop as keyof LawDocument];
                if (!field) {
                    continue;
                };
                if (typeof field === "object") continue;
                const value: string = field.toString();
                find = value.toLowerCase().includes(s);
                if (find) {
                    break;
                }
            }
            return find;
        };
    };
    /** функция-фильтр для документа */
    function filterRow(d: LawDocument) {
        const filterSource = (d: LawDocument) => filters.source !== "" ? d.source === filters.source : true;
        const filterScope = (d: LawDocument) => filters.scope !== "" ? d.scope === filters.scope : true;
        const filterLawType = (d: LawDocument) => filters.lawtype !== "" ? d.LawType === filters.lawtype : true;
        const filterDates = (d: LawDocument) => {
            if (!filters.dates) return true;
            if (!d.LastChangeDate) return false;
            else {
                const numberTime = d.LastChangeDate.getTime();
                const startDay = dayjs(filters.dates[0]).startOf("day").toDate();
                const endDay = dayjs(filters.dates[1]).endOf("day").toDate();
                return numberTime >= startDay.getTime() && numberTime <= endDay.getTime();
            }
        };
        return [
            filterSource,
            filterScope,
            filterDates,
            filterLawType,
        ].every(fn => fn(d));
    };

    const searchProps = computed(() => LawDocument.TableFieldsList().map(({ prop }) => prop));
    const searchText = computed(() => filters.search.toLowerCase());

    const filtersPipeline = (d: LawDocument) => {
        const textFilter = filterText(toValue(searchText), toValue(searchProps));
        return filterRow(d) && filterTypes(d) && textFilter(d);
    };

    return {
        filters,
        sliceStartEnd,
        resetFilters,
        filterTypes,
        filterText,
        filterRow,
        filtersPipeline,
        indexMethod,
    };
};
