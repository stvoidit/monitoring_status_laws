<template>
    <MainPageHeader
        v-model="fields"
        @download-p-d-f="downloadPDF"
        @download-x-l-s-x="downloadDocuments"
        @add-document="fetchDocuments"
        @update-all="fetchUpdateAll" />
    <el-container direction="horizontal">
        <el-aside width="250px">
            <FilterTypes
                v-model="filters.types"
                :statuses-count="statusesCount" />
        </el-aside>
        <el-main>
            <TableFilters
                v-model="filters"
                :scope-options="scopeOptions"
                @reset-filters="resetFilters" />
            <el-row style="padding-bottom: 10px;">
                <el-col :span="24">
                    <el-pagination
                        :key="filters.currentPage"
                        v-model:current-page="filters.currentPage"
                        v-model:page-size="filters.pageSize"
                        :page-sizes="[10, 50, 100]"
                        style="margin-top:0.5em"
                        background
                        layout="prev, pager, next, sizes"
                        :total="paginationSize" />
                </el-col>
            </el-row>
            <el-row>
                <TableDocuments />
            </el-row>
        </el-main>
    </el-container>
</template>

<script setup lang="ts">
import { onMounted, onBeforeMount } from "vue";
import useStore from "@/store";
import { storeToRefs } from "pinia";
import MainPageHeader from "./components/MainPageHeader.vue";
import TableFilters from "./components/TableFilters.vue";
import FilterTypes from "./components/FilterTypes.vue";
import TableDocuments from "./components/TableDocuments.vue";

const store = useStore();
const {
    resetFilters,
    fetchDocuments,
    fetchUpdateAll,
    downloadPDF,
    downloadDocuments,
    loadFieldsSettings,
} = store;
const {
    paginationSize,
    statusesCount,
    scopeOptions,
    filters,
    fields,
} = storeToRefs(store);

onMounted(fetchDocuments);
onBeforeMount(loadFieldsSettings);

</script>

<style lang="css">
th > div.cell {
    word-break: inherit !important;
    text-overflow: initial !important;
}
td > div.cell {
    word-break: break-word !important;
    text-overflow: initial !important;
}
.el-header {
    padding:0px;
}

.filter-col > div {
    padding-right: 10px;
}
</style>
