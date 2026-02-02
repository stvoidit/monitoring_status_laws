<template>
    <el-table
        ref="tableDocs"
        v-loading="loading"
        class-name="normalize-table"
        max-height="65vh"
        border
        fit
        table-layout="fixed"
        show-overflow-tooltip
        :tooltip-options="tooltipOptions"
        size="small"
        :data="computedPaginationTableData"
        style="width: 100%"
        @scroll.capture="onScroll"
        @header-dragend="resizeColumn"
        @row-click="rowClick"
        @sort-change="sortingTable">
        <el-table-column
            type="index"
            align="center"
            header-align="left"
            :index="store.indexMethod"
            label="#"
            width="50" />
        <el-table-column
            v-for="field in tableFields"
            :key="field.prop"
            v-bind="field"
            width="auto"
            header-align="left"
            resizable
            :formatter="formatterFn" />
    </el-table>
</template>

<script setup lang="ts">
import useStore from "@/store";
import { storeToRefs } from "pinia";
import type { LawDocument } from "@/api/models";
import { useTemplateRef, onUpdated, h } from "vue";
import { useRouter } from "vue-router";
import { type UseTooltipProps } from "element-plus";
import FavoriteButton from "@/components/FavoriteButton.vue";

const formatterFn = (row: LawDocument, col: { property: string }, cellValue) => {
    if (col.property !== "is_favorite") return String(cellValue);
    return h(FavoriteButton, {
        text: true,
        modelValue: row.is_favorite,
        projectId: row.id,
        ["onUpdate:modelValue"](value) {
            row.is_favorite = value;
        },
    });
};

const router = useRouter();
const store = useStore();
const {
    loading,
    fields,
    tableFields,
    computedPaginationTableData,
    sortOptions,
    scrollTable,
} = storeToRefs(store);

const sortingTable = ({ prop, order }) => {
    sortOptions.value.dict = order == "ascending" ? 1 : -1;
    sortOptions.value.prop = prop;
};

const rowClick = async ({ id }: LawDocument, column: { property: string }) => {
    if (column.property === "is_favorite") return;
    await router.push({ path: "/document", query: { id } });
};

const resizeColumn = (newWidth: number, _: number, column: { property: string }) => {
    const fieldIndex = fields.value.findIndex(f => f.prop === column.property);
    fields.value[fieldIndex].minWidth = newWidth;
    window.localStorage.setItem("fieldsSettings", JSON.stringify(fields.value));
};

const refTableDocs = useTemplateRef("tableDocs");
onUpdated(() => {
    refTableDocs.value?.setScrollTop(scrollTable.value.scrollTop);
});
function onScroll(event: Event) {
    const { scrollLeft, scrollTop } = event.target as HTMLBodyElement;
    [
        scrollTable.value.scrollLeft,
        scrollTable.value.scrollTop,
    ] = [
        scrollLeft,
        scrollTop,
    ];
}

const tooltipOptions: Partial<UseTooltipProps> = {
    strategy: "absolute",
    teleported: true,
    transition: "el-fade-in-linear",
    popperStyle: {
        width: "fit-content",
        maxWidth: "20vw",
    },
    gpuAcceleration: true,
};

</script>
